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
}

variable "github_ref" {
  type        = string
  default     = "refs/heads/main"
  description = "GitHub ref (branch or tag) the instance builds/deploys from."
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
    condition     = can(regex("^[0-9]+(ns|us|µs|ms|s|m|h)$", var.idle_timeout)) && !can(regex("^0(ns|us|µs|ms|s|m|h)$", var.idle_timeout))
    error_message = "idle_timeout must be a positive Go duration string (e.g. \"5m\"), matching the TUI's -idle-timeout flag."
  }
}

variable "max_session" {
  type        = string
  default     = "30m"
  description = "Duration string (Go time.ParseDuration syntax) passed to the TUI server as -max-session: hard cap on a single session's total duration, active or not. The Go program refuses to start on a non-positive value, so an invalid setting here produces a boot-crash-looping instance rather than a running one with the control silently disabled."

  validation {
    condition     = can(regex("^[0-9]+(ns|us|µs|ms|s|m|h)$", var.max_session)) && !can(regex("^0(ns|us|µs|ms|s|m|h)$", var.max_session))
    error_message = "max_session must be a positive Go duration string (e.g. \"30m\"), matching the TUI's -max-session flag."
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
