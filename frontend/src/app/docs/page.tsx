export default function DocsPage() {
  return (
    <div className="max-w-3xl mx-auto prose prose-invert">
      <h1 className="text-3xl font-bold mb-8">API Documentation</h1>

      <section className="mb-12">
        <h2 className="text-xl font-semibold mb-4">Base URL</h2>
        <code className="block rounded-lg bg-muted px-4 py-3 text-sm">
          https://sharepwd.io/v1
        </code>
      </section>

      <section className="mb-12">
        <h2 className="text-xl font-semibold mb-4">Authentication</h2>
        <p className="text-muted-foreground">
          All endpoints are public and do not require authentication. Rate limited to 30 requests/minute per IP.
        </p>
      </section>

      {/* TODO: Uncomment when paid plans with API keys are implemented
      <section className="mb-12">
        <h2 className="text-xl font-semibold mb-4">API Keys</h2>
        <p className="text-muted-foreground mb-4">
          For higher rate limits, create an API key and include it in requests:
        </p>
        <code className="block rounded-lg bg-muted px-4 py-3 text-sm mt-2">
          Authorization: Bearer spwd_your_api_key_here
        </code>
      </section>
      */}

      <section className="mb-12">
        <h2 className="text-xl font-semibold mb-4">Endpoints</h2>

        <div className="space-y-8">
          <div className="rounded-xl border border-border p-6">
            <div className="flex items-center gap-3 mb-3">
              <span className="rounded bg-success/20 text-success px-2 py-1 text-xs font-mono font-bold">POST</span>
              <code className="text-sm">/v1/secrets</code>
            </div>
            <p className="text-sm text-muted-foreground mb-4">Create a new encrypted secret.</p>
            <h4 className="text-sm font-medium mb-2">Request Body</h4>
            <pre className="rounded-lg bg-muted px-4 py-3 text-xs overflow-x-auto">{`{
  "encrypted_data": "base64_ciphertext",
  "iv": "base64_iv",
  "salt": "base64_salt_or_null",
  "max_views": 3,
  "expires_in": "24h",
  "burn_after_read": false,
  "content_type": "text"
}`}</pre>
            <h4 className="text-sm font-medium mt-4 mb-2">Response (201)</h4>
            <pre className="rounded-lg bg-muted px-4 py-3 text-xs overflow-x-auto">{`{
  "access_token": "abc123...",
  "creator_token": "def456...",
  "expires_at": "2025-01-01T00:00:00Z"
}`}</pre>
          </div>

          <div className="rounded-xl border border-border p-6">
            <div className="flex items-center gap-3 mb-3">
              <span className="rounded bg-primary/20 text-primary px-2 py-1 text-xs font-mono font-bold">GET</span>
              <code className="text-sm">/v1/secrets/:token</code>
            </div>
            <p className="text-sm text-muted-foreground">Get secret metadata (without decrypted content). Returns a challenge nonce for reveal.</p>
          </div>

          <div className="rounded-xl border border-border p-6">
            <div className="flex items-center gap-3 mb-3">
              <span className="rounded bg-success/20 text-success px-2 py-1 text-xs font-mono font-bold">POST</span>
              <code className="text-sm">/v1/secrets/:token/reveal</code>
            </div>
            <p className="text-sm text-muted-foreground mb-4">Reveal the encrypted secret. Consumes a view. Requires a valid challenge nonce.</p>
            <h4 className="text-sm font-medium mb-2">Request Body</h4>
            <pre className="rounded-lg bg-muted px-4 py-3 text-xs overflow-x-auto">{`{
  "challenge_nonce": "nonce_from_metadata"
}`}</pre>
          </div>

          <div className="rounded-xl border border-border p-6">
            <div className="flex items-center gap-3 mb-3">
              <span className="rounded bg-destructive/20 text-destructive px-2 py-1 text-xs font-mono font-bold">DELETE</span>
              <code className="text-sm">/v1/secrets/:token</code>
            </div>
            <p className="text-sm text-muted-foreground mb-4">Delete a secret. Requires the creator token.</p>
            <h4 className="text-sm font-medium mb-2">Request Body</h4>
            <pre className="rounded-lg bg-muted px-4 py-3 text-xs overflow-x-auto">{`{
  "creator_token": "your_creator_token"
}`}</pre>
          </div>

          <div className="rounded-xl border border-border p-6">
            <div className="flex items-center gap-3 mb-3">
              <span className="rounded bg-primary/20 text-primary px-2 py-1 text-xs font-mono font-bold">GET</span>
              <code className="text-sm">/v1/health</code>
            </div>
            <p className="text-sm text-muted-foreground">Health check endpoint.</p>
          </div>
        </div>
      </section>

      <section className="mb-12">
        <h2 className="text-xl font-semibold mb-4">Rate Limits</h2>
        <p className="text-muted-foreground mb-4">
          All endpoints are rate limited to <strong>30 requests per minute</strong> per IP address.
        </p>
        <p className="text-sm text-muted-foreground">
          When the limit is exceeded, the API returns <code className="text-xs bg-muted px-1.5 py-0.5 rounded">429 Too Many Requests</code>.
        </p>
      </section>

      <section className="mb-12">
        <h2 className="text-xl font-semibold mb-4">Encryption</h2>
        <p className="text-muted-foreground mb-4">
          SharePwd uses <strong>zero-knowledge encryption</strong>. The server never sees your plaintext data.
        </p>
        <ul className="list-disc pl-6 space-y-2 text-sm text-muted-foreground">
          <li>Algorithm: AES-256-GCM (256-bit key, 96-bit IV, 128-bit auth tag)</li>
          <li>Key derivation (passphrase): PBKDF2 with SHA-256, 600,000 iterations</li>
          <li>Without passphrase: random key embedded in URL fragment (#)</li>
          <li>The URL fragment is never sent to the server (per HTTP spec)</li>
        </ul>
      </section>
    </div>
  );
}
