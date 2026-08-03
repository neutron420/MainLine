variable "name" {
  description = "Resource name prefix (also the ECS cluster name)"
  type        = string
}

variable "environment" {
  description = "Environment label (dev/staging/production) — gates destructive settings"
  type        = string
}

variable "region" {
  description = "AWS region (used for awslogs configuration)"
  type        = string
}

variable "vpc_id" {
  description = "ID of the VPC"
  type        = string
}

variable "public_subnet_ids" {
  description = "Public subnet IDs for the ALB"
  type        = list(string)
}

variable "private_subnet_ids" {
  description = "Private subnet IDs for Fargate tasks"
  type        = list(string)
}

variable "backend_image" {
  description = "Backend container image (ECR/ghcr)"
  type        = string
}

variable "database_url" {
  description = "Postgres connection URL. Sensitive, no default."
  type        = string
  sensitive   = true
}

variable "redis_url" {
  description = "Redis connection URL. Sensitive, no default."
  type        = string
  sensitive   = true
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

variable "cpu" {
  description = "Task CPU units (Fargate)"
  type        = number
  default     = 512
}

variable "memory" {
  description = "Task memory MiB (Fargate)"
  type        = number
  default     = 1024
}

variable "desired_count" {
  description = "Initial desired task count"
  type        = number
  default     = 1
}

variable "min_capacity" {
  description = "Autoscaling minimum task count"
  type        = number
  default     = 1
}

variable "max_capacity" {
  description = "Autoscaling maximum task count"
  type        = number
  default     = 4
}

variable "log_level" {
  description = "Backend LOG_LEVEL"
  type        = string
  default     = "info"
}

variable "log_format" {
  description = "Backend LOG_FORMAT (json/text)"
  type        = string
  default     = "json"
}

variable "log_retention_days" {
  description = "CloudWatch log group retention"
  type        = number
  default     = 14
}

variable "frontend_url" {
  description = "Frontend origin (CORS allowlist)"
  type        = string
  default     = "https://app.schemahub.dev"
}

variable "db_pool_min" {
  description = "Backend DB_POOL_MIN"
  type        = number
  default     = 2
}

variable "db_pool_max" {
  description = "Backend DB_POOL_MAX"
  type        = number
  default     = 20
}

variable "certificate_arn" {
  description = "ACM certificate ARN for the HTTPS listener; null = HTTP-only (terminate TLS at CloudFront/Envoy instead)"
  type        = string
  default     = null
}

variable "tags" {
  description = "Additional tags applied to all resources"
  type        = map(string)
  default     = {}
}
