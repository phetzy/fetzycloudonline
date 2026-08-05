# ---------------------------------------------------------------------------
# The SSH TUI instance
# ---------------------------------------------------------------------------

locals {
  # The systemd unit, fully rendered before it is spliced into user_data.sh.
  # It must be rendered here (its own templatefile() call) rather than left
  # for user_data.sh to render itself, because Terraform does not re-scan
  # already-substituted text for template syntax on a second pass — plain
  # string interpolation is what makes embedding it safely into another
  # templatefile() call possible.
  service_unit_rendered = templatefile("${path.module}/files/fetzer.service", {
    ssh_port         = var.ssh_port
    idle_timeout     = var.idle_timeout
    max_session      = var.max_session
    max_conns_per_ip = var.max_conns_per_ip
    rate_per_min     = var.rate_per_min
  })

  # user_data.sh rendered at ~16,166 bytes is 218 bytes under EC2's hard
  # 16,384-byte user_data limit — close enough that any small change could
  # tip it over and fail RunInstances outright. base64gzip() below is what
  # actually protects against that (cloud-init decompresses gzip user data
  # transparently, and the gzipped form is roughly a third of the size), but
  # rendering it as its own local first keeps that reasoning next to the
  # value it applies to.
  user_data_rendered = templatefile("${path.module}/user_data.sh", {
    region             = var.region
    artifact_bucket    = aws_s3_bucket.artifacts.id
    host_key_parameter = aws_ssm_parameter.host_key.name
    service_unit       = local.service_unit_rendered
    update_script      = file("${path.module}/files/update-fetzer.sh")
  })
}

resource "aws_instance" "fetzer" {
  ami                    = data.aws_ssm_parameter.al2023_arm64.value
  instance_type          = var.instance_type
  iam_instance_profile   = aws_iam_instance_profile.fetzer.name
  vpc_security_group_ids = [aws_security_group.ssh_tui.id]

  # Gzipped: see the comment on local.user_data_rendered above. This is not
  # an optimization, it is what keeps RunInstances from failing.
  user_data_base64 = base64gzip(local.user_data_rendered)

  # A user_data change (a new host-key-fetch/service-unit render, etc.) forces
  # a new instance rather than silently leaving the old boot script's effects
  # in place forever. That's safe here specifically because the SSH host key
  # lives in aws_ssm_parameter.host_key, not on instance disk: a replacement
  # instance fetches the existing key from SSM (user_data.sh step 4) instead
  # of minting a new one, so replacement doesn't rotate the fingerprint under
  # returning visitors. Without user_data_replace_on_change, editing
  # user_data.sh would be a silent no-op on every subsequent apply, which is
  # the worse failure mode.
  user_data_replace_on_change = true

  metadata_options {
    http_tokens   = "required" # IMDSv2 only
    http_endpoint = "enabled"
  }

  # This is an unauthenticated public listener; under IMDSv1 an SSRF-shaped
  # bug in it would hand over the instance role's credentials to anyone who
  # asks. IMDSv2's session-token requirement is what closes that off.

  root_block_device {
    encrypted   = true
    volume_type = "gp3"
    volume_size = 8
  }
}

resource "aws_eip" "fetzer" {
  domain = "vpc"
}

# Associated separately (rather than via aws_instance.associate_public_ip_address)
# so the address survives a stop/start and, more importantly, an instance
# replacement triggered by user_data_replace_on_change above.
resource "aws_eip_association" "fetzer" {
  instance_id   = aws_instance.fetzer.id
  allocation_id = aws_eip.fetzer.id
}
