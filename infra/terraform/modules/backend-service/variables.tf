# Backend service module — ECS/Fargate or Fly.io (skeleton)
# TODO: implement aws_ecs_service / fly_service, env injection from secrets,
#       health check against gRPC reflection, autoscaling.

variable "name" {
  type = string
}

variable "image" {
  type = string
}

variable "database_url" {
  type      = string
  sensitive = true
}

variable "redis_url" {
  type      = string
  sensitive = true
}

variable "jwt_private_key" {
  type      = string
  sensitive = true
}

variable "jwt_public_key" {
  type      = string
  sensitive = true
}
