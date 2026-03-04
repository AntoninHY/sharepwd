# API Key Management

API keys allow programmatic access to SharePwd (CLI, scripts, CI/CD pipelines). This guide covers the full lifecycle: generating the admin secret, bootstrapping the first key, and managing keys.

## Overview

SharePwd uses a two-tier authentication model:

| Credential | Purpose | How to obtain |
|-----------|---------|---------------|
| **Admin secret** (`ADMIN_SECRET`) | Bootstrap the first API key | Generated manually, set in server `.env` |
| **API key** (`spwd_...`) | Programmatic access, key management | Created via admin secret or existing API key |

**Bootstrap flow:** Generate admin secret → set in `.env` → create first API key → export API key → use it for all further operations.

## Step 1 — Admin Secret

The admin secret is a shared secret between you and the server. It exists solely to create the first API key — after that, API keys are self-managing (any API key can create or revoke other keys).

### Generate

```bash
openssl rand -base64 48
```

### Configure

Copy the output and paste it into your server's `deploy/.env`:

```
ADMIN_SECRET=your_generated_secret_here
```

Restart the backend to apply:

```bash
cd deploy
docker compose restart backend
```

If `ADMIN_SECRET` is not set, the admin bootstrap endpoint (`POST /v1/admin/api-keys`) is disabled.

## Step 2 — Create the first API key

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

The `key` value is displayed **only once**. Save it now.

## Step 3 — Export your API key

Set the key as an environment variable so all subsequent commands can use it:

```bash
export SHAREPWD_API_KEY="spwd_...your_key_here..."
```

To persist across sessions, add the export to your shell profile (`~/.bashrc`, `~/.zshrc`, etc.).

You can also set the admin secret the same way to avoid passing it as a flag:

```bash
export SHAREPWD_ADMIN_SECRET="your_admin_secret"
```

From this point on, all examples assume these variables are set.

## Listing keys

```bash
sharepwd admin keys list
```

Outputs a table:

```
  ID         PREFIX      NAME           RATE   ACTIVE   CREATED
  550e84…    spwd_1a2b   initial-key    60     true     2026-03-01
```

With curl:

```bash
curl -s https://yourdomain.tld/v1/api-keys \
  -H "Authorization: Bearer $SHAREPWD_API_KEY"
```

## Creating more keys

Once you have an API key, use it to create additional keys:

```bash
sharepwd admin keys create \
  --name "ci-pipeline" \
  --rate-limit 120 \
  --expires-at "2026-12-31T23:59:59Z"
```

With curl:

```bash
curl -s -X POST https://yourdomain.tld/v1/api-keys \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $SHAREPWD_API_KEY" \
  -d '{"name": "ci-pipeline", "rate_limit": 120, "expires_at": "2026-12-31T23:59:59Z"}'
```

Optional fields:
- `rate_limit` — requests per minute (default: 60)
- `expires_at` — RFC 3339 expiration time (default: never)

## Revoking keys

```bash
sharepwd admin keys revoke --id "550e8400-e29b-41d4-a716-446655440000"
```

With curl:

```bash
curl -s -X DELETE https://yourdomain.tld/v1/api-keys/550e8400-e29b-41d4-a716-446655440000 \
  -H "Authorization: Bearer $SHAREPWD_API_KEY"
```

Revoked keys are immediately deactivated and cannot be used for any operations.

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
