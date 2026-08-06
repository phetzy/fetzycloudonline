# Terraform: SSH TUI infrastructure

Puts the Go SSH TUI (`cmd/fetzer`, see the root README's [SSH TUI](../README.md#ssh-tui)
section) on a public AWS address: one EC2 instance, an Elastic IP, an S3
artifact bucket, and a GitHub Actions deploy path over SSM. This document is
written for the owner reading it months from now, after the reasoning behind
these decisions has left their head — it tries to say not just what to run
but why, especially where the "why" would otherwise only live in a commit
message.

## Bootstrap, then apply

This config's state lives in S3, but the bucket that holds that state has to
exist before the config can use it as a backend — chicken and egg.
`terraform/bootstrap/` breaks that cycle: a small, separate config with local
state that creates just the state bucket. Read `terraform/bootstrap/README.md`
for the detail; the short version:

```bash
cd terraform/bootstrap
tofu init
tofu apply
```

Note the `state_bucket` output — that's the bucket name the main config needs
next.

The main config's backend block (`terraform/main.tf`) deliberately omits
`bucket`. This repository is public, and the bucket name is
`fetzer-tfstate-<ACCOUNT_ID>` — if `bucket` were written into the committed
backend block, the account ID would be committed too. Instead it's supplied
at `init` time, from the bootstrap output:

```bash
cd terraform
tofu init -backend-config="bucket=<state_bucket output from bootstrap>"
tofu apply
```

This is a one-time step per machine you run `tofu` from (the backend config
is cached locally in `.terraform/`, itself gitignored). If you ever need to
re-init — a new machine, a deleted `.terraform/` directory — repeat the
`-backend-config` flag; there's no other way to supply the bucket name
without committing it.

One prerequisite this config assumes rather than creates: an IAM OIDC
identity provider for `token.actions.githubusercontent.com` must already
exist in the account. `iam.tf` reads it with a data source
(`data.aws_iam_openid_connect_provider.github`) rather than creating one,
because this account already had one from earlier work, and a second
`aws_iam_openid_connect_provider` for the same URL fails with
`EntityAlreadyExists`. In this account that's a non-issue. In a fresh
account with no prior GitHub Actions OIDC setup, the very first `tofu apply`
fails at plan time with something like "no matching
IAM OpenID Connect provider found" — create the provider first (AWS
Console → IAM → Identity providers → Add provider, or
`aws iam create-open-id-connect-provider`) and then apply.

## `tofu apply` is never a safe no-op

The AMI is resolved from `/aws/service/ami-amazon-linux-latest/al2023-ami-kernel-default-arm64`
(`main.tf`) — it always tracks the *latest* AL2023 arm64 build, not a pinned
ID, and `ami` is one of the attributes that forces instance replacement.
Amazon republishes that AMI roughly monthly. That means any `tofu apply` run
after a republish — even one changing nothing you wrote — replaces the
instance on AWS's schedule, not yours. Expect a brief outage (a few minutes:
instance terminate, new instance boot, `user_data` re-run) every time this
happens. The Elastic IP survives — it's associated via a separate
`aws_eip_association` resource specifically so it isn't tied to the
instance's lifecycle — so the address itself never changes, only what's
listening behind it for a few minutes.

If a replacement ever happens *outside* of an apply you ran deliberately (say,
you run `tofu apply` for an unrelated reason weeks from now and it quietly
picks up a new AMI), see "Deploys start failing with AccessDenied" below —
it's the same event with a different symptom.

## The GitHub repository configuration

The deploy workflow (`.github/workflows/deploy-ssh.yml`) needs four pieces of
repository configuration: two **secrets** and two **variables**. GitHub
Actions treats these very differently — variables are visible in plain text
anywhere a workflow log echoes them; secrets are masked. Get the split wrong
and this repository's *public* Actions logs print exactly the string this
whole config exists to keep out of the tree.

Set these as **secrets** (Settings → Secrets and variables → Actions →
Secrets):

| Secret            | Source                                             |
| ------------------ | --------------------------------------------------- |
| `DEPLOY_ROLE_ARN`   | `tofu output -raw github_deploy_role_arn`           |
| `ARTIFACT_BUCKET`   | `tofu output -raw artifact_bucket`                  |

