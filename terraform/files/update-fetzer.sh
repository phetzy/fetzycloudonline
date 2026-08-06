#!/usr/bin/env bash
# Deploy-time update script for the SSH TUI binary.
#
# This file is committed verbatim (no templatefile() placeholders — the
# account ID and bucket name must never appear in a committed file) and
# copied onto the instance as-is by terraform/user_data.sh. It runs in two
# situations:
#   1. Once at first boot, invoked directly by user_data.sh, where "no
#      binary has been published yet" is expected and must be tolerated.
#   2. On every deploy, invoked by GitHub Actions via
#      `aws ssm send-command --document-name AWS-RunShellScript`, where a
#      real failure must be reported with a non-zero exit so the workflow
#      can see it.
#
# In both cases it runs as root (the boot script runs as root; SSM
# RunShellScript runs as root by default), which is required to write
# /usr/local/bin and restart a system unit. The service itself still runs
# as the unprivileged `fetzer` user — only the deploy mechanism is root.
#
# Bucket name and region are not passed as arguments or baked into this
# file; they come from /etc/fetzer/env, written by user_data.sh from
# Terraform-resolved values at boot time.
set -euo pipefail

env_file="/etc/fetzer/env"
install_path="/usr/local/bin/fetzer"

# The artifact bucket holds every build under builds/<commit-sha> (kept
# indefinitely, one object per deploy) plus a single "current" object that
# the deploy workflow overwrites, pointer-style, on every successful build.
# This script only ever reads "current" — it never touches the builds/
# prefix, so those SHA-keyed objects exist purely so a human (or a future
# rollback command) can `aws s3 cp s3://<bucket>/builds/<sha> s3://<bucket>/current`
# to point "current" at an older build and re-run the deploy.
s3_key="current"
unit_name="fetzer"

if [ ! -r "$env_file" ]; then
  echo "update-fetzer: $env_file not found or not readable; cannot determine region/bucket" >&2
  exit 1
fi
# shellcheck disable=SC1090
. "$env_file"

if [ -z "${AWS_REGION:-}" ] || [ -z "${ARTIFACT_BUCKET:-}" ]; then
  echo "update-fetzer: AWS_REGION and ARTIFACT_BUCKET must both be set in $env_file" >&2
  exit 1
fi

# Download to a temp file in the same directory as the final destination so
# the later `mv` is an atomic same-filesystem rename, not a cross-device
# copy that could leave a partially-written binary in place if interrupted.
tmp_bin=$(mktemp /usr/local/bin/.fetzer.XXXXXX)
cleanup() {
  rm -f "$tmp_bin"
}
trap cleanup EXIT

# The bucket name is not echoed here, even though it's masked twice over
# already (tofu output sensitive = true, GitHub secret masking) — both of
# those depend on the owner having pasted it in correctly, and dropping it
# from this one line removes that dependency at zero cost. The object key
# alone is enough to know what's happening.
echo "update-fetzer: pulling s3://<bucket>/${s3_key} (region $AWS_REGION)"
if ! aws s3 cp "s3://${ARTIFACT_BUCKET}/${s3_key}" "$tmp_bin" --region "$AWS_REGION"; then
  echo "update-fetzer: failed to pull binary from S3 (no build published yet, or a transient error); leaving the running binary alone" >&2
  exit 1
fi

chmod 755 "$tmp_bin"

# Verify the download is a working binary before it ever touches the
# install path. -h is handled by Go's flag package as ErrHelp, which exits
# 0 after printing usage — a real corrupt/incompatible binary will instead
# fail to exec, segfault, or exit non-zero.
if ! "$tmp_bin" -h >/dev/null 2>&1; then
  echo "update-fetzer: downloaded binary failed to run (fetzer -h did not exit cleanly); leaving the running binary alone" >&2
  exit 1
fi

# Keep a backup so a build that starts and then crashes can be rolled back,
# not just a build that fails the -h check above. None on the first-ever deploy.
[ -e "$install_path" ] && cp -a "$install_path" "$install_path.prev"

# Swap it in. This mv is the point of no return: from here a failure is
# reported loudly rather than silently leaving the old binary in place,
# because the workflow needs a non-zero exit to know the deploy is bad.
trap - EXIT
if ! mv -f "$tmp_bin" "$install_path"; then
  echo "update-fetzer: failed to install new binary at $install_path" >&2
  rm -f "$tmp_bin"
  exit 1
fi
chown root:root "$install_path"
chmod 755 "$install_path"

# Clears any prior start-limit-hit state so a stale failure from a previous
# bad deploy can't mask this restart's real result.
systemctl reset-failed "$unit_name" || true

echo "update-fetzer: restarting $unit_name"
systemctl restart "$unit_name" || true

# Give the unit a moment to either settle into "active" or fail fast (it has
# RestartSec=2s and Restart=on-failure, so a crash-on-start shows up here
# rather than looking transiently healthy).
sleep 3
if ! systemctl is-active --quiet "$unit_name"; then
  echo "update-fetzer: $unit_name is not active after restart" >&2
  systemctl status --no-pager "$unit_name" >&2 || true
  # Roll back so a bad build doesn't take the service down until fixed.
  if [ -e "$install_path.prev" ]; then
    echo "update-fetzer: rolling back to previous binary" >&2
    mv -f "$install_path.prev" "$install_path"
    systemctl reset-failed "$unit_name" || true
    systemctl restart "$unit_name" || true
  fi
  exit 1
fi

rm -f "$install_path.prev"
echo "update-fetzer: deploy complete, $unit_name is active"
