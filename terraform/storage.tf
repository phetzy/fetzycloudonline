locals {
  # The account ID never appears in a committed file; it is resolved at apply
  # time. This repository is public.
  artifact_bucket_name = "${var.project}-ssh-tui-artifacts-${data.aws_caller_identity.current.account_id}"
}

resource "aws_s3_bucket" "artifacts" {
  bucket = local.artifact_bucket_name

  # Unlike state, everything here is rebuildable from a commit, so no
  # prevent_destroy.
}

resource "aws_s3_bucket_versioning" "artifacts" {
  bucket = aws_s3_bucket.artifacts.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "artifacts" {
  bucket = aws_s3_bucket.artifacts.id
  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket_public_access_block" "artifacts" {
  bucket                  = aws_s3_bucket.artifacts.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_lifecycle_configuration" "artifacts" {
  bucket = aws_s3_bucket.artifacts.id

  rule {
    id     = "expire-noncurrent-versions"
    status = "Enabled"

    filter {}

    noncurrent_version_expiration {
      noncurrent_days = 30
    }
  }
}

# Deny any request to the bucket that isn't over TLS. The public access
# block stops anonymous access; it does nothing about an authenticated
# request over plain HTTP.
data "aws_iam_policy_document" "artifacts_deny_insecure_transport" {
  statement {
    sid    = "DenyInsecureTransport"
    effect = "Deny"

    principals {
      type        = "*"
      identifiers = ["*"]
    }

    actions = ["s3:*"]

    resources = [
      aws_s3_bucket.artifacts.arn,
      "${aws_s3_bucket.artifacts.arn}/*",
    ]

    condition {
      test     = "Bool"
      variable = "aws:SecureTransport"
      values   = ["false"]
    }
  }
}

resource "aws_s3_bucket_policy" "artifacts" {
  bucket = aws_s3_bucket.artifacts.id
  policy = data.aws_iam_policy_document.artifacts_deny_insecure_transport.json

  # Applying a bucket policy before the public access block is in place can
  # race BlockPublicPolicy; make the ordering explicit rather than relying
  # on implicit dependency inference.
  depends_on = [aws_s3_bucket_public_access_block.artifacts]
}

resource "aws_ssm_parameter" "host_key" {
  name  = "/${var.project}/ssh/host_key"
  type  = "SecureString"
  value = "placeholder-replaced-on-first-boot"

  # The instance generates the real SSH host key at first boot and writes it
  # here itself — OpenTofu never sets the real value. Without
  # ignore_changes, every routine `tofu apply` would revert this back to the
  # placeholder, changing the host key fingerprint and triggering a
  # man-in-the-middle warning in the SSH client of every returning visitor.
  # This parameter existing in SSM (rather than only on the instance's disk)
  # is what lets the key survive instance replacement; ignore_changes is
  # what makes that actually work.
  lifecycle {
    ignore_changes = [value]
  }
}

output "artifact_bucket" {
  value = aws_s3_bucket.artifacts.id
}

# Task 8 needs this ARN as a literal string in the GitHub Actions YAML, and
# it contains the account ID, which must never be committed. This output is
# how it gets read at apply time and copied into a repo secret by hand.
output "github_deploy_role_arn" {
  value = aws_iam_role.github_deploy.arn
}
