variable "name" {
  description = "Resource name prefix (keep short; used in resource names with 32-char limits)"
  type        = string
}

variable "environment" {
  description = "Environment label (dev/staging/production)"
  type        = string
}

variable "region" {
  description = "AWS region to deploy into"
  type        = string
  default     = "us-east-1"
}

variable "azs" {
  description = "Availability zones (empty = first 3 healthy AZs of the region)"
  type        = list(string)
  default     = []
}

variable "vpc_cidr" {
  description = "CIDR block for the VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "nat_gateway_count" {
  description = "NAT gateways (1 = cost mode; 3 = HA one-per-AZ)"
  type        = number
  default     = 1
}

# ---------------------------------------------------------------------------
# Database (RDS)
# ---------------------------------------------------------------------------

variable "db_name" {
  description = "Database name"
  type        = string
  default     = "schemahub"
}

variable "db_username" {
  description = "RDS master username"
  type        = string
  default     = "schemahub"
}

variable "db_password" {
  description = "RDS master password. No default — supply via tfvars or a secret manager. Sensitive."
  type        = string
  sensitive   = true
}

variable "rds_engine_version" {
  description = "PostgreSQL version for RDS (16.x)"
  type        = string
  default     = "16.4"
}

variable "rds_instance_class" {
  description = "RDS instance class"
  type        = string
  default     = "db.t4g.small"
}

variable "rds_allocated_storage" {
  description = "RDS storage in GiB"
  type        = number
  default     = 20
}

variable "rds_multi_az" {
  description = "Multi-AZ RDS (production: true)"
  type        = bool
  default     = false
}

# ---------------------------------------------------------------------------
# Redis (ElastiCache)
# ---------------------------------------------------------------------------

variable "redis_engine_version" {
  description = "Redis engine version"
  type        = string
  default     = "7.1"
}

variable "redis_node_type" {
  description = "ElastiCache node type"
  type        = string
  default     = "cache.t3.micro"
}

variable "redis_num_cache_clusters" {
  description = "Redis nodes (1 = single; 2+ = with replicas and auto-failover)"
  type        = number
  default     = 1
}

variable "redis_transit_encryption_enabled" {
  description = "TLS in transit to Redis (production: true, requires redis_auth_token)"
  type        = bool
  default     = false
}

variable "redis_auth_token" {
  description = "Redis AUTH token; required when transit encryption is on. Sensitive, no default."
  type        = string
  sensitive   = true
  default     = null
}

# ---------------------------------------------------------------------------
# Backend (ECS)
# ---------------------------------------------------------------------------

variable "backend_image" {
  description = "Backend container image (ECR or ghcr)"
  type        = string
  default     = "ghcr.io/schemahub/backend:latest"
}

variable "frontend_url" {
  description = "Frontend origin for CORS allowlisting"
  type        = string
  default     = "http://localhost:3000"
}

variable "jwt_private_key" {
  description = "JWT RS256 private key (PEM). Sensitive, no default."
  type        = string
  sensitive   = true
}

variable "jwt_public_key" {
  description = "JWT RS256 public key (PEM). Sensitive, no default."
  type        = string
  sensitive   = true
}

variable "encryption_key" {
  description = "Application encryption master key (>= 32 bytes). Sensitive, no default."
  type        = string
  sensitive   = true
}

variable "backend_cpu" {
  description = "Fargate CPU units"
  type        = number
  default     = 512
}

variable "backend_memory" {
  description = "Fargate memory MiB"
  type        = number
  default     = 1024
}

variable "backend_desired_count" {
  description = "Initial desired task count"
  type        = number
  default     = 1
}

variable "backend_min_capacity" {
  description = "Autoscaling minimum task count"
  type        = number
  default     = 1
}

variable "backend_max_capacity" {
  description = "Autoscaling maximum task count"
  type        = number
  default     = 4
}

variable "db_pool_min" {
  description = "Backend DB_POOL_MIN"
  type        = number
  default     = 2
}

variable "db_pool_max" {
  description = "Backend DB_POOL_MAX (keep low when the pooler-backed Neon endpoint is used)"
  type        = number
  default     = 20
}

variable "acm_certificate_arn" {
  description = "ACM certificate ARN for HTTPS on the ALB; null = HTTP-only (TLS terminated elsewhere)"
  type        = string
  default     = null
}
