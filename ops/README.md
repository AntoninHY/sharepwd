# ops/ — SharePwd Production Infrastructure

Server-side configuration files for running SharePwd in production. The application stack itself lives in `deploy/` (docker compose); this directory holds the host-level pieces: monitoring, log rotation, mail relay, and cron entries.

These files are environment-agnostic templates. All site-specific values (recipient email, server IP, webhook URLs, SMTP credentials) are set in `/etc/sharepwd-monitoring.env` and `/etc/msmtprc`, which are **not** part of this directory and must be created on each host.

## Files

| File | Target on host | Description |
|------|----------------|-------------|
| `sharepwd-health.sh` | `/usr/local/sbin/sharepwd-health.sh` | Daily health report (HTTP, containers, SSL, host metrics), emailed via msmtp + optional Rocket.Chat webhook |
| `sharepwd-monitoring.env.example` | `/etc/sharepwd-monitoring.env` | Documented template for the monitoring config sourced by the health script |
| `logrotate.d/sharepwd` | `/etc/logrotate.d/sharepwd` | Weekly rotation of health, certbot and msmtp logs |
| `msmtprc.example` | `/etc/msmtprc` | SMTP relay templates (Gmail, Mailgun, Postmark, local relay) |
| `crontab.example` | root crontab | Cron entries: certbot renew (03:00 UTC) + health check (07:00 UTC) |
| `deploy.sh` | — | Idempotent installer for everything above except `msmtprc` |

## Quick start

On a server where the SharePwd Docker stack is already running (`deploy/`), from the repo root:

```bash
sudo ./ops/deploy.sh
```

The script is idempotent — re-running it is safe. It installs `sharepwd-health.sh`, the logrotate config, and the cron entries; it copies `sharepwd-monitoring.env.example` to `/etc/sharepwd-monitoring.env` only if missing, so you can edit it without losing your changes on the next deploy.

The script does **not** touch `/etc/msmtprc`. The mail relay configuration is environment-specific (provider, credentials) and must be installed manually — see below.

## Configure the monitoring (`/etc/sharepwd-monitoring.env`)

Edit `/etc/sharepwd-monitoring.env` (the deploy script created it from the example). The file uses simple `KEY=VALUE` shell syntax. All keys are optional; defaults work for the canonical `sharepwd.io` deployment.

```sh
RECIPIENT=ops@example.com
SSL_DOMAIN=sharepwd.example.com
ROCKET_WEBHOOK_URL=https://chat.example.com/hooks/AAA/BBB
```

The full list of supported keys is documented in `sharepwd-monitoring.env.example`.

## Configure the mail relay (`/etc/msmtprc`)

The health script delivers via `sendmail` (provided by `bsd-mailx` + `msmtp-mta`). On Debian / Ubuntu:

```bash
sudo apt install msmtp msmtp-mta bsd-mailx
sudo cp ops/msmtprc.example /etc/msmtprc
sudo chmod 600 /etc/msmtprc
sudo chown root:root /etc/msmtprc
sudo $EDITOR /etc/msmtprc      # uncomment one provider, fill in credentials
```

`msmtprc.example` ships templates for Gmail, Mailgun, Postmark and a local IP-whitelisted relay. Test with:

```bash
echo "test" | sudo mail -s "smtp test" you@example.com
```

If the test fails with `554 Relay access denied`, your relay does not accept messages from this server's IP — either authenticate (`auth on`, `user`, `password`) or have your relay's operator whitelist the IP.

## Health check

`sharepwd-health.sh` runs daily at 07:00 UTC and reports on:

- HTTP reachability of the public URL
- Docker container count (expected: 7 — backend, frontend, nginx, postgres, redis, minio, umami)
- Backend container errors over the last 24h
- TLS certificate days remaining
- Host CPU / memory / swap / disk
- UFW status, fail2ban bans, SSH failed logins, pending security updates

Output is logged to `/var/log/sharepwd-health.log` and delivered via:

- **Email** to `RECIPIENT` (always attempted)
- **Rocket.Chat webhook** if `ROCKET_WEBHOOK_URL` is set

The script is intentionally written without `set -e` / `pipefail`: a monitoring script must always reach the notification step, even when individual checks fail. Each probe is wrapped in a timeout so a single hanging command cannot block the report.

To run a one-off check without waiting for cron:

```bash
sudo /usr/local/sbin/sharepwd-health.sh
```

## Cron

```cron
0 3 * * * certbot renew --webroot -w /var/lib/docker/volumes/deploy_certbot_webroot/_data --deploy-hook "docker exec deploy-nginx-1 nginx -s reload" >> /var/log/certbot-renew.log 2>&1
0 7 * * * /usr/local/sbin/sharepwd-health.sh >> /var/log/sharepwd-health.log 2>&1
```

The certbot `--deploy-hook` reloads nginx in-place after a successful renewal. The renewal config in `/etc/letsencrypt/renewal/<domain>.conf` should also have `renew_hook = docker restart deploy-nginx-1` (or `nginx -s reload`) as a fallback for renewals triggered by certbot's systemd timer rather than this cron line.

## Adapting to a non-default deployment

| If your... | Set in `/etc/sharepwd-monitoring.env` |
|-----------|----------------------------------------|
| compose project is not named `deploy` | `DOCKER_COMPOSE_PROJECT=yourproject` |
| domain is not `sharepwd.io` | `SSL_DOMAIN=share.example.com` |
| stack expects fewer or more containers | `EXPECTED_CONTAINERS=N` |
| public URL differs from `https://${SSL_DOMAIN}/` | `HTTP_URL=https://share.example.com/api/health` |
