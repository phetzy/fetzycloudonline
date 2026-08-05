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
data "aws_kms_alias" "ssm_default" {
  name = "alias/aws/ssm"
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
  statement {
    effect    = "Allow"
    actions   = ["ssm:GetParameter", "ssm:PutParameter"]
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
  # instance in the account. The instance itself isn't created until Task 7,
  # so it can't be referenced by resource yet; this is scoped as tightly as
  # possible in the meantime (every EC2 instance ARN in this account and
  # region, rather than `*`), and should be narrowed to the instance's own
  # ARN once Task 7 creates it.
  statement {
    effect  = "Allow"
    actions = ["ssm:SendCommand"]
    resources = [
      "arn:aws:ssm:${var.region}::document/AWS-RunShellScript",
      "arn:aws:ec2:${var.region}:${data.aws_caller_identity.current.account_id}:instance/*",
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
