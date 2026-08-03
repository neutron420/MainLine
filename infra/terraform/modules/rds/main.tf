# RDS PostgreSQL module — managed Postgres 16 for the SchemaHub control plane.
#
# Design decisions:
#   - Single-instance RDS in the private subnets (Neon remains the supported
#     primary database deployment for multi-tenant target databases; this RDS
#     instance backs the SchemaHub application itself on AWS).
#   - Storage is always encrypted (default KMS key unless `kms_key_id` is set).
#   - Automated backups enabled with a 7-day retention (RPO target: 5 minutes
#     with PITR, see docs/BACKUP_DR.md).
#   - Deletion protection is on; a final snapshot is taken on destroy.
#   - Security group allows port 5432 ONLY from the application security
#     group id passed in `allowed_security_group_id` — never from 0.0.0.0/0.
#   - The connection URL is stored as a SecureString SSM parameter and is
#     also returned as a sensitive output for the ECS module.

resource "aws_db_subnet_group" "this" {
  name       = "${var.name}-db-subnet"
  subnet_ids = var.subnet_ids

  tags = merge(var.tags, {
    Name = "${var.name}-db-subnet"
  })
}

resource "aws_security_group" "this" {
  name        = "${var.name}-rds"
  description = "PostgreSQL access: 5432 from the application SG only"
  vpc_id      = var.vpc_id

  ingress {
    description     = "PostgreSQL from the application tier"
    from_port       = 5432
    to_port         = 5432
    protocol        = "tcp"
    security_groups = [var.allowed_security_group_id]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = merge(var.tags, {
    Name = "${var.name}-rds-sg"
  })
}

resource "aws_db_parameter_group" "this" {
  name   = "${var.name}-pg"
  family = "postgres16"

  parameter {
    name  = "log_statement"
    value = "ddl"
  }

  parameter {
    name  = "log_min_duration_statement"
    value = "1000"
  }

  parameter {
    name  = "idle_in_transaction_session_timeout"
    value = "60000"
  }

  tags = merge(var.tags, {
    Name = "${var.name}-pg"
  })
}

resource "aws_db_instance" "this" {
  identifier                      = var.name
  engine                          = "postgres"
  engine_version                  = var.engine_version
  instance_class                  = var.instance_class
  allocated_storage               = var.allocated_storage
  storage_type                    = "gp3"
  storage_encrypted               = true
  kms_key_id                      = var.kms_key_id
  db_name                         = var.db_name
  username                        = var.db_username
  password                        = var.db_password
  db_subnet_group_name            = aws_db_subnet_group.this.name
  vpc_security_group_ids          = [aws_security_group.this.id]
  parameter_group_name            = aws_db_parameter_group.this.name
  multi_az                        = var.multi_az
  backup_retention_period         = var.backup_retention_period
  backup_window                   = var.backup_window
  maintenance_window              = var.maintenance_window
  skip_final_snapshot             = false
  final_snapshot_identifier       = "${var.name}-final"
  deletion_protection             = true
  enabled_cloudwatch_logs_exports = ["postgresql"]
  monitoring_interval             = 60
  monitoring_role_arn             = var.monitoring_role_arn
  auto_minor_version_upgrade      = true
  performance_insights_enabled    = true

  tags = merge(var.tags, {
    Name = "${var.name}-rds"
  })
}

# Connection URL for the application, kept out of code and out of the plan
# logs — SecureString so the value is not visible in plaintext.
resource "aws_ssm_parameter" "connection_url" {
  name  = "/${var.name}/database/connection-url"
  type  = "SecureString"
  value = "postgres://${aws_db_instance.this.username}:${urlencode(aws_db_instance.this.password)}@${aws_db_instance.this.address}:${aws_db_instance.this.port}/${aws_db_instance.this.db_name}?sslmode=require"

  tags = merge(var.tags, {
    Name = "${var.name}-db-connection-url"
  })
}
