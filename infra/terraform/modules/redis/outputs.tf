# Redis module — managed Redis / ElastiCache (skeleton)
# TODO: implement aws_elasticache_replication_group (or Upstash), secret store.

variable "name" {
  type = string
}

variable "subnet_ids" {
  type = list(string)
}

output "connection_url" {
  description = "Redis connection URL for the backend"
  value       = ""
  sensitive   = true
}
