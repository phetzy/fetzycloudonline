# The instance runs in the account's default VPC. This is a personal account
# with a single instance; a bespoke VPC would be ceremony without benefit.
data "aws_vpc" "default" {
  default = true
}

# ---------------------------------------------------------------------------
# Security group for the SSH TUI instance
# ---------------------------------------------------------------------------

resource "aws_security_group" "ssh_tui" {
  name        = "${var.project}-ssh-tui"
  description = "Security group for the SSH TUI instance"
  vpc_id      = data.aws_vpc.default.id
}

# World-open port for the public SSH TUI server is the entire point. This is
# an unauthenticated listener serving a résumé — not an admin port, not a
# management interface. sshd is disabled on the instance; this port belongs
# entirely to the TUI. Session Manager via SSM is the only admin path, and it
# establishes connections outbound-only.
resource "aws_vpc_security_group_ingress_rule" "ssh_ipv4" {
  security_group_id = aws_security_group.ssh_tui.id

  description = "SSH TUI from anywhere (IPv4)"
  from_port   = var.ssh_port
  to_port     = var.ssh_port
  ip_protocol = "tcp"
  cidr_ipv4   = "0.0.0.0/0"
}

resource "aws_vpc_security_group_ingress_rule" "ssh_ipv6" {
  security_group_id = aws_security_group.ssh_tui.id

  description = "SSH TUI from anywhere (IPv6)"
  from_port   = var.ssh_port
  to_port     = var.ssh_port
  ip_protocol = "tcp"
  cidr_ipv6   = "::/0"
}

# Unrestricted outbound: SSM agent needs HTTPS, and the S3 pull of the binary
# at boot needs it too.
resource "aws_vpc_security_group_egress_rule" "all" {
  security_group_id = aws_security_group.ssh_tui.id

  description = "Allow all outbound traffic"
  ip_protocol = "-1"
  cidr_ipv4   = "0.0.0.0/0"
}

resource "aws_vpc_security_group_egress_rule" "all_ipv6" {
  security_group_id = aws_security_group.ssh_tui.id

  description = "Allow all outbound traffic (IPv6)"
  ip_protocol = "-1"
  cidr_ipv6   = "::/0"
}
