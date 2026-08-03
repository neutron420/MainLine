variable "name" {
  description = "Resource name prefix"
  type        = string
}

variable "vpc_id" {
  description = "ID of the VPC to place the Redis cluster in"
  type        = string
}

variable "subnet_ids" {
  description = "Private subnet IDs for the cache subnet group"
  type        = list(string)
}

variable "allowed_security_group_id" {
  description = "Security group allowed to reach port 6379 (the ECS application SG)"
  type        = string
}

variable "engine_version" {
  description = "Redis engine version"
  type        = string
  default     = "7.1"
}

variable "node_type" {
  description = "ElastiCache node type"
  type        = string
  default     = "cache.t3.micro"
}

variable "num_cache_clusters" {
  description = "Total nodes (1 = single primary; 2+ = primary + replicas with auto-failover)"
  type        = number
  default     = 1
}

variable "transit_encryption_enabled" {
  description = "Enable TLS in transit (forces an auth token, see docs)"
  type        = bool
  default     = false
}

variable "auth_token" {
  description = "Redis AUTH token; REQUIRED when transit_encryption_enabled = true. Sensitive, no default."
  type        = string
  sensitive   = true
  default     = null

  validation {
    condition     = var.auth_token != null || !var.transit_encryption_enabled
    error_message = "ElastiCache requires an auth_token when transit encryption is enabled."
  }
}

variable "snapshot_retention_limit" {
  description = "Days of automatic snapshots (0 = disabled)"
  type        = number
  default     = 7
}

variable "snapshot_window" {
  description = "Preferred snapshot window (UTC)"
  type        = string
  default     = "03:00-04:00"
}

variable "maintenance_window" {
  description = "Preferred maintenance window (UTC)"
  type        = string
  default     = "sun:06:00-sun:07:00"
}

variable "tags" {
  description = "Additional tags applied to all resources"
  type        = map(string)
  default     = {}
}