Both of these embed the AWS account ID directly — the bucket name is
`<project>-ssh-tui-artifacts-<ACCOUNT_ID>` (`storage.tf`), and an IAM role
ARN is `arn:aws:iam::<ACCOUNT_ID>:role/<name>`. Both outputs are marked
`sensitive = true` in `outputs.tf` for the same reason: it costs one extra
`-raw` flag to read them, in exchange for them not printing in plain
`tofu output` or unattended `apply`/`plan` console output. This mirrors a
mistake caught during development of this branch — the first draft of the
deploy workflow put these two values in `vars`, on the reasoning that
"neither is sensitive," which was wrong: GitHub renders `vars` into the log
with no masking at all, sensitive or not, and this repo's Actions logs are
public. Every run would have printed the account ID.

Set these as **variables** (Settings → Secrets and variables → Actions →
Variables):

| Variable      | Source                            |
| -------------- | ------------------------------------ |
| `INSTANCE_ID`  | `tofu output -raw instance_id`      |
| `AWS_REGION`   | `tofu output -raw region`           |

Neither of these reveals anything on its own, so there's no cost to leaving
them visible in logs — and `INSTANCE_ID` in particular is *only* usable
through this workflow's own scoped IAM role, which grants `ssm:SendCommand`
against exactly one instance ARN, not the account in general.

`INSTANCE_ID` deserves its own note: the deploy role's IAM policy
(`iam.tf`) intentionally does **not** grant `ec2:DescribeInstances` or
`ssm:DescribeInstanceInformation` — least privilege, since the workflow only
ever needs to command the one instance it's scoped to, never enumerate
instances in the account. The consequence is that the workflow has no way to
look the instance ID up itself; it has to arrive out-of-band, as this
variable. Without it, the deploy workflow's SSM step cannot run at all.

## DNS: `ssh.fetzycloud.online` via Cloudflare

`fetzycloud.online` is on Cloudflare, with the apex `A` record pointed at
the Vercel-hosted website. This is decided, not deferred: add a subdomain,
not touch the apex.

In the Cloudflare dashboard, DNS → Records → Add record:

