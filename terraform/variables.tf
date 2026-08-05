locals {
  # Reference timestamp used only to validate the idle_timeout/max_session
  # duration strings below via timeadd/timecmp — chosen arbitrarily, its
  # value has no other significance.
  duration_validation_epoch = "2000-01-01T00:00:00Z"
}

variable "region" {
  type        = string
  default     = "us-west-2"
  description = "AWS region the SSH TUI infrastructure is deployed into."
}

variable "project" {
  type        = string
  default     = "fetzer"
  description = "Project name applied as the Project default tag on every resource."
}

variable "instance_type" {
  type        = string
  default     = "t4g.micro"
  description = "EC2 instance type for the SSH TUI host. Must be an arm64-compatible type to match the AL2023 arm64 AMI."
}

variable "github_repo" {
  type        = string
  default     = "phetzy/fetzycloudonline"
  description = "GitHub repository (owner/name) the instance pulls the SSH TUI build from."

  # This value is interpolated verbatim into the deploy role's trust policy
  # sub condition (StringEquals, not StringLike), so a malformed value fails
  # closed rather than widening trust — but a bad value would still surface
  # as an opaque STS AssumeRoleWithWebIdentity denial deep inside CI, which
  # is expensive to diagnose. Fail fast here instead.
  validation {
    condition     = can(regex("^[^/]+/[^/]+$", var.github_repo))
    error_message = "github_repo must be in \"owner/name\" form, e.g. \"phetzy/fetzycloudonline\"."
  }
}

variable "github_ref" {
  type        = string
  default     = "refs/heads/main"
  description = "GitHub ref (branch or tag) the instance builds/deploys from."

  # Same reasoning as github_repo's validation: this is interpolated into
  # the trust policy's sub condition, and a value like "main" instead of
  # "refs/heads/main" is an easy mistake that otherwise fails closed with an
  # opaque STS denial in CI rather than here at plan time.
  validation {
    condition     = can(regex("^refs/(heads|tags)/.+$", var.github_ref))
    error_message = "github_ref must be a full ref, e.g. \"refs/heads/main\" or \"refs/tags/v1\" — not a bare branch or tag name."
  }
}

variable "ssh_port" {
  type        = number
  default     = 22
  description = "TCP port the security group opens for inbound SSH TUI connections."

  validation {
    condition     = var.ssh_port > 0 && var.ssh_port <= 65535
    error_message = "ssh_port must be a valid TCP port between 1 and 65535."
  }
}

variable "idle_timeout" {
  type        = string
  default     = "5m"
  description = "Duration string (Go time.ParseDuration syntax) passed to the TUI server as -idle-timeout: disconnects a session after this long with no activity. The Go program refuses to start on a non-positive value, so an invalid setting here produces a boot-crash-looping instance rather than a running one with the control silently disabled."

  validation {
    # timeadd parses Go duration syntax, so this accepts any value the Go
    # program itself accepts, including compound durations like "1h30m", and
    # can(...) makes an unparseable value ("banana") fail this condition
    # instead of erroring out entirely. timecmp then rejects zero and
    # negative durations, including zero-padded forms like "00s" that a
    # regex-based check misses. A regex was tried first and dropped: it
    # either let zero-padded zero durations through or rejected legitimate
    # compound durations, whereas timeadd/timecmp gets both parseability and
    # positivity right in one expression.
    condition     = can(timecmp(timeadd(local.duration_validation_epoch, var.idle_timeout), local.duration_validation_epoch)) && timecmp(timeadd(local.duration_validation_epoch, var.idle_timeout), local.duration_validation_epoch) > 0
    error_message = "idle_timeout must be a positive Go duration string (e.g. \"5m\" or \"1h30m\"), matching the TUI's -idle-timeout flag."
  }
}

variable "max_session" {
  type        = string
  default     = "30m"
  description = "Duration string (Go time.ParseDuration syntax) passed to the TUI server as -max-session: hard cap on a single session's total duration, active or not. The Go program refuses to start on a non-positive value, so an invalid setting here produces a boot-crash-looping instance rather than a running one with the control silently disabled."

  validation {
    # See idle_timeout's validation comment above for why timeadd/timecmp is
    # used instead of a regex.
    condition     = can(timecmp(timeadd(local.duration_validation_epoch, var.max_session), local.duration_validation_epoch)) && timecmp(timeadd(local.duration_validation_epoch, var.max_session), local.duration_validation_epoch) > 0
    error_message = "max_session must be a positive Go duration string (e.g. \"30m\" or \"1h30m\"), matching the TUI's -max-session flag."
  }
}

variable "max_conns_per_ip" {
  type        = number
  default     = 3
  description = "Maximum concurrent connections allowed from a single IP, passed to the TUI server as -max-conns-per-ip. The Go program refuses to start on a non-positive value."

  validation {
    condition     = var.max_conns_per_ip > 0
    error_message = "max_conns_per_ip must be a positive integer, matching the TUI's -max-conns-per-ip flag."
  }
}

variable "rate_per_min" {
  type        = number
  default     = 20
  description = "Maximum new connections allowed from a single IP per minute, passed to the TUI server as -rate-per-min. The Go program refuses to start on a non-positive value."

  validation {
    condition     = var.rate_per_min > 0
    error_message = "rate_per_min must be a positive integer, matching the TUI's -rate-per-min flag."
  }
}
