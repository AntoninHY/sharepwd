// SharePwd.io — Burn After Reading
// Zero-knowledge secret sharing. Client-side AES-256-GCM encryption, secrets self-destruct after viewing.
// Copyright (c) 2025-2026 Antonin HILY — CTO, Jizo AI

import type { Metadata } from "next";
import { getTranslations } from "next-intl/server";

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations("docs");
  return {
    title: t("title"),
  };
}

export default async function DocsPage() {
  const t = await getTranslations("docs");

  return (
    <div className="max-w-3xl mx-auto prose prose-invert">
      <h1 className="text-3xl font-bold mb-8">{t("title")}</h1>

      <section className="mb-12">
        <h2 className="text-xl font-semibold mb-4">{t("baseUrl")}</h2>
        <code className="block rounded-lg bg-muted px-4 py-3 text-sm">
          https://sharepwd.io/v1
        </code>
      </section>

      <section className="mb-12">
        <h2 className="text-xl font-semibold mb-4">{t("auth")}</h2>
        <p className="text-muted-foreground">{t("authDesc")}</p>
      </section>

      <section className="mb-12">
        <h2 className="text-xl font-semibold mb-4">{t("endpoints")}</h2>

        <div className="space-y-8">
          <div className="rounded-xl border border-border p-6">
            <div className="flex items-center gap-3 mb-3">
              <span className="rounded bg-success/20 text-success px-2 py-1 text-xs font-mono font-bold">POST</span>
              <code className="text-sm">/v1/secrets</code>
            </div>
            <p className="text-sm text-muted-foreground mb-4">{t("createSecretDesc")}</p>
            <h4 className="text-sm font-medium mb-2">{t("requestBody")}</h4>
            <pre className="rounded-lg bg-muted px-4 py-3 text-xs overflow-x-auto">{`{
  "encrypted_data": "base64_ciphertext",
  "iv": "base64_iv",
  "salt": "base64_salt_or_null",
  "max_views": 3,
  "expires_in": "24h",
  "burn_after_read": false,
  "content_type": "text"
}`}</pre>
            <h4 className="text-sm font-medium mt-4 mb-2">{t("response")}</h4>
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
            <p className="text-sm text-muted-foreground">{t("getMetadataDesc")}</p>
          </div>

          <div className="rounded-xl border border-border p-6">
            <div className="flex items-center gap-3 mb-3">
              <span className="rounded bg-success/20 text-success px-2 py-1 text-xs font-mono font-bold">POST</span>
              <code className="text-sm">/v1/secrets/:token/reveal</code>
            </div>
            <p className="text-sm text-muted-foreground mb-4">{t("revealDesc")}</p>
            <h4 className="text-sm font-medium mb-2">{t("requestBody")}</h4>
            <pre className="rounded-lg bg-muted px-4 py-3 text-xs overflow-x-auto">{`{
  "challenge_nonce": "nonce_from_metadata"
}`}</pre>
          </div>

          <div className="rounded-xl border border-border p-6">
            <div className="flex items-center gap-3 mb-3">
              <span className="rounded bg-destructive/20 text-destructive px-2 py-1 text-xs font-mono font-bold">DELETE</span>
              <code className="text-sm">/v1/secrets/:token</code>
            </div>
            <p className="text-sm text-muted-foreground mb-4">{t("deleteDesc")}</p>
            <h4 className="text-sm font-medium mb-2">{t("requestBody")}</h4>
            <pre className="rounded-lg bg-muted px-4 py-3 text-xs overflow-x-auto">{`{
  "creator_token": "your_creator_token"
}`}</pre>
          </div>

          <div className="rounded-xl border border-border p-6">
            <div className="flex items-center gap-3 mb-3">
              <span className="rounded bg-primary/20 text-primary px-2 py-1 text-xs font-mono font-bold">GET</span>
              <code className="text-sm">/v1/health</code>
            </div>
            <p className="text-sm text-muted-foreground">{t("healthDesc")}</p>
          </div>
        </div>
      </section>

      <section className="mb-12">
        <h2 className="text-xl font-semibold mb-4">{t("rateLimits")}</h2>
        <p className="text-muted-foreground mb-4">
          {t("rateLimitsDesc", { limit: 30 })}
        </p>
        <p className="text-sm text-muted-foreground">
          {t("rateLimitsExceeded")} <code className="text-xs bg-muted px-1.5 py-0.5 rounded">429 Too Many Requests</code>.
        </p>
      </section>

      <section className="mb-12">
        <h2 className="text-xl font-semibold mb-4">{t("encryption")}</h2>
        <p className="text-muted-foreground mb-4">{t("encryptionDesc")}</p>
        <ul className="list-disc pl-6 space-y-2 text-sm text-muted-foreground">
          <li>{t("encAlgo")}</li>
          <li>{t("encKdf")}</li>
          <li>{t("encNoPass")}</li>
          <li>{t("encFragment")}</li>
        </ul>
      </section>
    </div>
  );
}
