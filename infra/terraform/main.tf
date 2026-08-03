# SchemaHub terraform root module — AWS deployment.
#
# Wires the four building blocks:
#   vpc  -> subnets/IGW/NAT (foundation)
#   rds  -> managed Postgres 16 for the SchemaHub control plane
#   redis-> ElastiCache Redis (cluster mode disabled)
#   ecs  -> Fargate backend + ALB (gRPC-Web entry point)
#
# Secrets flow: tfvars supplies the sensitive values (no defaults),
# the rds/redis modules store connection URLs in SSM SecureString, the ecs
# module writes app secrets to SecretsManager and references them from the
# task definition. Nothing secret is written into the state in plaintext.

terraform {
  required_version = ">= 1.8"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  # Uncomment once the S3 bucket + DynamoDB lock table exist (see provider.tf):
  # backend "s3" {
  #   bucket         = "schemahub-terraform-state"
  #   key            = "terraform.tfstate"
  #   region         = "us-east-1"
  #   encrypt        = true
  #   dynamodb_table = "terraform-locks"
  # }
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

locals {
  # Default to the region's AZs when the caller did not pin specific ones.
  azs = length(var.azs) > 0 ? var.azs : [
    for i in range(min(3, length(data.aws_availability_zones.available.names))) :
    data.aws_availability_zones.available.names[i]
  ]

  # Tag set forwarded to every module resource.
  tags = {
    Environment = var.environment
  }
}

module "vpc" {
  source = "./modules/vpc"

  name              = var.name
  vpc_cidr          = var.vpc_cidr
  azs               = local.azs
  az_count          = length(local.azs)
  nat_gateway_count = var.nat_gateway_count
  tags              = local.tags
}

module "rds" {
  source = "./modules/rds"

  name                      = var.name
  vpc_id                    = module.vpc.vpc_id
  subnet_ids                = module.vpc.private_subnet_ids
  allowed_security_group_id = module.ecs.app_security_group_id

  db_name     = var.db_name
  db_username = var.db_username
  db_password = var.db_password

  engine_version    = var.rds_engine_version
  instance_class    = var.rds_instance_class
  allocated_storage = var.rds_allocated_storage
  multi_az          = var.rds_multi_az

  tags = local.tags
}

module "redis" {
  source = "./modules/redis"

  name                      = var.name
  vpc_id                    = module.vpc.vpc_id
  subnet_ids                = module.vpc.private_subnet_ids
  allowed_security_group_id = module.ecs.app_security_group_id

  engine_version             = var.redis_engine_version
  node_type                  = var.redis_node_type
  num_cache_clusters         = var.redis_num_cache_clusters
  transit_encryption_enabled = var.redis_transit_encryption_enabled
  auth_token                 = var.redis_auth_token

  tags = local.tags
}

module "ecs" {
  source = "./modules/ecs"

  name               = var.name
  environment        = var.environment
  region             = var.region
  vpc_id             = module.vpc.vpc_id
  public_subnet_ids  = module.vpc.public_subnet_ids
  private_subnet_ids = module.vpc.private_subnet_ids

  backend_image = var.backend_image
  frontend_url  = var.frontend_url

  database_url    = module.rds.connection_url
  redis_url       = module.redis.connection_url
  jwt_private_key = var.jwt_private_key
  jwt_public_key  = var.jwt_public_key
  encryption_key  = var.encryption_key

  cpu             = var.backend_cpu
  memory          = var.backend_memory
  desired_count   = var.backend_desired_count
  min_capacity    = var.backend_min_capacity
  max_capacity    = var.backend_max_capacity
  db_pool_min     = var.db_pool_min
  db_pool_max     = var.db_pool_max
  certificate_arn = var.acm_certificate_arn

  tags = local.tags
}
