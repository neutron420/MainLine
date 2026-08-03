# ECS Fargate module — backend service + ALB.
#
# Runs the SchemaHub backend on Fargate in the private subnets behind an
# internet-facing ALB. The ALB listener terminates browser HTTP/1.1 and
# forwards to the container over HTTP/2 (h2c) on 50051 — the protocol the
# backend's plaintext gRPC server speaks, and the same path the Envoy
# gRPC-Web bridge uses in the compose topology.
#
# Health checks: ALB checks GET /metrics on container port 9091 (the
# Prometheus endpoint, which returns 200) — this avoids depending on gRPC
# health service support in the ALB health check.
#
# Secrets: no secret value is placed in the task definition. All sensitive
# configuration is written to SecretsManager (created by this module) and
# referenced via `valueFrom` in the task's `secrets` block. Non-sensitive
# configuration is injected as plain environment variables.
#
# The frontend is intentionally NOT deployed here (keep the module focused):
# it is static hosting on Vercel per the deployment docs. Revisit if a
# containerized frontend becomes the deployment target.

locals {
  # Secrets referenced by the task definition (valueFrom).
  secret_arns = [
    aws_secretsmanager_secret.database_url.arn,
    aws_secretsmanager_secret.redis_url.arn,
    aws_secretsmanager_secret.jwt_private_key.arn,
    aws_secretsmanager_secret.jwt_public_key.arn,
    aws_secretsmanager_secret.encryption_key.arn,
  ]

  container_definitions = [
    {
      name      = "backend"
      image     = var.backend_image
      essential = true

      # Fargate enforces that container cpu/memory match the task values.
      cpu    = var.cpu
      memory = var.memory

      port_mappings = [
        { container_port = 50051, protocol = "tcp" }, # gRPC (h2c)
        { container_port = 9091, protocol = "tcp" },  # Prometheus /metrics
      ]

      environment = [
        { name = "PORT", value = "50051" },
        { name = "METRICS_PORT", value = "9091" },
        { name = "LOG_LEVEL", value = var.log_level },
        { name = "LOG_FORMAT", value = var.log_format },
        { name = "FRONTEND_URL", value = var.frontend_url },
        { name = "DB_POOL_MIN", value = tostring(var.db_pool_min) },
        { name = "DB_POOL_MAX", value = tostring(var.db_pool_max) },
      ]

      secrets = [
        { name = "DATABASE_URL", value_from = aws_secretsmanager_secret.database_url.arn },
        { name = "REDIS_URL", value_from = aws_secretsmanager_secret.redis_url.arn },
        { name = "JWT_PRIVATE_KEY", value_from = aws_secretsmanager_secret.jwt_private_key.arn },
        { name = "JWT_PUBLIC_KEY", value_from = aws_secretsmanager_secret.jwt_public_key.arn },
        { name = "ENCRYPTION_MASTER_KEY", value_from = aws_secretsmanager_secret.encryption_key.arn },
      ]

      log_configuration = {
        log_driver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.this.name
          "awslogs-region"        = var.region
          "awslogs-stream-prefix" = "backend"
        }
      }
    },
  ]
}

# ---------------------------------------------------------------------------
# Cluster
# ---------------------------------------------------------------------------

resource "aws_ecs_cluster" "this" {
  name = var.name

  setting {
    name  = "containerInsights"
    value = "enabled"
  }

  tags = merge(var.tags, {
    Name = "${var.name}-cluster"
  })
}

# ---------------------------------------------------------------------------
# Secrets (values are supplied as variables — never in code)
# ---------------------------------------------------------------------------

resource "aws_secretsmanager_secret" "database_url" {
  name_prefix = "${var.name}/backend/database-url"
  description = "SchemaHub backend DATABASE_URL"
  tags        = var.tags
}

resource "aws_secretsmanager_secret_version" "database_url" {
  secret_id     = aws_secretsmanager_secret.database_url.id
  secret_string = var.database_url
}

resource "aws_secretsmanager_secret" "redis_url" {
  name_prefix = "${var.name}/backend/redis-url"
  description = "SchemaHub backend REDIS_URL"
  tags        = var.tags
}

resource "aws_secretsmanager_secret_version" "redis_url" {
  secret_id     = aws_secretsmanager_secret.redis_url.id
  secret_string = var.redis_url
}

resource "aws_secretsmanager_secret" "jwt_private_key" {
  name_prefix = "${var.name}/backend/jwt-private-key"
  description = "SchemaHub backend JWT_PRIVATE_KEY (RS256 PEM)"
  tags        = var.tags
}

