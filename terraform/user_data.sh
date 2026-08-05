#!/usr/bin/env bash
# Boot script for the SSH TUI instance, rendered by templatefile() in
# terraform/main.tf (Task 7) and passed to the instance as user_data.
#
# This runs once per instance via cloud-init, but a reboot of a *replacement*
# instance (new instance ID, same host-key SSM parameter) runs it again from
# scratch, so every step below is written to be safe to repeat: package
# installs are idempotent by nature, the system user is created only if
# missing, sshd is disabled/masked (itself idempotent), and the host key is
# fetched from SSM rather than generated blindly so a replacement instance
# picks up the existing key instead of minting a new one.
#
# Template variables (substituted by templatefile() before bash ever sees
# this file — anything written as $${...} below is a literal bash variable
# reference that survives that substitution unchanged):
#   region              - AWS region, e.g. "us-west-2"
#   artifact_bucket     - S3 bucket the binary is published to
#   host_key_parameter  - SSM parameter name holding the SSH host key
#   service_unit        - fully rendered contents of files/fetzer.service
#                         (already run through its own templatefile() call,
#                         so it is plain text by the time it reaches here)
#   update_script       - contents of files/update-fetzer.sh, verbatim
set -euo pipefail

# Every line of this script's output goes to both stdout (captured in the
# cloud-init log) and a dedicated log file, because SSM Session Manager is
# the only way onto this box and that log is the only diagnostic trail if
# something below fails. No secret material is ever echoed: the host key is
# written straight to disk (step 4) and never passed through a command whose
# output is logged.
exec > >(tee -a /var/log/fetzer-bootstrap.log) 2>&1

echo "=== fetzer bootstrap starting at $(date -u --iso-8601=seconds) ==="

region="${region}"
artifact_bucket="${artifact_bucket}"
host_key_parameter="${host_key_parameter}"

service_user="fetzer"
service_group="fetzer"
state_dir="/var/lib/fetzer"
host_key_path="$${state_dir}/host_key"
env_file="/etc/fetzer/env"
update_script_path="/usr/local/bin/update-fetzer.sh"
unit_path="/etc/systemd/system/fetzer.service"

# ---------------------------------------------------------------------------
# 1. Make sure the AWS CLI is present. AL2023 ships it, but do not assume.
# ---------------------------------------------------------------------------
if ! command -v aws >/dev/null 2>&1; then
  echo "aws CLI not found, installing awscli-2"
  dnf -y install awscli-2
fi

# ---------------------------------------------------------------------------
# 2. Create the service system user: no login shell, no home-directory login.
# ---------------------------------------------------------------------------
if ! id "$service_user" >/dev/null 2>&1; then
  echo "creating system user $service_user"
  useradd --system --no-create-home --shell /usr/sbin/nologin "$service_user"
fi

# ---------------------------------------------------------------------------
# 3. Disable and mask sshd. The TUI owns port 22; there is no admin SSH path
#    by design (SSM Session Manager is the only admin path). Masking, not
#    just disabling, matters: a package update re-enabling sshd would put
#    two things fighting over port 22, one of them an admin login.
# ---------------------------------------------------------------------------
systemctl disable --now sshd || true
systemctl mask sshd

# ---------------------------------------------------------------------------
# 4. Fetch or generate the SSH host key.
#
#    The SSM parameter always exists by the time this runs (Task 3 created
#    it as a placeholder before the instance ever boots), so this must never
#    treat "parameter missing" as the normal case. It fetches the current
#    value and checks whether it is still the placeholder string. Only the
#    placeholder (or a failed fetch) triggers key generation; any other
#    value is a real key from an earlier boot and is written out unchanged.
#    Getting this backwards regenerates the host key on every boot, which
#    breaks host-key pinning for every returning visitor.
# ---------------------------------------------------------------------------
mkdir -p "$state_dir"
chown "$service_user:$service_group" "$state_dir"
chmod 700 "$state_dir"

placeholder="placeholder-replaced-on-first-boot"
fetch_err_file=$(mktemp)
fetched_value=""
fetch_status=0
fetched_value=$(aws ssm get-parameter \
  --region "$region" \
  --name "$host_key_parameter" \
  --with-decryption \
  --query 'Parameter.Value' \
  --output text 2>"$fetch_err_file") || fetch_status=$?

