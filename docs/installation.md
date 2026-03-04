# Installation & Deployment

Self-hosted deployment guide for SharePwd.

## Prerequisites

- **Docker Engine** 24+ with the **Compose plugin** (`docker compose`)
- A **domain name** with DNS A record pointing to your server
- **TLS certificates** — [Let's Encrypt](https://letsencrypt.org/) with Certbot recommended
- A Linux server (Ubuntu 22.04+, Debian 12+, or equivalent)

## 1. Clone the repository

```bash
git clone https://github.com/AntoninHY/sharepwd.git
cd sharepwd
```

## 2. Configure environment

```bash
cp deploy/.env.example deploy/.env
```

Open `deploy/.env` in your editor. You need to set every value marked below.

### Passwords and secrets

Generate a unique random value for each variable and paste it into `.env`:

```bash
# Run this once per variable — copy each output into .env
openssl rand -base64 32
```

| Variable in `.env` | What it protects |
|----------|---------|
| `POSTGRES_PASSWORD` | PostgreSQL database access |
| `MINIO_ROOT_PASSWORD` | MinIO (S3) storage access |
| `UMAMI_DB_PASSWORD` | Umami analytics database access |
| `UMAMI_APP_SECRET` | Umami session signing |

For the admin secret (used to bootstrap API keys), use a longer value:

```bash
openssl rand -base64 48
# Copy the output and paste it as the ADMIN_SECRET value in .env
```

| Variable in `.env` | What it protects |
|----------|---------|
| `ADMIN_SECRET` | API key creation via admin endpoint ([details](api-keys.md)) |

### URLs

Replace all URL variables in `.env` with your actual domain:

```ini
BASE_URL=https://yourdomain.tld
CORS_ORIGINS=https://yourdomain.tld
NEXT_PUBLIC_API_URL=https://yourdomain.tld
NEXT_PUBLIC_APP_URL=https://yourdomain.tld
```

## 3. TLS certificates

Obtain certificates with Certbot:

```bash
sudo apt install certbot
sudo certbot certonly --standalone -d yourdomain.tld
```

This places certificates at:

```
/etc/letsencrypt/live/yourdomain.tld/fullchain.pem
/etc/letsencrypt/live/yourdomain.tld/privkey.pem
```

If your certificates are stored elsewhere, update the volume mount in `deploy/compose.yaml` under the `nginx` service.

### Automatic renewal

Certbot installs a systemd timer for automatic renewal. Verify with:

```bash
sudo systemctl status certbot.timer
```

After renewal, reload Nginx:

```bash
docker compose -f deploy/compose.yaml exec nginx nginx -s reload
```

## 4. Deploy

```bash
cd deploy
docker compose up -d --build
```

This starts 6 containers: PostgreSQL, MinIO, Backend, Frontend, Umami, and Nginx.

## 5. Verify

```bash
# All 6 containers should be running
docker compose ps

# Health check
curl https://yourdomain.tld/v1/health
# → {"status":"ok"}
```

## 6. Create the first API key

After deployment, bootstrap your first API key using the admin secret:

```bash
curl -X POST https://yourdomain.tld/v1/admin/api-keys \
  -H "Content-Type: application/json" \
  -H "X-Admin-Secret: $ADMIN_SECRET" \
  -d '{"name": "initial-key"}'
```

Or with the CLI:

```bash
sharepwd admin keys create --name "initial-key" \
  --admin-secret "$ADMIN_SECRET" \
  --server https://yourdomain.tld
```

Save the returned API key — it is only shown once. See [API Key Management](api-keys.md) for details.

## Updating

Pull the latest changes and rebuild:

```bash
cd /path/to/sharepwd
git pull
cd deploy
docker compose up -d --build
```

Docker Compose recreates only the containers whose images changed. Persistent data (PostgreSQL, MinIO) is stored in Docker volumes and is preserved across updates.

## Troubleshooting

### Containers not starting

```bash
# Check logs for a specific service
docker compose logs backend
docker compose logs postgres

# Check all logs
docker compose logs
```

### Database connection errors

Ensure PostgreSQL is healthy before the backend starts:

```bash
docker compose ps postgres
# Should show "healthy"
```

If PostgreSQL is restarting, check that `POSTGRES_PASSWORD` is set in `.env` and matches across services.

### Certificate errors

- Verify certificates exist: `ls /etc/letsencrypt/live/yourdomain.tld/`
- Check Nginx can read them: `docker compose exec nginx nginx -t`
- Ensure port 443 is not blocked by a firewall

### MinIO bucket not found

The backend creates the S3 bucket automatically on startup. If it fails, check MinIO credentials match between `MINIO_ROOT_USER`/`MINIO_ROOT_PASSWORD` in `.env` and that MinIO is healthy:

```bash
docker compose ps minio
```

### Health check returns connection refused

The backend listens on port 8080 internally. Verify Nginx is proxying correctly:

```bash
# Test backend directly (from the server)
curl http://127.0.0.1:8080/v1/health

# Test through Nginx
curl https://yourdomain.tld/v1/health
```
