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

### Zero-Knowledge Encryption

- **AES-256-GCM** encryption/decryption happens exclusively in the browser
- **PBKDF2** with 600,000 iterations for passphrase-based key derivation (SHA-256)
- Encryption key is stored in the **URL fragment** (`#key`) — never sent to the server
- Optional **passphrase** for additional protection (key derived from passphrase instead of URL)

### 5-Layer Anti-Bot Defense

SharePwd uses 5 composable defense layers to make automated scraping of secrets economically impractical, without degrading user experience:

| Layer | Mechanism | Purpose |
|-------|-----------|---------|
| **1. Grace Period** | Server-enforced minimum time (1.5s) between nonce issuance and reveal | Prevents instant automated reveals |
| **2. Proof-of-Work** | SHA-256 hashcash (configurable difficulty) solved in a Web Worker | Forces computational cost per reveal — invisible to users |
| **3. Hardened Nonces** | Single-use, IP-bound, per-IP limit (3 active), 5-minute TTL | Prevents nonce farming and replay attacks |
| **4. Behavioral Analysis** | Passive mouse movement entropy, velocity variance, straight-line ratio scoring | Detects automated interactions without CAPTCHAs |
| **5. Environment Fingerprint** | Detects `navigator.webdriver`, missing plugins, zero-size screens, etc. | Blocks headless browsers (Puppeteer, Playwright) |

All layers are validated server-side. Bypassing one layer is possible — bypassing all five simultaneously at scale is not.

**Rollout modes:**
- `DEFENSE_STRICT_MODE=false` (default) — all layers active but missing proofs are logged, not blocked. Safe for gradual deployment.
- `DEFENSE_STRICT_MODE=true` — all proofs required. Requests without PoW, behavioral proof, or env fingerprint are rejected with 403.

**API key holders** (CLI, programmatic access) skip layers 4 and 5 automatically.

### Infrastructure Security

- **Bot detection** blocks link previews (Slack, Teams, Discord, WhatsApp, 40+ patterns) and empty User-Agents
- **Rate limiting** — 30 req/min per IP globally, 10 req/min on metadata endpoint
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

```bash
git clone https://github.com/AntoninHY/sharepwd.git
cd sharepwd
cp deploy/.env.example deploy/.env
# Edit deploy/.env — set passwords, admin secret, and domain URLs
cd deploy
docker compose up -d --build
```

Verify: `curl https://yourdomain.tld/v1/health` → `{"status":"ok"}`

See the [Installation Guide](docs/installation.md) for TLS setup, first API key creation, and troubleshooting.

## Documentation

| Guide | Description |
|-------|-------------|
| [Installation](docs/installation.md) | Full deployment guide — prerequisites, `.env` config, TLS, Docker Compose, troubleshooting |
| [CLI Reference](docs/cli.md) | Commands, flags, environment variables, exit codes |
| [API Key Management](docs/api-keys.md) | Admin secret setup, bootstrap flow, creating/listing/revoking keys |

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

### Public Endpoints

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

### Admin Endpoints (authenticated)

Require `X-Admin-Secret` header or `Authorization: Bearer <api_key>`. See [API Key Management](docs/api-keys.md).

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `POST` | `/v1/admin/api-keys` | Admin secret | Bootstrap first API key |
| `POST` | `/v1/api-keys` | API key | Create additional API keys |
| `GET` | `/v1/api-keys` | API key | List all API keys |
| `DELETE` | `/v1/api-keys/:id` | API key | Revoke an API key |

### Example: Create and reveal a secret

```bash
# Create (encrypted_data and iv are produced client-side)
curl -X POST https://sharepwd.io/v1/secrets \
  -H "Content-Type: application/json" \
  -d '{"encrypted_data":"...","iv":"...","expires_in":"24h","max_views":1}'

# Response
# {"access_token":"abc123","creator_token":"def456","expires_at":"..."}

# Get metadata + challenge nonce + PoW challenge
curl https://sharepwd.io/v1/secrets/abc123
# Response includes: challenge_nonce, pow_challenge, pow_difficulty

# Reveal (submit challenge nonce + PoW solution + proofs)
curl -X POST https://sharepwd.io/v1/secrets/abc123/reveal \
  -H "Content-Type: application/json" \
  -d '{"challenge_nonce":"...","pow_solution":12345}'

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
| `ADMIN_SECRET` | Admin secret for API key bootstrapping ([details](docs/api-keys.md)) | — (optional) |
| `BASE_URL` | Public URL of the instance | `https://sharepwd.io` |
| `CORS_ORIGINS` | Allowed CORS origins | `https://sharepwd.io` |
| `MAX_TEXT_SIZE` | Max text secret size (bytes) | `102400` |
| `MAX_FILE_SIZE` | Max file size (bytes) | `104857600` |
| `CLEANUP_INTERVAL` | Expired secrets cleanup interval | `60s` |
| `RATE_LIMIT_PUBLIC` | Requests per minute per IP | `30` |
| `STORAGE_BACKEND` | File storage: `s3` or `local` | `s3` |
| `DEFENSE_STRICT_MODE` | Enforce all defense layers (reject if missing) | `false` |
| `POW_DIFFICULTY` | Proof-of-Work difficulty (leading zero bits) | `20` |
| `CHALLENGE_MIN_SOLVE_TIME` | Minimum time between nonce issuance and reveal | `1500ms` |
| `CHALLENGE_TTL` | Nonce time-to-live | `5m` |
| `MAX_NONCES_PER_IP` | Max active nonces per IP address | `3` |
| `METADATA_RATE_LIMIT` | Requests per minute on metadata endpoint | `10` |
| `BEHAVIORAL_MIN_SCORE` | Minimum behavioral score to pass (0-100) | `30` |
| `ENV_MIN_SCORE` | Minimum environment fingerprint score (0-50) | `20` |

## License

[GNU Affero General Public License v3.0](LICENSE)

Built by [Jizo AI](https://jizo.ai).