# Distinguish "the parameter is genuinely missing" from every other failure.
# The parameter is created before the instance ever boots (Task 3), so
# ParameterNotFound here means something is badly wrong with the deploy, not
# the expected first-boot state — but it is the one failure mode that still
# means "no real key exists yet". Any other failure (throttling, a network
# blip reaching the SSM endpoint, an IAM problem) must NOT fall through to
# key generation: doing so would silently overwrite a real host key on a
# transient error and rotate the fingerprint under every returning visitor.
# Those failures abort the boot instead.
if [ "$fetch_status" -ne 0 ] && ! grep -q "ParameterNotFound" "$fetch_err_file"; then
  echo "fatal: aws ssm get-parameter failed for a reason other than ParameterNotFound:"
  cat "$fetch_err_file"
  rm -f "$fetch_err_file"
  exit 1
fi
rm -f "$fetch_err_file"

if [ "$fetch_status" -eq 0 ] && [ "$fetched_value" != "$placeholder" ] && [ -n "$fetched_value" ]; then
  echo "existing host key found in SSM, installing it unchanged"
  # printf, not echo, keeps the key out of anything that would inspect argv
  # of a program other than printf itself, and the umask keeps the file from
  # ever being briefly world-readable between creation and chmod.
  ( umask 077 && printf '%s\n' "$fetched_value" > "$host_key_path" )
else
  echo "no usable host key in SSM (placeholder or fetch failure), generating one"
  rm -f "$host_key_path" "$${host_key_path}.pub"
  ssh-keygen -t ed25519 -N '' -f "$host_key_path" >/dev/null
  rm -f "$${host_key_path}.pub"
  aws ssm put-parameter \
    --region "$region" \
    --name "$host_key_parameter" \
    --type SecureString \
    --value "file://$host_key_path" \
    --overwrite
fi

chown "$service_user:$service_group" "$host_key_path"
chmod 600 "$host_key_path"

# ---------------------------------------------------------------------------
# 5. Write the runtime env file update-fetzer.sh reads on every invocation
#    (both this initial run and later ones triggered by `ssm send-command`).
#    This is how the bucket name and region reach the script without either
#    being baked into the committed, unrendered files/update-fetzer.sh.
# ---------------------------------------------------------------------------
mkdir -p /etc/fetzer
cat > "$env_file" <<ENV_EOF
AWS_REGION=$region
ARTIFACT_BUCKET=$artifact_bucket
ENV_EOF
chmod 644 "$env_file"

# ---------------------------------------------------------------------------
# 6. Install the update script the deploy pipeline invokes via SSM.
# ---------------------------------------------------------------------------
cat > "$update_script_path" <<'UPDATE_SCRIPT_EOF'
${update_script}
UPDATE_SCRIPT_EOF
chown root:root "$update_script_path"
chmod 755 "$update_script_path"

# ---------------------------------------------------------------------------
# 7. Run it once to pull whatever binary already exists. The very first
#    apply happens before the first deploy, so "no binary published yet" is
#    expected and must not fail the boot.
# ---------------------------------------------------------------------------
echo "running initial update-fetzer.sh (missing binary is non-fatal here)"
"$update_script_path" || echo "initial pull did not complete (likely no binary published yet); continuing"

# ---------------------------------------------------------------------------
# 8. Install and start the unit.
# ---------------------------------------------------------------------------
cat > "$unit_path" <<'UNIT_EOF'
${service_unit}
UNIT_EOF
chown root:root "$unit_path"
chmod 644 "$unit_path"

systemctl daemon-reload

# Only start the unit now if a binary actually exists. On the very first
# apply, nothing has been deployed yet: ExecStart would point at a missing
# file, and enabling with --now would let systemd burn through its default
# start-limit-burst restarting a unit that can never succeed, landing it
# permanently in "failed" until something manually resets it. Enabling
# without starting means the unit is armed for every future boot, and the
# first real deploy's `systemctl restart fetzer` (in update-fetzer.sh) is
# what actually brings it up.
if [ -x /usr/local/bin/fetzer ]; then
  systemctl enable --now fetzer
else
  echo "no fetzer binary present yet; enabling the unit without starting it"
  systemctl enable fetzer
fi

echo "=== fetzer bootstrap finished at $(date -u --iso-8601=seconds) ==="
