#!/usr/bin/env bash
# SharePwd ops — idempotent host-side installer.
#
# Installs:
#   /usr/local/sbin/sharepwd-health.sh       (mode 0755)
#   /etc/logrotate.d/sharepwd                (mode 0644)
#   /etc/sharepwd-monitoring.env             (mode 0600, only if missing)
#   root crontab entries                     (deduplicated)
#
# Does NOT touch /etc/msmtprc — see ops/msmtprc.example, install manually.
#
# Usage (from repo root):
#   sudo ./ops/deploy.sh

set -euo pipefail

if [[ ${EUID} -ne 0 ]]; then
    echo "ERROR: must be run as root (sudo ./ops/deploy.sh)" >&2
    exit 1
fi

OPS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "==> Installing /usr/local/sbin/sharepwd-health.sh"
install -m 0755 -o root -g root "${OPS_DIR}/sharepwd-health.sh" /usr/local/sbin/sharepwd-health.sh

echo "==> Installing /etc/logrotate.d/sharepwd"
install -m 0644 -o root -g root "${OPS_DIR}/logrotate.d/sharepwd" /etc/logrotate.d/sharepwd

if [[ ! -f /etc/sharepwd-monitoring.env ]]; then
    echo "==> Installing /etc/sharepwd-monitoring.env (from example, edit to enable Rocket.Chat)"
    install -m 0600 -o root -g root "${OPS_DIR}/sharepwd-monitoring.env.example" /etc/sharepwd-monitoring.env
else
    echo "==> /etc/sharepwd-monitoring.env already present, leaving untouched"
fi

if [[ ! -f /etc/msmtprc ]]; then
    echo
    echo "WARNING: /etc/msmtprc is missing — mail delivery will fail."
    echo "         Adapt ${OPS_DIR}/msmtprc.example to your relay and install it"
    echo "         as root with mode 0600. See ops/README.md for details."
    echo
fi

echo "==> Updating root crontab (deduplicated)"
TMP=$(mktemp)
trap 'rm -f "${TMP}"' EXIT
{
    crontab -l 2>/dev/null || true
    grep -vE '^\s*(#|$)' "${OPS_DIR}/crontab.example"
} | awk '!seen[$0]++' > "${TMP}"
crontab "${TMP}"

echo
echo "Done. Verify with:"
echo "  crontab -l"
echo "  /usr/local/sbin/sharepwd-health.sh   # manual run"
