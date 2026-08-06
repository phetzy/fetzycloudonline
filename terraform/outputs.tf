# Every output the owner needs after `tofu apply` — to reach the box, and to
# populate the four GitHub repository settings the deploy workflow
# (.github/workflows/deploy-ssh.yml) requires. See terraform/README.md for
# which output feeds which GitHub secret or variable, and why the split
# between the two is not arbitrary.
#
# artifact_bucket and github_deploy_role_arn used to live in storage.tf and
# iam.tf's neighbor respectively (artifact_bucket was added in Task 3,
# github_deploy_role_arn in Task 4, before this file existed) — consolidated
# here as the single place to look for every output.

output "public_ip" {
  description = "Elastic IP address of the SSH TUI instance. This is what a visitor connects to: `ssh -p <ssh_port> <public_ip>` (or a hostname pointed at it — see the DNS section of terraform/README.md, which is unresolved as of this apply)."
  value       = aws_eip.fetzer.public_ip
}

output "connect_command" {
  description = "The ssh command a visitor (or the owner, to smoke-test) would run to reach the TUI, with the current public IP and configured port already filled in."
  value       = "ssh -p ${var.ssh_port} ${aws_eip.fetzer.public_ip}"
}

output "instance_id" {
  description = "EC2 instance ID. Required as the GitHub repository variable INSTANCE_ID — the deploy role deliberately has no ec2:DescribeInstances or ssm:DescribeInstanceInformation, so the workflow cannot discover this on its own. Without it every deploy run fails at the SSM step."
  value       = aws_instance.fetzer.id
}

output "region" {
  description = "AWS region the infrastructure is deployed into. Required as the GitHub repository variable AWS_REGION."
  value       = var.region
}

# Embeds the account ID (bucket name is
# "<project>-ssh-tui-artifacts-<account-id>", see storage.tf). This value
# goes into the GitHub repository *secret* ARTIFACT_BUCKET, not a variable —
# see terraform/README.md for why. Marked sensitive so it does not appear in
# plain sight in unattended `tofu apply`/`plan` console output; read it
# explicitly with `tofu output -raw artifact_bucket` when it's actually
# needed for pasting into GitHub.
output "artifact_bucket" {
  description = "S3 bucket the deploy workflow uploads built binaries to. Required as the GitHub repository secret ARTIFACT_BUCKET."
  value       = aws_s3_bucket.artifacts.id
  sensitive   = true
}

# Embeds the account ID (an IAM role ARN is
# arn:aws:iam::<account-id>:role/<name>). Same reasoning and same treatment
# as artifact_bucket above: GitHub repository *secret* DEPLOY_ROLE_ARN, and
# sensitive = true here for the same defense-in-depth reason.
output "github_deploy_role_arn" {
  description = "ARN of the IAM role GitHub Actions assumes via OIDC to deploy. Required as the GitHub repository secret DEPLOY_ROLE_ARN."
  value       = aws_iam_role.github_deploy.arn
  sensitive   = true
}
