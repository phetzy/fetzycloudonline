terraform {
  required_version = ">= 1.11.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  # Native S3 locking — verified supported by the installed tofu 1.11.4, so no
  # DynamoDB table is needed. The bucket comes from terraform/bootstrap.
  backend "s3" {
    key          = "ssh-tui/terraform.tfstate"
    region       = "us-west-2"
    encrypt      = true
    use_lockfile = true
    # bucket is supplied at init time: see terraform/README.md
  }
}

provider "aws" {
  region = var.region

  default_tags {
    tags = {
      Project   = var.project
      ManagedBy = "opentofu"
      Repo      = "phetzy/fetzycloudonline"
    }
  }
}

data "aws_caller_identity" "current" {}

# Amazon Linux 2023, arm64, always current rather than a pinned AMI ID.
data "aws_ssm_parameter" "al2023_arm64" {
  name = "/aws/service/ami-amazon-linux-latest/al2023-ami-kernel-default-arm64"
}
