output "alb_dns_name" {
  description = "gRPC-Web entry point (point NEXT_PUBLIC_API_URL at http://<dns>)"
  value       = module.ecs.alb_dns_name
}

output "alb_arn" {
  description = "ARN of the ALB"
  value       = module.ecs.alb_arn
}

output "ecs_cluster_name" {
  description = "ECS cluster name"
  value       = module.ecs.cluster_name
}

output "ecs_service_name" {
  description = "ECS service name (backend)"
  value       = module.ecs.service_name
}

output "ecs_log_group" {
  description = "CloudWatch log group for backend logs"
  value       = module.ecs.log_group_name
}

output "vpc_id" {
  description = "VPC id"
  value       = module.vpc.vpc_id
}

output "public_subnet_ids" {
  description = "Public subnet ids"
  value       = module.vpc.public_subnet_ids
}

output "private_subnet_ids" {
  description = "Private subnet ids"
  value       = module.vpc.private_subnet_ids
}

output "database_host" {
  description = "RDS endpoint (SSM SecureString also holds the full connection URL)"
  value       = module.rds.host
}

output "database_connection_url_ssm_parameter" {
  description = "SSM parameter holding the RDS connection URL (SecureString)"
  value       = module.rds.connection_url_ssm_parameter
}

output "redis_primary_endpoint" {
  description = "ElastiCache primary endpoint"
  value       = module.redis.primary_endpoint_address
}

output "redis_connection_url_ssm_parameter" {
  description = "SSM parameter holding the Redis connection URL (SecureString)"
  value       = module.redis.connection_url_ssm_parameter
}

output "app_security_group_id" {
  description = "Application security group id (attach additional ingress as needed)"
  value       = module.ecs.app_security_group_id
}
