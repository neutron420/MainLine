# AWS provider configuration for SchemaHub.
#
# Remote state: when this becomes a shared/CI-managed stack, enable an S3
# backend (bucket must exist before `terraform init`):
#
#   terraform {
#     backend "s3" {
#       bucket         = "schemahub-terraform-state"
#       key            = "terraform.tfstate"
#       region         = "us-east-1"
#       encrypt        = true
#       dynamodb_table = "terraform-locks"
#     }
#   }
#
# The backend block lives in the root terraform block (see main.tf) because
# it cannot be parameterized. The commented block above is the shape to use;
# keep the DynamoDB lock table out of band (it must exist before init).

provider "aws" {
  region = var.region

  default_tags {
    tags = {
      Project     = "SchemaHub"
      Environment = var.environment
      ManagedBy   = "terraform"
    }
  }
}
