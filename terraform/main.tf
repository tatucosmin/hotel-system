variable "aws_access_key_id" {
    type = string
}

variable "aws_secret_access_key" {
    type = string
}

variable "aws_default_region" {
    type = string
}

variable "s3_bucket" {
    type = string
}

variable "localstack_s3_endpoint" {
    type = string
}

provider "aws" {

  access_key                  = var.aws_access_key_id
  secret_key                  = var.aws_secret_access_key
  region                      = var.aws_default_region

  s3_use_path_style           = true
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true

  endpoints {
    s3             = var.localstack_s3_endpoint
  }
}

resource "aws_s3_bucket" "ticketr-s3" {
  bucket = var.s3_bucket
}