output "connection_url" {
  description = "Redis connection URL for the backend (sensitive)"
  value       = aws_ssm_parameter.connection_url.value
  sensitive   = true
}

output "connection_url_ssm_parameter" {
  description = "Name of the SSM SecureString parameter holding the connection URL"
  value       = aws_ssm_parameter.connection_url.name
}

output "primary_endpoint_address" {
  description = "Primary endpoint hostname"
  value       = aws_elasticache_replication_group.this.primary_endpoint_address
}

output "port" {
  description = "Redis port (6379)"
  value       = aws_elasticache_replication_group.this.port
}

output "security_group_id" {
  description = "Security group id of the Redis cluster"
  value       = aws_security_group.this.id
}

output "replication_group_id" {
  description = "ElastiCache replication group id"
  value       = aws_elasticache_replication_group.this.id
}
