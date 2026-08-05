# Design: SSH TUI infrastructure — OpenTofu on AWS, deployed by GitHub Actions

Date: 2026-08-05
Status: approved, ready for implementation planning
Branch: `infra-ssh`

## Scope

Sub-project 2b, and the last piece of the SSH deliverable. It puts the Go SSH TUI —
built and merged in 2a — onto a public address, and gives it a deploy path.

Two things are explicitly **not** in scope: running `tofu apply`, and changing DNS. This
sub-project produces configuration and documentation. The owner applies it.

## What already exists

- `cmd/fetzer` — a Wish + Bubble Tea SSH server, merged to `main`, verified working over
  a real `ssh` connection. PTY-only, no shell or exec or subsystem path, per-IP
  concurrency and rate limits, idle and max-session timeouts, a session-channel cap, and
  it refuses to start on non-positive security config.
- `content.json` — shared with the React website deployed at fetzycloud.online.
- A React site on Vercel, which this sub-project must not disturb.

## Decisions taken

| Decision | Choice |
| --- | --- |
| Tool | OpenTofu 1.11.4 (the installed binary; 1.12.1 is in the CachyOS repos but not what is on PATH) |
| Region | `us-west-2` |
| Instance | `t4g.micro`, Graviton arm64, Amazon Linux 2023 |
| Address | One Elastic IP |
| Admin access | SSM Session Manager. **`sshd` is disabled** — port 22 belongs to the TUI |
| Host key | SSM Parameter Store `SecureString`, so it survives instance replacement |
| Binary delivery | GitHub Actions → S3 → `ssm send-command` |
| CI credentials | GitHub OIDC federation. No long-lived AWS keys in the repository |
| Deploy trigger | Push to `main` touching `cmd/`, `internal/`, `content.json`, or `go.mod` |
| State | S3 backend, bootstrapped by a separate config with local state |
| Local credentials | `AWS_PROFILE=resume-tui`, with `AdministratorAccess` |

## Layout

```
terraform/
  bootstrap/            local state, run once
    main.tf             the state bucket: versioned, encrypted, public access blocked
    README.md
  main.tf               providers, backend, data sources
  network.tf            security group
  compute.tf            instance, EIP, user_data
  iam.tf                instance profile; the GitHub OIDC role (provider is a data source)
  storage.tf            artifact bucket, host-key parameter
  variables.tf
  outputs.tf
  user_data.sh          boot script
  fetzer.service        systemd unit
.github/workflows/
  deploy-ssh.yml
```

## The GitHub OIDC provider already exists

The account already has an IAM OIDC provider for
`token.actions.githubusercontent.com`, from earlier work. Creating a second one fails
with `EntityAlreadyExists`, which would break the first apply partway through.

So this config **reads it with a data source** rather than creating it, and creates only
the role that trusts it. That also means `tofu destroy` will not remove a provider other
projects in the account may depend on — the right behaviour for a shared account-level
resource this config does not own.

## Account identity is never committed

The state bucket and the artifact bucket both need globally unique names, and the usual
shortcut is to append the AWS account ID. This repository is **public**, so the account
ID is resolved at apply time from `data.aws_caller_identity.current.account_id` and
interpolated into bucket names. It lands in state — which is private — and never in a
committed file. Account IDs are not secrets, but they are an identifier worth not
publishing.

## The instance

Amazon Linux 2023 on `t4g.micro`. `user_data` runs once at first boot:

1. Create a `fetzer` system user with no login shell.
2. **Disable and mask `sshd`.** The TUI binds 22; a second thing listening there is a
   conflict, and an admin SSH port is exactly what SSM Session Manager exists to replace.
3. Fetch the host key from SSM Parameter Store. If the parameter does not exist —
   the first boot ever — generate an ed25519 key, store it as a `SecureString`, and
   continue. Every later boot, including a replacement instance, reads the same key, so
   returning visitors never see a changed-fingerprint warning.
4. Pull the binary from the artifact bucket, if one has been published.
5. Install and start the systemd unit.

