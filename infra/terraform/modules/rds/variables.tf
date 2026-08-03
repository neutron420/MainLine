variable "name" {
  description = "Resource name prefix"
  type        = string
}

variable "vpc_id" {
  description = "ID of the VPC to place the RDS instance in"
  type        = string
}

variable "subnet_ids" {
  description = "Private subnet IDs for the DB subnet group"
  type        = list(string)
}

variable "allowed_security_group_id" {
  description = "Security group allowed to reach port 5432 (the ECS application SG)"
  type        = string
}

variable "db_name" {
  description = "Name of the database created inside the instance"
  type        = string
  default     = "schemahub"
}

variable "db_username" {
  description = "Master username (must not be a reserved word such as 'postgres')"
  type        = string
  default     = "schemahub"
}

variable "db_password" {
  description = "Master password. No default: must be supplied via tfvars / secret manager at apply time."
  type        = string
  sensitive   = true
}

variable "engine_version" {
  description = "PostgreSQL major.minor version (16.x)"
  type        = string
  default     = "16.4"
}

variable "instance_class" {
  description = "RDS instance class"
  type        = string
  default     = "db.t4g.small"
}

variable "allocated_storage" {
  description = "Allocated storage in GiB (gp3)"
  type        = number
  default     = 20
}

variable "multi_az" {
  description = "Enable Multi-AZ (standby replica) for HA"
  type        = bool
  default     = false
}

variable "backup_retention_period" {
  description = "Days of automated backups (PITR window)"
  type        = number
  default     = 7
}

variable "backup_window" {
  description = "Preferred backup window (UTC)"
  type        = string
  default     = "02:00-03:00"
}

variable "maintenance_window" {
  description = "Preferred maintenance window (UTC)"
  type        = string
  default     = "sun:05:00-sun:06:00"
}

variable "kms_key_id" {
  description = "KMS key ARN for storage encryption; null = AWS managed key"
  type        = string
  default     = null
}

variable "monitoring_role_arn" {
  description = "IAM role ARN for enhanced monitoring; null = enhanced monitoring disabled"
  type        = string
  default     = null
}

variable "tags" {
  description = "Additional tags applied to all resources"
  type        = map(string)
  default     = {}
}
