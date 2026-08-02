# SchemaHub terraform root module.
# NOTE: This is a skeleton. Modules are declared but not yet implemented —
# see infra/README.md. Nothing here provisions real resources yet.

terraform {
  required_version = ">= 1.8"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

# The module calls below are commented out until each module is filled in.
# module "networking" {
#   source = "./modules/networking"
#   name   = var.name
#   vpc_cidr = var.vpc_cidr
# }
#
# module "database" {
#   source  = "./modules/database"
#   name    = var.name
#   subnet_ids = module.networking.private_subnet_ids
# }
#
# module "redis" {
#   source  = "./modules/redis"
#   name    = var.name
#   subnet_ids = module.networking.private_subnet_ids
# }
#
# module "backend_service" {
#   source       = "./modules/backend-service"
#   name         = var.name
#   image        = var.backend_image
#   database_url = module.database.connection_url
#   redis_url    = module.redis.connection_url
#   jwt_private_key = var.jwt_private_key
#   jwt_public_key  = var.jwt_public_key
# }