- Type `A`
- Name `ssh`
- IPv4 address: the `public_ip` output (`tofu output -raw public_ip` or
  `tofu output public_ip` — it isn't sensitive)
- **Proxy status: DNS only (grey cloud), not Proxied (orange cloud).**

That last point is the one to get right. The natural instinct is to flip
every Cloudflare record to the orange cloud, but Cloudflare's proxy only
understands HTTP/HTTPS; routing arbitrary TCP like SSH through it requires
Spectrum, a separate paid product, and this record doesn't have it enabled.
A proxied record here doesn't degrade — it stops port `<ssh_port>` from
answering on that hostname at all. Grey cloud only.

The apex and `www` records are untouched by this — they keep resolving to
Vercel exactly as before, so the website is never at risk from this change.

Verify once the record has propagated:

```bash
dig +short ssh.fetzycloud.online
ssh ssh.fetzycloud.online
```

There is deliberately no `AAAA` record alongside it. The security group
(`network.tf`) accepts inbound SSH from `::/0` as well as `0.0.0.0/0`, but
the instance runs in the account's default VPC, which has no IPv6 CIDR
block associated — the instance never gets an IPv6 address to publish, so
the `::/0` rule is currently unreachable rather than wrong.

Worth stating honestly: a grey-cloud record publishes the Elastic IP as the
DNS answer, in the clear, to anyone who looks — Cloudflare's proxy is what
normally hides an origin IP, and this record doesn't use it. That's an
acceptable trade here specifically because this is a deliberately public,
unauthenticated listener; the IP is not a secret this config is trying to
protect (`tofu output public_ip` already hands it out), and hiding it
behind a proxy would only be theater for a service anyone can already
connect to by design.

## First apply: what to expect, not a failure

Two things about a fresh `tofu apply` look like something went wrong and
aren't.

**The service starts enabled, not running.** `user_data.sh` installs and
enables the `fetzer` systemd unit, but does not start it if no binary has
ever been published (`if [ -x /usr/local/bin/fetzer ]`) — there's nothing to
run yet on a brand-new instance. Port `<ssh_port>` (default 22) is closed
until the first successful deploy from the GitHub Actions workflow installs
a binary and starts the unit. `ssh`-ing to the Elastic IP before that first
deploy will simply be refused; that's expected, not a misconfiguration.

That first deploy does not happen on its own. The workflow's `paths:`
filter (`.github/workflows/deploy-ssh.yml`) only fires on changes to the Go
source, `content.json`, `go.mod`/`go.sum`, or the workflow file itself, so a
Terraform-only or README-only commit — which is all a fresh apply typically
follows — will never trigger it, and there's nothing else to push right
after `tofu apply` finishes. Once the four repository secrets/variables
above are set, go to the repository's Actions tab, select the `deploy-ssh`
workflow, and run it manually on `main` via **Run workflow**
(`workflow_dispatch`, already wired into the `on:` block). That trigger is
accepted by the OIDC trust policy the same way a push to `main` is — both
produce the `repo:<owner>/<name>:ref:refs/heads/main` `sub` claim the trust
policy checks — so a manual run authenticates exactly like a normal deploy.

**The host key survives instance replacement; the diagnostic trail does
not.** The SSH host key is stored as an SSM `SecureString` parameter
(`aws_ssm_parameter.host_key` in `storage.tf`) with `ignore_changes =
[value]`, specifically so that instance replacement — whether from an AMI
refresh or anything else — fetches the existing key back out of SSM instead
of minting a new one, and returning visitors never see a host-key warning.
But `/var/log/fetzer-bootstrap.log` (the full boot log) and the `.prev`
rollback binary that `update-fetzer.sh` keeps both live only on the
instance's local disk. A replacement instance starts with neither: the boot
log resets, and there's no previous binary to roll back to until the first
deploy onto the new instance completes.

## What is not proven by this config alone

`tofu validate` (run as part of verification below) proves syntax and type
correctness. It does not prove the config works. Two things worth being
explicit about:

- **It does not evaluate variable `validation` blocks.** This was confirmed
  empirically during this branch's development:
  `tofu validate -var 'idle_timeout=00s'` returns `Success!` even though
  `00s` is a value that variable's own `validation` block is written to
  reject. The validation blocks in `variables.tf` are real and do run at
  plan/apply time against real values — they just aren't exercised by
  `validate` alone, so a clean `validate` says nothing about whether they're
  correct.
- **Nothing in this repository has been applied.** Whether `user_data.sh`
  actually completes end-to-end on a real instance, whether AWS accepts the
  GitHub OIDC trust policy (`data.aws_iam_openid_connect_provider.github` in
  `iam.tf`) the way it's written, and whether `ssm send-command` actually
  reaches the instance and returns a result — all three are first-apply
  findings, not settled facts. Treat the first real `tofu apply` as the
  point where these get tested, and watch the three admin commands below
  when it happens.

One more limit, on the shell scripts specifically: `shellcheck` was
unavailable on the machine this branch was developed on (not installed, and
installing it wasn't this branch's call to make) and could not be run
against `user_data.sh` or `files/update-fetzer.sh`. Both scripts were
instead reviewed by hand, more than once, specifically for the classes of
bug `shellcheck` catches (unquoted expansions, unintended word-splitting,
`set -e` interactions). That is real but weaker evidence than the tool
itself. Run `shellcheck` against both files if it's ever installed on a
machine working on this repo.

## Administration: SSM Session Manager, not SSH

There is no SSH admin path onto this instance, by design. `sshd` is disabled
and *masked* (not just disabled — masking stops a package update from
re-enabling it and fighting the TUI for port 22) in `user_data.sh` step 3.
Port `<ssh_port>` belongs entirely to the TUI server. The only way onto the
box is AWS Systems Manager Session Manager, which the instance role
(`aws_iam_role.instance` in `iam.tf`) is granted via the
`AmazonSSMManagedInstanceCore` managed policy, and which connects
outbound-only from the instance — no inbound port needed for admin access at
all.

```bash
aws ssm start-session --target "$(tofu output -raw instance_id)" --region "$(tofu output -raw region)"
```

Once connected, the first two things to check when the service is down:

```bash
journalctl -u fetzer          # service logs: crashes, restarts, why it's not active
cat /var/log/fetzer-bootstrap.log   # the full boot script transcript (see caveat below)
```

`journalctl -u fetzer` matters more than it might look: `aws ssm
get-command-invocation` — what the deploy workflow polls to learn whether a
deploy succeeded — truncates `StandardOutputContent` and
`StandardErrorContent` at 2500 characters each. A chatty failure from
`update-fetzer.sh` (a full `systemctl status` dump, for instance) can be cut
off mid-message in the GitHub Actions log. `journalctl -u fetzer` over a
Session Manager connection is the complete, untruncated record; treat the
workflow log as a summary and the journal as the source of truth when they
seem to disagree.

## Rollback

`builds/<commit-sha>` objects in the artifact bucket are kept indefinitely —
the S3 lifecycle rule only expires *noncurrent* versions, and each
`builds/<sha>` key is the current (only) version of a key distinct from
`current`, so it's never touched. That's what makes rollback to any past
commit possible at any time, not just recently.

Rolling back is deliberately **not** a workflow action. The deploy role
(`aws_iam_role.github_deploy` in `iam.tf`) holds `s3:PutObject` but not
`s3:GetObject` — it can publish a new build, but it cannot read `builds/<sha>`
back out to copy it over `current`. That's least privilege doing its job:
rollback requires broader read access than routine deploys need, so it stays
an admin action, run locally with your own AWS credentials, never something
the workflow can do on its own. Concretely:

```bash
BUCKET=$(tofu output -raw artifact_bucket)
aws s3 cp "s3://$BUCKET/builds/<known-good-sha>" "s3://$BUCKET/current"
aws ssm send-command \
  --instance-ids "$(tofu output -raw instance_id)" \
  --document-name AWS-RunShellScript \
  --parameters 'commands=["/usr/local/bin/update-fetzer.sh"]' \
  --region "$(tofu output -raw region)"
```

Also worth being explicit about: **re-running the GitHub Actions workflow
does not roll back anything.** `workflow_dispatch` rebuilds whatever is
currently at the tip of `main` and overwrites `current` with that build — the
same thing a normal push does. If `main` itself is broken, re-running the
workflow just redeploys the break. Rolling back to a prior commit requires
either reverting `main` and letting the workflow redeploy, or the manual `s3
cp` + `ssm send-command` above.

## Troubleshooting: deploys fail with an opaque AccessDenied

If the deploy workflow's SSM step starts failing with `AccessDenied` on
`ssm:SendCommand`, and nothing about the workflow or the IAM policy has
changed, the most likely cause is an out-of-band instance replacement: the
deploy role's policy pins `ssm:SendCommand` to one specific instance ARN
(`aws_instance.fetzer.arn`, in `iam.tf`) — not a wildcard, deliberately, so
this role can never command any other instance in the account. If the
instance has been replaced (an AMI refresh via `tofu apply`, a manual
`terminate-instances`, anything) since the deploy role was last applied, its
ARN is stale: the policy still names the *old* instance ARN, and the *new*
instance's ARN doesn't match it.

The error itself gives no hint of this — it just reads as "the workflow is
not allowed to do this," which looks like a workflow or credentials problem,
not an instance-identity problem. The fix is two steps, in order:

1. `tofu apply` — this re-reads `aws_instance.fetzer.arn` and updates the
   deploy role's policy to point at the current instance.
2. Update the `INSTANCE_ID` repository variable with the new value from
   `tofu output -raw instance_id`.

Both steps are required. Running only the first fixes the IAM policy but
leaves the workflow sending commands to the old (now-stale) instance ID from
the `INSTANCE_ID` variable; running only the second points the workflow at
the right instance ID but the IAM policy still denies it.

## The Elastic IP is the only durable public identity here

`aws_eip.fetzer` is the one thing in this config that, once released, cannot
be gotten back — a new Elastic IP allocation gets a different address, full
stop. Everything else (the instance, the artifact bucket's contents, the
host key parameter, even the S3 state bucket via bootstrap) can be
recreated or restored. `tofu destroy` releases the Elastic IP permanently.
If this address is ever pointed to from DNS (see the DNS section above),
losing it means updating DNS again, not just re-running `tofu apply`.

## Verification performed

Confirmed for this task: `tofu fmt -check -diff` and `tofu validate` clean
on both `terraform/` and `terraform/bootstrap/`; `actionlint` clean on
`.github/workflows/deploy-ssh.yml`; `bash -n` clean on both
`user_data.sh` and `files/update-fetzer.sh`; the real AWS account ID does
not appear in any file tracked by git. None of this touched AWS — no
`tofu plan` or `tofu apply` ran against the real backend, and no AWS API
call besides read-only `sts get-caller-identity` (used only to confirm the
account ID's absence from the tree, never written anywhere) was made.