resource "aws_secretsmanager_secret_version" "jwt_private_key" {
  secret_id     = aws_secretsmanager_secret.jwt_private_key.id
  secret_string = var.jwt_private_key
}

resource "aws_secretsmanager_secret" "jwt_public_key" {
  name_prefix = "${var.name}/backend/jwt-public-key"
  description = "SchemaHub backend JWT_PUBLIC_KEY (RS256 PEM)"
  tags        = var.tags
}

resource "aws_secretsmanager_secret_version" "jwt_public_key" {
  secret_id     = aws_secretsmanager_secret.jwt_public_key.id
  secret_string = var.jwt_public_key
}

resource "aws_secretsmanager_secret" "encryption_key" {
  name_prefix = "${var.name}/backend/encryption-key"
  description = "SchemaHub backend ENCRYPTION_MASTER_KEY (>= 32 bytes)"
  tags        = var.tags
}

resource "aws_secretsmanager_secret_version" "encryption_key" {
  secret_id     = aws_secretsmanager_secret.encryption_key.id
  secret_string = var.encryption_key
}

# ---------------------------------------------------------------------------
# IAM (task execution: ECR pull, CloudWatch logs, SecretsManager read)
# ---------------------------------------------------------------------------

resource "aws_iam_role" "task_execution" {
  name = "${var.name}-exec-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect    = "Allow"
        Principal = { Service = "ecs-tasks.amazonaws.com" }
        Action    = "sts:AssumeRole"
      },
    ]
  })

  tags = var.tags
}

resource "aws_iam_role_policy_attachment" "task_execution" {
  role       = aws_iam_role.task_execution.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

resource "aws_iam_role_policy" "task_execution_secrets" {
  name = "secretsmanager-read"
  role = aws_iam_role.task_execution.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = ["secretsmanager:GetSecretValue"]
        Resource = local.secret_arns
      },
      {
        # kms:Decrypt for AWS-managed keys used by SecretsManager.
        Effect   = "Allow"
        Action   = ["kms:Decrypt"]
        Resource = "*"
      },
    ]
  })
}

# ---------------------------------------------------------------------------
# Logging
# ---------------------------------------------------------------------------

resource "aws_cloudwatch_log_group" "this" {
  name              = "/ecs/${var.name}/backend"
  retention_in_days = var.log_retention_days

  tags = var.tags
}

# ---------------------------------------------------------------------------
# Networking: application SG + ALB + ALB SG
# ---------------------------------------------------------------------------
#
# The two SGs reference each other (app allows ALB, ALB egresses to app),
# which would be a cycle if expressed with inline rules. AWS SGs have no
# ordering requirement, so the cross-references are expressed as
# aws_security_group_rule resources — the SGs themselves stay rule-free and
# the rules attach after both exist.

resource "aws_security_group" "app" {
  name        = "${var.name}-app"
  description = "Backend containers: gRPC + metrics from the ALB only"
  vpc_id      = var.vpc_id

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = merge(var.tags, {
    Name = "${var.name}-app-sg"
  })
}

resource "aws_security_group_rule" "app_ingress_grpc" {
  type                     = "ingress"
  security_group_id        = aws_security_group.app.id
  source_security_group_id = aws_security_group.alb.id
  from_port                = 50051
  to_port                  = 50051
  protocol                 = "tcp"
  description              = "gRPC (h2c) from the ALB"
}

resource "aws_security_group_rule" "app_ingress_metrics" {
  type                     = "ingress"
  security_group_id        = aws_security_group.app.id
  source_security_group_id = aws_security_group.alb.id
  from_port                = 9091
  to_port                  = 9091
  protocol                 = "tcp"
  description              = "Prometheus /metrics from the ALB (health check)"
}

