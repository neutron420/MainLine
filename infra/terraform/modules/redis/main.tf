# ElastiCache Redis module — managed Redis for caching, pub/sub (event
# streaming) and rate limiting.
#
# Design decisions:
#   - Cluster mode DISABLED (single shard): SchemaHub's workload is a small
#     cache + pub/sub bus; a single primary with optional replicas is the
#     right size. `num_cache_clusters` controls the total node count
#     (1 = single node; 2+ = primary + replicas).
#   - At-rest encryption always on. Transit encryption is optional; when it
#     is enabled an auth token MUST be supplied (ElastiCache requires it).
#     The backend redis client already supports `rediss://` + auth.
#   - Backups enabled (snapshot retention matches the RDS 7-day window).
#   - Security group allows 6379 ONLY from the application security group.
#   - The connection URL is stored as a SecureString SSM parameter.

locals {
  scheme   = var.transit_encryption_enabled ? "rediss" : "redis"
  endpoint = aws_elasticache_replication_group.this.primary_endpoint_address
  # rediss://:password@host:port mirrors what internal/pkg/redis parses
  # (url.User with no username).
  connection_url = (
    var.transit_encryption_enabled
    ? "${local.scheme}://:${var.auth_token}@${local.endpoint}:6379"
    : "${local.scheme}://${local.endpoint}:6379"
  )
}

resource "aws_elasticache_subnet_group" "this" {
  name       = "${var.name}-redis-subnet"
  subnet_ids = var.subnet_ids

  tags = merge(var.tags, {
    Name = "${var.name}-redis-subnet"
  })
}

resource "aws_security_group" "this" {
  name        = "${var.name}-redis"
  description = "Redis access: 6379 from the application SG only"
  vpc_id      = var.vpc_id

  ingress {
    description     = "Redis from the application tier"
    from_port       = 6379
    to_port         = 6379
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
    Name = "${var.name}-redis-sg"
  })
}

resource "aws_elasticache_parameter_group" "this" {
  name   = "${var.name}-redis-pg"
  family = "redis7"

  # Mirrors docker/redis/redis.conf so local and prod behaviour match:
  # LRU eviction when maxmemory is hit (caches must never block writes).
  parameter {
    name  = "maxmemory-policy"
    value = "allkeys-lru"
  }

  tags = merge(var.tags, {
    Name = "${var.name}-redis-pg"
  })
}

resource "aws_elasticache_replication_group" "this" {
  replication_group_id       = "${var.name}-redis"
  description                = "SchemaHub Redis (cluster mode disabled)"
  engine                     = "redis"
  engine_version             = var.engine_version
  node_type                  = var.node_type
  num_cache_clusters         = var.num_cache_clusters
  parameter_group_name       = aws_elasticache_parameter_group.this.name
  subnet_group_name          = aws_elasticache_subnet_group.this.name
  security_group_ids         = [aws_security_group.this.id]
  port                       = 6379
  at_rest_encryption_enabled = true
  transit_encryption_enabled = var.transit_encryption_enabled
  auth_token                 = var.auth_token
  automatic_failover_enabled = var.num_cache_clusters > 1 ? true : false
  snapshot_retention_limit   = var.snapshot_retention_limit
  snapshot_window            = var.snapshot_window
  maintenance_window         = var.maintenance_window
  auto_minor_version_upgrade = true
  apply_immediately          = false

  tags = merge(var.tags, {
    Name = "${var.name}-redis"
  })
}

resource "aws_ssm_parameter" "connection_url" {
  name  = "/${var.name}/redis/connection-url"
  type  = "SecureString"
  value = local.connection_url

  tags = merge(var.tags, {
    Name = "${var.name}-redis-connection-url"
  })
}
