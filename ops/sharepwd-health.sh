#!/usr/bin/env bash
# SharePwd Daily Health Report
# Robust monitoring — guaranteed delivery via mail + optional Rocket.Chat webhook.
#
# Runs daily via cron at 07:00 UTC (see ops/crontab.example).
#
# Design principles:
#   - NEVER abort on partial failure (no `set -e` / `pipefail`).
#     The whole point of monitoring is to detect failure; aborting on first
#     error defeats the purpose. Each check is isolated with timeouts.
#   - HTTP check FIRST (most critical signal).
#   - Always reach the notification step, no matter what.
#   - Two redundant channels: SMTP + Rocket.Chat webhook.
#
# Configuration (override via /etc/sharepwd-monitoring.env, mode 0600):
#   RECIPIENT=ops@example.com
#   SERVER_IP=                          # default: hostname -I
#   EXPECTED_CONTAINERS=7
#   DOCKER_COMPOSE_PROJECT=deploy
#   BACKEND_CONTAINER=                  # default: ${DOCKER_COMPOSE_PROJECT}-backend-1
#   SSL_DOMAIN=sharepwd.io
#   HTTP_URL=                           # default: https://${SSL_DOMAIN}/
#   SENDER_FROM=                        # default: SharePwd Monitor <noreply@${SSL_DOMAIN}>
#   ROCKET_WEBHOOK_URL=                 # leave empty to disable webhook
#
# See ops/sharepwd-monitoring.env.example for a documented template.

set -u

# ── Config (override via /etc/sharepwd-monitoring.env) ────────────
if [[ -f /etc/sharepwd-monitoring.env ]]; then
    # shellcheck disable=SC1091
    source /etc/sharepwd-monitoring.env
fi

RECIPIENT="${RECIPIENT:-root@localhost}"
SERVER_IP="${SERVER_IP:-$(hostname -I 2>/dev/null | awk '{print $1}')}"
EXPECTED_CONTAINERS="${EXPECTED_CONTAINERS:-7}"
DOCKER_COMPOSE_PROJECT="${DOCKER_COMPOSE_PROJECT:-deploy}"
BACKEND_CONTAINER="${BACKEND_CONTAINER:-${DOCKER_COMPOSE_PROJECT}-backend-1}"
SSL_DOMAIN="${SSL_DOMAIN:-sharepwd.io}"
HTTP_URL="${HTTP_URL:-https://${SSL_DOMAIN}/}"
SENDER_FROM="${SENDER_FROM:-SharePwd Monitor <noreply@${SSL_DOMAIN}>}"
ROCKET_WEBHOOK_URL="${ROCKET_WEBHOOK_URL:-}"

DATE_STR=$(date -u '+%Y-%m-%d')
TIME_STR=$(date -u '+%Y-%m-%d %H:%M UTC')

# ── Status tracking ───────────────────────────────────────────────
WORST_STATUS="OK"
REPORT=""
add_line() { REPORT="${REPORT}$1"$'\n'; }

set_status() {
    case "$1" in
        CRIT) WORST_STATUS="CRIT" ;;
        WARN) [[ "$WORST_STATUS" != "CRIT" ]] && WORST_STATUS="WARN" ;;
    esac
}

# ── Build report ─────────────────────────────────────────────────

add_line "========================================"
add_line "  SharePwd Daily Health Report"
add_line "  Server: ${SERVER_IP}"
add_line "  Date:   ${TIME_STR}"
add_line "========================================"
add_line ""

# === 1) HTTP check (first — most critical) ===
HTTP_RESULT=$(curl -sS -o /dev/null -w "%{http_code} %{time_total}" --max-time 10 "${HTTP_URL}" 2>/dev/null || echo "000 timeout")
HTTP_CODE=$(echo "${HTTP_RESULT}" | awk '{print $1}')
HTTP_TIME=$(echo "${HTTP_RESULT}" | awk '{print $2}')
case "${HTTP_CODE}" in
    200|301|302|307|308)
        add_line "[OK]   HTTP ${HTTP_URL} → ${HTTP_CODE} (${HTTP_TIME}s)"
        ;;
    *)
        add_line "[CRIT] HTTP ${HTTP_URL} → ${HTTP_CODE} (${HTTP_TIME}s)  *** SITE DOWN ***"
        set_status "CRIT"
        ;;
esac

