# SharePwd — Burn After Reading

Self-hosted, zero-knowledge secret sharing. Encrypt passwords, API keys, and files in your browser — the server never sees plaintext.

**Live instance**: [sharepwd.io](https://sharepwd.io)

## Why SharePwd?

Most "secret sharing" tools decrypt on the server. SharePwd encrypts and decrypts entirely in the browser using AES-256-GCM. The server stores only ciphertext and has no way to read your data.

| Feature | SharePwd | Typical alternatives |
|---|---|---|
| Encryption | Client-side AES-256-GCM | Server-side or none |
| Key derivation | PBKDF2 600k iterations | Varies |
| Server access to plaintext | Never | Often yes |
| File sharing | Chunked + encrypted | Rarely supported |
| Self-hostable | Yes (Docker Compose) | Sometimes |
| Bot protection | Built-in (Slack, Teams, etc.) | No |
| Open source | AGPLv3 | Varies |

## Architecture

```
┌──────────────┐     HTTPS     ┌─────────┐     ┌──────────┐
│   Browser    │◄─────────────►│  Nginx  │────►│ Frontend │  Next.js 15
│ (encryption) │               │  (TLS)  │     └──────────┘
└──────────────┘               │         │     ┌──────────┐
                               │         │────►│ Backend  │  Go 1.24
                               └─────────┘     └────┬─────┘
                                                    │
                                          ┌─────────┴─────────┐
                                          │                   │
                                     ┌────▼─────┐       ┌─────▼─────┐
                                     │PostgreSQL│       │   MinIO   │
                                     │   16     │       │  (S3)     │
                                     └──────────┘       └───────────┘
```

**6 containers**: Nginx (reverse proxy + TLS), Frontend (Next.js 15 / React 19), Backend (Go / Chi), PostgreSQL 16, MinIO (S3-compatible file storage), Umami (privacy-focused analytics).

## Security Model

- **AES-256-GCM** encryption/decryption happens exclusively in the browser
- **PBKDF2** with 600,000 iterations for passphrase-based key derivation (SHA-256)
- Encryption key is stored in the **URL fragment** (`#key`) — never sent to the server
- Optional **passphrase** for additional protection (key derived from passphrase instead of URL)
- **Challenge-nonce** system prevents replay attacks on secret reveal
- **Bot detection** blocks link previews (Slack, Teams, Discord, WhatsApp, etc.) from consuming views
- **Rate limiting** (30 req/min per IP)
- Non-root containers, security headers (HSTS, CSP, X-Frame-Options DENY)

## Features

- **Text secrets** — up to 100,000 characters
- **Encrypted file sharing** — chunked upload, up to 100MB
- **Burn after reading** — secret destroyed after first view (with grace period)
- **Expiration** — 5 minutes to 30 days
- **Max views** — limit how many times a secret can be revealed
- **API keys** — programmatic access with per-key rate limits
- **Self-hosted analytics** — Umami, no third-party tracking

## Quick Start

### Prerequisites

- Docker Engine + Docker Compose plugin
- A domain with DNS pointing to your server
- TLS certificates (Let's Encrypt recommended)

### 1. Clone

```bash
git clone https://github.com/AntoninHY/sharepwd.git
cd sharepwd
```

### 2. Configure

```bash
cp deploy/.env.example deploy/.env
```

Edit `deploy/.env` and set **strong, unique passwords** for:
- `POSTGRES_PASSWORD`
- `MINIO_ROOT_PASSWORD`
- `UMAMI_DB_PASSWORD`
- `UMAMI_APP_SECRET`

Update `BASE_URL`, `CORS_ORIGINS`, `NEXT_PUBLIC_API_URL`, and `NEXT_PUBLIC_APP_URL` to match your domain.

### 3. TLS Certificates

Place your certificates in the standard Let's Encrypt path or update `deploy/nginx/nginx.conf`:

```
/etc/letsencrypt/live/yourdomain.tld/fullchain.pem
/etc/letsencrypt/live/yourdomain.tld/privkey.pem
```

### 4. Deploy

```bash
cd deploy
docker compose up -d --build
```

### 5. Verify

```bash
# All containers should be running
docker ps

# Health check
curl https://yourdomain.tld/v1/health
# → {"status":"ok"}
```

## Development

```bash
# Start infrastructure (PostgreSQL + MinIO)
make infra

# Start all services in dev mode (hot reload)
make dev

# View logs
make logs
```

Dev mode uses [Air](https://github.com/air-verse/air) for Go hot reload and `next dev` for the frontend.

## API

Full API documentation is available at `/docs` on any running instance.

### Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/v1/secrets` | Create an encrypted secret |
| `GET` | `/v1/secrets/:token` | Get secret metadata (expiry, views, type) |
| `POST` | `/v1/secrets/:token/reveal` | Reveal encrypted content (with challenge nonce) |
| `DELETE` | `/v1/secrets/:token` | Delete a secret (requires creator token) |
| `POST` | `/v1/files/init` | Initialize chunked file upload |
| `POST` | `/v1/files/:id/chunks/:index` | Upload a file chunk |
| `POST` | `/v1/files/:id/complete` | Finalize file upload |
| `GET` | `/v1/files/:id/chunks/:index` | Download a file chunk |
| `GET` | `/v1/health` | Health check |

### Example: Create and reveal a secret

```bash
# Create (encrypted_data and iv are produced client-side)
curl -X POST https://sharepwd.io/v1/secrets \
  -H "Content-Type: application/json" \
  -d '{"encrypted_data":"...","iv":"...","expires_in":"24h","max_views":1}'

# Response
# {"access_token":"abc123","creator_token":"def456","expires_at":"..."}

# Get metadata + challenge nonce
curl https://sharepwd.io/v1/secrets/abc123

# Reveal (submit the challenge nonce)
curl -X POST https://sharepwd.io/v1/secrets/abc123/reveal \
  -H "Content-Type: application/json" \
  -d '{"challenge_nonce":"..."}'

# Response
# {"encrypted_data":"...","iv":"..."}
```

## Tech Stack

| Component | Technology |
|-----------|-----------|
| Backend | Go 1.24, Chi router, pgx |
| Frontend | Next.js 15, React 19, Tailwind CSS 4 |
| Database | PostgreSQL 16 |
| File storage | MinIO (S3-compatible) |
| Reverse proxy | Nginx (TLS termination) |
| Analytics | Umami (self-hosted) |
| Encryption | Web Crypto API (AES-256-GCM, PBKDF2) |

## Configuration

All configuration is done through environment variables. See [`deploy/.env.example`](deploy/.env.example) for the full list.

| Variable | Description | Default |
|----------|-------------|---------|
| `POSTGRES_PASSWORD` | PostgreSQL password | — (required) |
| `MINIO_ROOT_PASSWORD` | MinIO admin password | — (required) |
| `BASE_URL` | Public URL of the instance | `https://sharepwd.io` |
| `CORS_ORIGINS` | Allowed CORS origins | `https://sharepwd.io` |
| `MAX_TEXT_SIZE` | Max text secret size (bytes) | `102400` |
| `MAX_FILE_SIZE` | Max file size (bytes) | `104857600` |
| `CLEANUP_INTERVAL` | Expired secrets cleanup interval | `60s` |
| `RATE_LIMIT_PUBLIC` | Requests per minute per IP | `30` |
| `STORAGE_BACKEND` | File storage: `s3` or `local` | `s3` |

## License

[GNU Affero General Public License v3.0](LICENSE)

Built by [Jizo AI](https://jizo.ai).
