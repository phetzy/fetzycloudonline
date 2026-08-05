# Already exists in this account from earlier work. Creating a second one
# fails with EntityAlreadyExists, and destroying this config must not remove
# a provider other projects may depend on.
data "aws_iam_openid_connect_provider" "github" {
  url = "https://token.actions.githubusercontent.com"
}

# The account's default AWS-managed key for SSM SecureString parameters.
# Referenced by its alias, but IAM identity policies must grant against the
# underlying CMK's ARN (an alias ARN in a resource element does not by
# itself grant access to the key), so the target key ARN is what the
# instance policy below actually uses.
#
# AWS creates the aws/ssm managed key lazily, on the first SecureString ever
# written in the region/account — until then the alias exists but its
# target key ID is null, and this data source resolves to an empty ARN. The
# depends_on forces this read to happen at apply time, after
# aws_ssm_parameter.host_key (a SecureString) has been created and the key
# is guaranteed to exist, rather than at plan time before it does. It is a
# no-op once the key exists, so it stays correct on every later apply. Do
# not remove this as "redundant" — without it, the first apply in a fresh
# account silently grants access to nothing.
data "aws_kms_alias" "ssm_default" {
  name       = "alias/aws/ssm"
  depends_on = [aws_ssm_parameter.host_key]
}

# ---------------------------------------------------------------------------
# EC2 instance role and profile
# ---------------------------------------------------------------------------

data "aws_iam_policy_document" "instance_assume_role" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ec2.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "instance" {
  name               = "${var.project}-ssh-tui-instance"
  assume_role_policy = data.aws_iam_policy_document.instance_assume_role.json
}

# Session Manager is the only admin path onto the box (sshd is disabled), so
# the instance role needs the managed policy that lets the SSM agent
# register and accept sessions/commands.
resource "aws_iam_role_policy_attachment" "instance_ssm_core" {
  role       = aws_iam_role.instance.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}

data "aws_iam_policy_document" "instance_permissions" {
  # The binary. ListBucket is separate from GetObject so a missing object
  # comes back as a clean 404 instead of an ambiguous 403.
  statement {
    effect    = "Allow"
    actions   = ["s3:GetObject"]
    resources = ["${aws_s3_bucket.artifacts.arn}/*"]
  }

  statement {
    effect    = "Allow"
    actions   = ["s3:ListBucket"]
    resources = [aws_s3_bucket.artifacts.arn]
  }

  # The instance generates its own SSH host key at first boot and writes it
  # here itself, so it needs both read and write on this one parameter.
  # GetParameters (plural) is included alongside GetParameter (singular)
  # because they're distinct IAM actions mapped to distinct API calls/agent
  # resolution paths (e.g. `aws ssm get-parameters` or the SSM agent's
  # {{ssm-secure:...}} substitution) — the boot script for Task 7 hasn't
  # been written yet, and SSM is the only way onto the box if it guesses
  # wrong. Same resource ARN either way, so nothing extra is exposed.
  statement {
    effect    = "Allow"
    actions   = ["ssm:GetParameter", "ssm:GetParameters", "ssm:PutParameter"]
    resources = [aws_ssm_parameter.host_key.arn]
  }

  # Required to read/write the SecureString above.
  statement {
    effect    = "Allow"
    actions   = ["kms:Decrypt", "kms:Encrypt"]
    resources = [data.aws_kms_alias.ssm_default.target_key_arn]
  }
}

resource "aws_iam_role_policy" "instance" {
  name   = "${var.project}-ssh-tui-instance"
  role   = aws_iam_role.instance.id
  policy = data.aws_iam_policy_document.instance_permissions.json
}

resource "aws_iam_instance_profile" "fetzer" {
  name = "${var.project}-ssh-tui-instance"
  role = aws_iam_role.instance.name
}

# ---------------------------------------------------------------------------
# GitHub Actions deploy role (assumed via the existing OIDC provider)
# ---------------------------------------------------------------------------

data "aws_iam_policy_document" "github_trust" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRoleWithWebIdentity"]

    principals {
      type        = "Federated"
      identifiers = [data.aws_iam_openid_connect_provider.github.arn]
    }

    condition {
      test     = "StringEquals"
      variable = "token.actions.githubusercontent.com:aud"
      values   = ["sts.amazonaws.com"]
    }

    # Scoped to one ref on one repository. `repo:owner/name:*` would let any
    # branch — including one pushed by a fork's pull request — assume this
    # role. StringEquals, not StringLike, so no wildcard can creep in.
    condition {
      test     = "StringEquals"
      variable = "token.actions.githubusercontent.com:sub"
      values   = ["repo:${var.github_repo}:ref:${var.github_ref}"]
    }
  }
}

resource "aws_iam_role" "github_deploy" {
  name               = "${var.project}-ssh-tui-github-deploy"
  assume_role_policy = data.aws_iam_policy_document.github_trust.json
}

data "aws_iam_policy_document" "github_deploy_permissions" {
  statement {
    effect    = "Allow"
    actions   = ["s3:PutObject"]
    resources = ["${aws_s3_bucket.artifacts.arn}/*"]
  }

  # ssm:SendCommand requires both the document and the target instance to be
  # granted — granting only the document would let this role command every
  # instance in the account. Scoped to the one instance this role should ever
  # touch.
  statement {
    effect  = "Allow"
    actions = ["ssm:SendCommand"]
    resources = [
      "arn:aws:ssm:${var.region}::document/AWS-RunShellScript",
      aws_instance.fetzer.arn,
    ]
  }

  # AWS does not support resource-level permissions for GetCommandInvocation.
  statement {
    effect    = "Allow"
    actions   = ["ssm:GetCommandInvocation"]
    resources = ["*"]
  }
}

resource "aws_iam_role_policy" "github_deploy" {
  name   = "${var.project}-ssh-tui-github-deploy"
  role   = aws_iam_role.github_deploy.id
  policy = data.aws_iam_policy_document.github_deploy_permissions.json
}