# === 2) System metrics ===
NUM_CORES=$(nproc 2>/dev/null || echo 1)
UPTIME_STR=$(uptime -p 2>/dev/null | sed 's/^up //' || echo "?")
LOAD_AVG=$(awk '{print $1, $2, $3}' /proc/loadavg 2>/dev/null || echo "? ? ?")
LOAD_1=$(awk '{print $1}' /proc/loadavg 2>/dev/null || echo 0)

if (( $(awk "BEGIN{print (${LOAD_1} > ${NUM_CORES}) ? 1 : 0}") )); then
    add_line "[WARN] Uptime: ${UPTIME_STR}, load: ${LOAD_AVG} (>${NUM_CORES} cores)"
    set_status "WARN"
else
    add_line "[OK]   Uptime: ${UPTIME_STR}, load: ${LOAD_AVG}"
fi

# CPU (idle from top) — guard against empty result
CPU_IDLE=$(top -bn2 -d1 2>/dev/null | grep '^%Cpu' | tail -1 | awk '{print $8}')
if [[ -n "${CPU_IDLE}" ]]; then
    CPU_USAGE=$(awk "BEGIN{printf \"%.0f\", 100 - ${CPU_IDLE}}")
    if (( CPU_USAGE > 95 )); then
        add_line "[CRIT] CPU: ${CPU_USAGE}%"; set_status "CRIT"
    elif (( CPU_USAGE > 80 )); then
        add_line "[WARN] CPU: ${CPU_USAGE}%"; set_status "WARN"
    else
        add_line "[OK]   CPU: ${CPU_USAGE}%"
    fi
else
    add_line "[WARN] CPU: unable to measure"; set_status "WARN"
fi

# RAM
RAM_LINE=$(free -h 2>/dev/null | awk '/^Mem:/{print $3" / "$2}')
RAM_PCT=$(free 2>/dev/null | awk '/^Mem:/{ if($2>0) printf "%.0f", $3/$2*100; else print 0 }')
if [[ -n "${RAM_PCT}" ]]; then
    if (( RAM_PCT > 95 )); then
        add_line "[CRIT] RAM: ${RAM_LINE} (${RAM_PCT}%)"; set_status "CRIT"
    elif (( RAM_PCT > 80 )); then
        add_line "[WARN] RAM: ${RAM_LINE} (${RAM_PCT}%)"; set_status "WARN"
    else
        add_line "[OK]   RAM: ${RAM_LINE} (${RAM_PCT}%)"
    fi
fi

# Swap
SWAP_LINE=$(free -h 2>/dev/null | awk '/^Swap:/{print $3" / "$2}')
SWAP_PCT=$(free 2>/dev/null | awk '/^Swap:/{ if($2>0) printf "%.0f", $3/$2*100; else print 0 }')
if [[ -n "${SWAP_PCT}" ]] && (( SWAP_PCT > 50 )); then
    add_line "[WARN] Swap: ${SWAP_LINE} (${SWAP_PCT}%)"; set_status "WARN"
else
    add_line "[OK]   Swap: ${SWAP_LINE} (${SWAP_PCT:-0}%)"
fi

# Disk /
DISK_LINE=$(df -h / 2>/dev/null | awk 'NR==2{print $3" / "$2}')
DISK_PCT=$(df / 2>/dev/null | awk 'NR==2{gsub(/%/,""); print $5}')
if [[ -n "${DISK_PCT}" ]]; then
    if (( DISK_PCT > 90 )); then
        add_line "[CRIT] Disk /: ${DISK_LINE} (${DISK_PCT}%)"; set_status "CRIT"
    elif (( DISK_PCT > 80 )); then
        add_line "[WARN] Disk /: ${DISK_LINE} (${DISK_PCT}%)"; set_status "WARN"
    else
        add_line "[OK]   Disk /: ${DISK_LINE} (${DISK_PCT}%)"
    fi
fi