resource "aws_security_group" "alb" {
  name        = "${var.name}-alb"
  description = "ALB: HTTP(S) from the internet, outbound to the app SG"
  vpc_id      = var.vpc_id

  ingress {
    description = "HTTP"
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  dynamic "ingress" {
    # TLS listener exists only when a certificate ARN is provided.
    for_each = var.certificate_arn == null ? [] : [1]
    content {
      description = "HTTPS"
      from_port   = 443
      to_port     = 443
      protocol    = "tcp"
      cidr_blocks = ["0.0.0.0/0"]
    }
  }

  tags = merge(var.tags, {
    Name = "${var.name}-alb-sg"
  })
}

resource "aws_security_group_rule" "alb_egress_app" {
  type                     = "egress"
  security_group_id        = aws_security_group.alb.id
  source_security_group_id = aws_security_group.app.id
  from_port                = 0
  to_port                  = 0
  protocol                 = "-1"
  description              = "To backend gRPC + metrics ports"
}

resource "aws_lb" "this" {
  name               = substr("${var.name}-alb", 0, 32)
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.alb.id]
  subnets            = var.public_subnet_ids

  enable_deletion_protection = var.environment == "production" ? true : false
  enable_http2               = true
  idle_timeout               = 60
  drop_invalid_header_fields = true

  tags = merge(var.tags, {
    Name = "${var.name}-alb"
  })
}

# Target group: HTTP2 (h2c) to the container's plaintext gRPC port.
resource "aws_lb_target_group" "backend" {
  name             = substr("${var.name}-tg", 0, 32)
  port             = 50051
  protocol         = "HTTP"
  protocol_version = "HTTP2"
  target_type      = "ip"
  vpc_id           = var.vpc_id

  # Health check against the /metrics endpoint on 9091 (always 200 when the
  # process is alive). Unhealthy nodes are drained by the service.
  health_check {
    enabled             = true
    interval            = 30
    timeout             = 5
    healthy_threshold   = 3
    unhealthy_threshold = 3
    port                = 9091
    path                = "/metrics"
    protocol            = "HTTP"
    matcher             = "200-399"
  }

  tags = merge(var.tags, {
    Name = "${var.name}-tg"
  })
}

resource "aws_lb_listener" "http" {
  load_balancer_arn = aws_lb.this.arn
  port              = 80
  protocol          = "HTTP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.backend.arn
  }
}

resource "aws_lb_listener" "https" {
  count = var.certificate_arn == null ? 0 : 1

  load_balancer_arn = aws_lb.this.arn
  port              = 443
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-TLS13-1-2-2021-06"
  certificate_arn   = var.certificate_arn

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.backend.arn
  }
}

# ---------------------------------------------------------------------------
# Task definition + service
# ---------------------------------------------------------------------------

resource "aws_ecs_task_definition" "backend" {
  family                   = "${var.name}-backend"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = var.cpu
  memory                   = var.memory
  execution_role_arn       = aws_iam_role.task_execution.arn
  # Task role reuses the execution role: the backend does not call AWS APIs
  # at runtime today. Introduce a dedicated minimal task role the day it does.
  task_role_arn         = aws_iam_role.task_execution.arn
  container_definitions = jsonencode(local.container_definitions)

  tags = var.tags
}

resource "aws_ecs_service" "backend" {
  name             = "backend"
  cluster          = aws_ecs_cluster.this.id
  task_definition  = aws_ecs_task_definition.backend.arn
  desired_count    = var.desired_count
  launch_type      = "FARGATE"
  platform_version = "LATEST"

  network_configuration {
    subnets          = var.private_subnet_ids
    security_groups  = [aws_security_group.app.id]
    assign_public_ip = false
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.backend.arn
    container_name   = "backend"
    container_port   = 50051
  }

  # Grace period absorbs the container's boot + DB connection before the
  # ALB starts failing health checks.
  health_check_grace_period_seconds = 60

  deployment_minimum_healthy_percent = 50
  deployment_maximum_percent         = 200

  depends_on = [
    aws_lb_listener.http,
    aws_iam_role_policy.task_execution_secrets,
  ]

  tags = var.tags
}

# ---------------------------------------------------------------------------
# Autoscaling: CPU target tracking between min/max instances
# ---------------------------------------------------------------------------

resource "aws_appautoscaling_target" "backend" {
  service_namespace  = "ecs"
  resource_id        = "service/${aws_ecs_cluster.this.name}/${aws_ecs_service.backend.name}"
  scalable_dimension = "ecs:service:DesiredCount"
  min_capacity       = var.min_capacity
  max_capacity       = var.max_capacity
}

resource "aws_appautoscaling_policy" "backend_cpu" {
  name               = "${var.name}-cpu-target"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.backend.resource_id
  scalable_dimension = aws_appautoscaling_target.backend.scalable_dimension
  service_namespace  = aws_appautoscaling_target.backend.service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageCPUUtilization"
    }
    target_value       = 70.0
    scale_in_cooldown  = 300
    scale_out_cooldown = 60
  }
}
