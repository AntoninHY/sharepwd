# API Key Management

API keys allow programmatic access to SharePwd (CLI, scripts, CI/CD pipelines). This guide covers the full lifecycle: generating the admin secret, bootstrapping the first key, and managing keys.

## Overview

SharePwd uses a two-tier authentication model:

| Credential | Purpose | How to obtain |
|-----------|---------|---------------|
| **Admin secret** (`ADMIN_SECRET`) | Bootstrap the first API key | Generated manually, set in server `.env` |
| **API key** (`spwd_...`) | Programmatic access, key management | Created via admin secret or existing API key |

**Bootstrap flow:** Generate admin secret → set in `.env` → create first API key → use API key for all further operations.

## Admin Secret

The admin secret is a shared secret between you and the server. It exists solely to create the first API key — after that, API keys are self-managing (any API key can create or revoke other keys).

### Generate

```bash
openssl rand -base64 48
```

### Configure

Add the generated value to your server's `deploy/.env`:

```
ADMIN_SECRET=your_generated_secret_here
```

Restart the backend to apply:

```bash
cd deploy
docker compose restart backend
```

If `ADMIN_SECRET` is not set, the admin bootstrap endpoint (`POST /v1/admin/api-keys`) is disabled.

## Creating API keys

### Bootstrap: First key (admin secret)

**With the CLI:**

```bash
sharepwd admin keys create \
  --name "initial-key" \
  --admin-secret "$ADMIN_SECRET" \
  --server https://yourdomain.tld
```

**With curl:**

```bash
curl -s -X POST https://yourdomain.tld/v1/admin/api-keys \
  -H "Content-Type: application/json" \
  -H "X-Admin-Secret: $ADMIN_SECRET" \
  -d '{"name": "initial-key"}'
```

Response:

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "key": "spwd_1a2b3c4d5e6f...",
  "key_prefix": "spwd_1a2b",
  "name": "initial-key",
  "rate_limit": 60
}
```

Save the `key` value immediately — it is displayed only once.

### Subsequent keys (API key)

Once you have an API key, use it to create more:

**With the CLI:**

```bash
sharepwd admin keys create \
  --name "ci-pipeline" \
  --rate-limit 120 \
  --expires-at "2026-12-31T23:59:59Z" \
  --admin-secret "$ADMIN_SECRET" \
  --server https://yourdomain.tld
```

**With curl:**

```bash
curl -s -X POST https://yourdomain.tld/v1/api-keys \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $API_KEY" \
  -d '{"name": "ci-pipeline", "rate_limit": 120, "expires_at": "2026-12-31T23:59:59Z"}'
```

Optional fields:
- `rate_limit` — requests per minute (default: 60)
- `expires_at` — RFC 3339 expiration time (default: never)

## Listing keys

**With the CLI:**

```bash
sharepwd admin keys list --api-key "$API_KEY"
```

Outputs a table:

```
  ID         PREFIX      NAME           RATE   ACTIVE   CREATED
  550e84…    spwd_1a2b   initial-key    60     true     2026-03-01
  661f95…    spwd_9x8w   ci-pipeline    120    true     2026-03-02
```

**With curl:**

```bash
curl -s https://yourdomain.tld/v1/api-keys \
  -H "Authorization: Bearer $API_KEY"
```

## Revoking keys

**With the CLI:**

```bash
sharepwd admin keys revoke \
  --id "550e8400-e29b-41d4-a716-446655440000" \
  --api-key "$API_KEY"
```

**With curl:**

```bash
curl -s -X DELETE https://yourdomain.tld/v1/api-keys/550e8400-e29b-41d4-a716-446655440000 \
  -H "Authorization: Bearer $API_KEY"
```

Revoked keys are immediately deactivated and cannot be used for any operations.

## Environment variables

Set these to avoid passing credentials as flags:

```bash
export SHAREPWD_ADMIN_SECRET="your_admin_secret"
export SHAREPWD_API_KEY="spwd_your_api_key"
```

Then commands simplify to:

```bash
sharepwd admin keys create --name "new-key"
sharepwd admin keys list
sharepwd admin keys revoke --id "..."
```

## API key format

Keys follow the format `spwd_` followed by 64 hex characters (32 random bytes):

```
spwd_a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2
```

The server stores only the SHA-256 hash — the full key cannot be recovered after creation.

## What API keys unlock

API key holders bypass anti-bot defense layers 4 (behavioral analysis) and 5 (environment fingerprint) automatically. Layers 1–3 (grace period, proof-of-work, nonce validation) still apply.

## Security best practices

- **Rotate keys regularly** — create a new key, update your systems, then revoke the old one
- **Revoke unused keys** — list keys periodically and revoke any that are no longer needed
- **Use environment variables** — never pass secrets as command-line flags in shared environments (they appear in process listings)
- **Set expiration dates** — use `--expires-at` for temporary or CI/CD keys
- **Use per-purpose keys** — create separate keys for different systems (CI, monitoring, scripts) so you can revoke individually
- **Protect the admin secret** — store it securely and only use it for bootstrapping. For day-to-day operations, use API keys