# === 3) Docker containers ===
CONTAINER_LIST=$(docker ps --format '{{.Names}}\t{{.Status}}' --filter "label=com.docker.compose.project=${DOCKER_COMPOSE_PROJECT}" 2>/dev/null || echo "")
RUNNING_COUNT=$(echo "${CONTAINER_LIST}" | grep -c "Up" || true)
RUNNING_COUNT=${RUNNING_COUNT//[^0-9]/}
RUNNING_COUNT=${RUNNING_COUNT:-0}

if [[ -z "${CONTAINER_LIST}" ]]; then
    add_line "[CRIT] Docker: unable to list containers"; set_status "CRIT"
elif (( RUNNING_COUNT < EXPECTED_CONTAINERS )); then
    add_line "[CRIT] Docker: ${RUNNING_COUNT}/${EXPECTED_CONTAINERS} containers running"
    set_status "CRIT"
else
    add_line "[OK]   Docker: ${RUNNING_COUNT}/${EXPECTED_CONTAINERS} containers running"
fi

# === 4) SSL certificate ===
SSL_EXPIRY=$(echo | timeout 8 openssl s_client -servername "${SSL_DOMAIN}" -connect "${SSL_DOMAIN}:443" 2>/dev/null | openssl x509 -noout -enddate 2>/dev/null | cut -d= -f2)
if [[ -n "${SSL_EXPIRY}" ]]; then
    SSL_EPOCH=$(date -d "${SSL_EXPIRY}" +%s 2>/dev/null || echo 0)
    NOW_EPOCH=$(date +%s)
    SSL_DAYS=$(( (SSL_EPOCH - NOW_EPOCH) / 86400 ))
    if (( SSL_DAYS < 7 )); then
        add_line "[CRIT] SSL: ${SSL_DAYS} days remaining"; set_status "CRIT"
    elif (( SSL_DAYS < 30 )); then
        add_line "[WARN] SSL: ${SSL_DAYS} days remaining"; set_status "WARN"
    else
        add_line "[OK]   SSL: ${SSL_DAYS} days remaining"
    fi
else
    add_line "[WARN] SSL: cannot reach ${SSL_DOMAIN}:443 to read cert"
    set_status "WARN"
fi

# === 5) UFW ===
UFW_STATUS=$(sudo -n ufw status 2>/dev/null || echo "")
if echo "${UFW_STATUS}" | grep -q "Status: active"; then
    UFW_RULES=$(echo "${UFW_STATUS}" | grep -cE '^[0-9]')
    add_line "[OK]   UFW: active (${UFW_RULES} rules)"
else
    add_line "[CRIT] UFW: inactive"; set_status "CRIT"
fi

# === 6) Fail2ban ===
F2B_BANNED_IPS=""
F2B_TOTAL_BANNED=0
if F2B_STATUS=$(sudo -n fail2ban-client status 2>/dev/null); then
    F2B_JAILS=$(echo "${F2B_STATUS}" | grep "Jail list" | sed 's/.*:\s*//' | tr ',' '\n' | sed 's/^ //')
    for jail in ${F2B_JAILS}; do
        JAIL_STATUS=$(sudo -n fail2ban-client status "${jail}" 2>/dev/null || echo "")
        JAIL_BANNED=$(echo "${JAIL_STATUS}" | grep "Currently banned" | awk '{print $NF}')
        JAIL_IPS=$(echo "${JAIL_STATUS}" | grep "Banned IP list" | sed 's/.*:\s*//')
        F2B_TOTAL_BANNED=$((F2B_TOTAL_BANNED + ${JAIL_BANNED:-0}))
        [[ -n "${JAIL_IPS}" ]] && F2B_BANNED_IPS="${F2B_BANNED_IPS}${JAIL_IPS} "
    done
    add_line "[OK]   Fail2ban: active, ${F2B_TOTAL_BANNED} IPs banned"
else
    add_line "[CRIT] Fail2ban: inactive or not installed"; set_status "CRIT"
fi

# === 7) SSH failed attempts (24h) ===
SSH_FAILED=$(sudo -n journalctl -u ssh --since "24 hours ago" 2>/dev/null | grep -c "Failed password" || true)
SSH_FAILED=${SSH_FAILED//[^0-9]/}
add_line "[INFO] SSH: ${SSH_FAILED:-0} failed password attempts (24h)"

# === 8) Security updates ===
SEC_UPDATES=$(apt list --upgradable 2>/dev/null | grep -ic security || true)
SEC_UPDATES=${SEC_UPDATES//[^0-9]/}
if (( ${SEC_UPDATES:-0} > 0 )); then
    add_line "[WARN] Security updates: ${SEC_UPDATES} pending"; set_status "WARN"
else
    add_line "[OK]   Security updates: none pending"
fi

# === 9) Detail sections ===
add_line ""
add_line "--- Docker containers ---"
if [[ -n "${CONTAINER_LIST}" ]]; then
    while IFS= read -r line; do add_line "${line}"; done <<< "${CONTAINER_LIST}"
else
    add_line "(unable to list containers)"
fi

add_line ""
add_line "--- Backend errors (last 24h) ---"
BACKEND_ERRORS=$(docker logs "${BACKEND_CONTAINER}" --since 24h 2>&1 | grep -iE '(error|panic|fatal)' | grep -v "level=info" | tail -20 || true)
if [[ -n "${BACKEND_ERRORS}" ]]; then
    add_line "${BACKEND_ERRORS}"
    set_status "WARN"
else
    add_line "(none)"
fi

add_line ""
add_line "--- Fail2ban banned IPs ---"
if [[ -n "${F2B_BANNED_IPS}" ]]; then
    for ip in ${F2B_BANNED_IPS}; do add_line "${ip}"; done
else
    add_line "(none)"
fi

add_line ""
add_line "--- Recent logins (last 24h) ---"
RECENT_LOGINS=$(last --since yesterday 2>/dev/null | head -20 || true)
[[ -n "${RECENT_LOGINS}" ]] && add_line "${RECENT_LOGINS}" || add_line "(none)"

add_line ""
add_line "--- Last certbot renewal ---"
add_line "$(sudo -n tail -5 /var/log/letsencrypt/letsencrypt.log 2>/dev/null || echo "(no log found)")"

# ── Notifications ─────────────────────────────────────────────────

SUBJECT="[${WORST_STATUS}] SharePwd Daily Report -- ${DATE_STR}"

# 1) SMTP — fail fast (no 5min sleep retry)
EMAIL_SENT=false
{
    printf 'From: %s\n' "${SENDER_FROM}"
    printf 'To: %s\n' "${RECIPIENT}"
    printf 'Subject: %s\n' "${SUBJECT}"
    printf 'Date: %s\n' "$(date -uR)"
    printf 'Message-ID: <%s.%s@%s>\n' "$(date -u +%Y%m%d%H%M%S)" "$$" "${SSL_DOMAIN}"
    printf 'MIME-Version: 1.0\n'
    printf 'Content-Type: text/plain; charset=UTF-8\n'
    printf 'Content-Transfer-Encoding: 8bit\n'
    printf '\n'
    printf '%s' "${REPORT}"
} | timeout 30 sendmail "${RECIPIENT}" >/dev/null 2>&1 && EMAIL_SENT=true

# 2) Rocket.Chat webhook (optional)
ROCKET_SENT=false
if [[ -n "${ROCKET_WEBHOOK_URL}" ]]; then
    PAYLOAD=$(WORST_STATUS="${WORST_STATUS}" DATE_STR="${DATE_STR}" REPORT_VAR="${REPORT}" python3 <<'PYEOF'
import json, os
status = os.environ['WORST_STATUS']
date = os.environ['DATE_STR']
report = os.environ['REPORT_VAR']
emoji = {'CRIT': ':rotating_light:', 'WARN': ':warning:', 'OK': ':white_check_mark:'}.get(status, ':grey_question:')
color = {'CRIT': '#FF0000', 'WARN': '#FFA500', 'OK': '#36A64F'}.get(status, '#808080')
print(json.dumps({
    "text": f"{emoji} *[{status}] SharePwd Daily Report — {date}*",
    "attachments": [{"color": color, "text": "```\n" + report + "\n```"}]
}))
PYEOF
    )
    HTTP=$(curl -sS -o /dev/null -w "%{http_code}" --max-time 10 \
        -X POST -H "Content-Type: application/json" -d "${PAYLOAD}" \
        "${ROCKET_WEBHOOK_URL}" 2>/dev/null || echo "000")
    [[ "${HTTP}" == "200" ]] && ROCKET_SENT=true
fi

# 3) Log result
NOW=$(date -u '+%Y-%m-%d %H:%M:%S UTC')
echo "[${NOW}] status=${WORST_STATUS} email=${EMAIL_SENT} rocket=${ROCKET_SENT} recipient=${RECIPIENT}"

# Exit non-zero if status non-OK and BOTH channels failed (cron will mail root)
if [[ "${WORST_STATUS}" != "OK" && "${EMAIL_SENT}" != true && "${ROCKET_SENT}" != true ]]; then
    echo "[${NOW}] FATAL: status=${WORST_STATUS} but no notification channel succeeded" >&2
    exit 1
fi

exit 0