`IMDSv2` is required (`http_tokens = "required"`). The TUI is an unauthenticated network
service; if it ever has an SSRF-shaped bug, IMDSv1 would hand over the instance role's
credentials. The root EBS volume is encrypted.

### The systemd unit

Runs as `fetzer`, not root, with `AmbientCapabilities=CAP_NET_BIND_SERVICE` so it can
bind port 22 without privilege. Hardened with `NoNewPrivileges`, `ProtectSystem=strict`,
`ProtectHome`, `PrivateTmp`, `RestrictAddressFamilies=AF_INET AF_INET6`, and a
`MemoryMax`. Restarts on failure with a backoff.

The memory ceiling matters more than it looks: the ledger from 2a records that the SSH
library accumulates `env` requests per channel without a cap, and that panic coverage is
partial. `MemoryMax` turns an unbounded-growth bug into a restart rather than an
instance-wide OOM.

## Security group

Inbound: `22/tcp` from `0.0.0.0/0`. That is the entire point — an unauthenticated public
listener.

Nothing else inbound. No admin port, because there is no admin service: SSM Session
Manager reaches the instance outbound-only through the SSM agent.

Outbound: unrestricted, which the SSM agent and the S3 pull both need.

## Deploy

A push to `main` touching the Go paths triggers the workflow:

1. Assume the OIDC role. The trust policy is scoped to
   `repo:phetzy/fetzycloudonline:ref:refs/heads/main` — not the whole repository, and not
   `*`, either of which would let any branch or any fork's pull request assume it.
2. Build `GOOS=linux GOARCH=arm64 go build -trimpath`.
3. Upload to the artifact bucket under a commit-SHA key, and update a `current` pointer.
4. `ssm send-command` runs the instance's update script: pull, verify, restart, and
   report.

The permissions the CI role gets are: write to the artifact bucket, and `SendCommand`
against that one instance for that one document. Not `ssm:*`, and not a wildcard resource.

## What this deliberately does not do

- **No DNS.** `ssh fetzycloud.online` will still not work after this. The apex A record
  points at Vercel, and this creates an Elastic IP with nothing pointed at it. A
  subdomain such as `ssh.fetzycloud.online` is the low-friction answer, but that is the
  owner's call and remains deferred. The README documents both paths and states plainly
  that the address is an IP until someone decides.
- **No apply.** Everything here is configuration. The owner runs `tofu apply`.
- **No changes to the website or to `content.json`.**

## Testing and verification

Infrastructure code cannot be unit-tested the way the TUI was. What is available:

- `tofu fmt -check`, `tofu init -backend=false`, and `tofu validate` on both configs — catches syntax,
  type, and reference errors before anything reaches AWS.
- `actionlint` on the workflow, and a manual read of the OIDC trust policy against
  GitHub's documented claim format, since a wrong `sub` condition either fails closed
  (annoying) or opens the role to other repositories (serious).
- `shellcheck` on `user_data.sh` and the update script.

What is **not** available without an apply: whether `user_data` actually completes,
whether the OIDC trust policy is accepted by AWS, and whether `ssm send-command` reaches
the instance. Those are first-apply findings. The README will say so rather than implying
the config is proven.

## Residual risks carried in from 2a

These live in the deployed program, not in this configuration, and putting it on a public
address is what makes them matter:

- Panic coverage is partial. `wish/recover` protects the Bubble Tea program and the
  middleware it wraps; `loggingMiddleware`'s own body and the SSH library's goroutines
  remain process-fatal. `Restart=on-failure` in the unit converts a crash into downtime
  measured in seconds rather than until someone notices.
- The SSH library accumulates `env` requests per channel without a cap. `MemoryMax`
  bounds the blast radius.
- No global connection cap and no `HandshakeTimeout` — per-IP is capped at 3, total is
  not.

None of these block putting it up. All are worth a follow-up, and the honest framing is
that this is a personal site, not a service with an availability target.

## Out of scope

Monitoring and alerting, backups (there is no state to lose — the host key is in SSM and
everything else is rebuilt from the repository), multi-region anything, autoscaling, and
a CDN.
