# CLI Reference

The `sharepwd` CLI lets you share and retrieve secrets from the command line.

## Installation

Build from source (requires Go 1.24+):

```bash
cd cli
go build -o /tmp/sharepwd -ldflags "-s -w" .
sudo mv /tmp/sharepwd /usr/local/bin/sharepwd
```

Verify:

```bash
sharepwd version
```

## Commands

### `sharepwd push` — Share a secret

Encrypt and upload a secret to SharePwd.

```bash
# Text (positional argument)
sharepwd push "db_password=S3cureP@ss!"

# Text (stdin pipe)
echo "secret" | sharepwd push

# File
sharepwd push -f /path/to/secret.pdf
```

**Flags:**

| Flag | Description |
|------|-------------|
| `-f <path>` | Path to file to encrypt and share |
| `-p` | Protect with a passphrase (interactive prompt) |
| `--burn` | Burn after reading (delete after first view) |
| `--ttl <duration>` | Time to live (e.g., `1h`, `24h`, `7d`) |
| `--max-views <n>` | Maximum number of views (0 = unlimited) |
| `--server <url>` | SharePwd server URL (default: `https://sharepwd.io`) |
| `--json` | Output as JSON |

**Output:** The share URL is printed to stdout (pipe-friendly). Status messages go to stderr.

```bash
# Copy URL to clipboard
sharepwd push "secret" --burn --ttl 1h | pbcopy

# JSON output
sharepwd push "secret" --json
# {"url":"https://sharepwd.io/s/abc123#key","creator_token":"def456","expires_at":"..."}
```

---

### `sharepwd pull` — Retrieve a secret

Download and decrypt a secret.

```bash
# Text secret
sharepwd pull https://sharepwd.io/s/abc123#key456

# Passphrase-protected secret
sharepwd pull https://sharepwd.io/s/abc123 -p

# File secret (save to disk)
sharepwd pull https://sharepwd.io/f/xyz789#key -o ./output.pdf
```

**Flags:**

| Flag | Description |
|------|-------------|
| `-p` | Decrypt with passphrase (interactive prompt) |
| `-o <path>` | Output file path (for file secrets) |
| `--server <url>` | Override API server URL |
| `--json` | Output as JSON |

---

### `sharepwd delete` — Delete a secret

Delete a secret using the creator token received during `push`.

```bash
sharepwd delete https://sharepwd.io/s/abc123 --creator-token def789
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--creator-token <token>` | Creator token (required) |
| `--server <url>` | Override API server URL |
| `--json` | Output as JSON |

---

### `sharepwd admin keys` — API key management

Manage API keys for programmatic access. See [API Key Management](api-keys.md) for the full guide.

#### `sharepwd admin keys create`

Create a new API key. Requires the admin secret.

```bash
sharepwd admin keys create --name "ci-pipeline" --admin-secret "$ADMIN_SECRET"
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--name <name>` | Key name (required) |
| `--admin-secret <secret>` | Admin secret (or `SHAREPWD_ADMIN_SECRET` env var) |
| `--rate-limit <n>` | Rate limit in requests/minute (0 = server default of 60) |
| `--expires-at <rfc3339>` | Expiration time (e.g., `2026-12-31T23:59:59Z`) |
| `--server <url>` | SharePwd server URL |
| `--json` | Output as JSON |

The raw API key is printed to stdout (pipe-friendly). Save it immediately — it cannot be retrieved later.

```bash
# Save key to a variable
KEY=$(sharepwd admin keys create --name "bot" --admin-secret "$SECRET")
```

#### `sharepwd admin keys list`

List all API keys. Requires an existing API key.

```bash
sharepwd admin keys list --api-key "$API_KEY"
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--api-key <key>` | API key (or `SHAREPWD_API_KEY` env var) |
| `--server <url>` | SharePwd server URL |
| `--json` | Output as JSON |

#### `sharepwd admin keys revoke`

Revoke an API key. Requires an existing API key.

```bash
sharepwd admin keys revoke --id "550e8400-e29b-41d4-a716-446655440000" --api-key "$API_KEY"
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--id <uuid>` | API key ID to revoke (required) |
| `--api-key <key>` | API key (or `SHAREPWD_API_KEY` env var) |
| `--server <url>` | SharePwd server URL |
| `--json` | Output as JSON |

---

### `sharepwd version`

Print version, commit hash, and build date.

```bash
sharepwd version
```

## Environment Variables

| Variable | Description | Used by |
|----------|-------------|---------|
| `SHAREPWD_ADMIN_SECRET` | Admin secret for key creation | `admin keys create` |
| `SHAREPWD_API_KEY` | API key for authenticated operations | `admin keys list`, `admin keys revoke` |

Environment variables are overridden by their corresponding flags.

## Self-Hosted Instances

By default, all commands target `https://sharepwd.io`. Use `--server` to point to your own instance:

```bash
sharepwd push "secret" --server https://pwd.example.com
sharepwd pull https://pwd.example.com/s/abc123#key --server https://pwd.example.com
sharepwd admin keys list --server https://pwd.example.com
```

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | Usage or argument error |
| `2` | API error (server returned an error) |
| `3` | Encryption or decryption error |
| `4` | I/O error (file read/write, stdin) |
