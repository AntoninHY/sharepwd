import type { Metadata } from "next";
import { getTranslations } from "next-intl/server";

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations("cliDocs");
  return {
    title: t("title"),
  };
}

export default async function CliDocsPage() {
  const t = await getTranslations("cliDocs");

  return (
    <div className="max-w-3xl mx-auto prose prose-invert">
      <h1 className="text-3xl font-bold mb-4">{t("title")}</h1>

      <div className="rounded-xl border border-yellow-500/30 bg-yellow-500/5 p-4 mb-8">
        <p className="text-sm text-yellow-400/90 m-0">
          {t("selfHostNotice")}
        </p>
      </div>

      <section className="mb-12">
        <h2 className="text-xl font-semibold mb-4">{t("commands")}</h2>

        <div className="space-y-8">
          <div className="rounded-xl border border-border p-6">
            <div className="flex items-center gap-3 mb-3">
              <span className="rounded bg-success/20 text-success px-2 py-1 text-xs font-mono font-bold">push</span>
              <span className="text-sm text-muted-foreground">{t("pushDesc")}</span>
            </div>
            <pre className="rounded-lg bg-muted px-4 py-3 text-xs overflow-x-auto">{`# Share a text secret
sharepwd push "db_password=S3cureP@ss!" --burn --ttl 1h

# Share a file
sharepwd push -f secret.pdf --burn --ttl 24h

# From stdin (pipe)
echo "secret" | sharepwd push --burn

# With passphrase protection
sharepwd push -p "my secret" --ttl 1h

# JSON output (for scripts)
sharepwd push --json --ttl 1h "secret" 2>/dev/null | jq .url

# Copy URL to clipboard directly
sharepwd push "secret" 2>/dev/null | pbcopy`}</pre>
            <h4 className="text-sm font-medium mt-4 mb-2">{t("flags")}</h4>
            <div className="overflow-x-auto">
              <table className="w-full text-xs">
                <thead>
                  <tr className="border-b border-border">
                    <th className="text-left py-2 pr-4 font-medium">{t("flag")}</th>
                    <th className="text-left py-2 font-medium">{t("description")}</th>
                  </tr>
                </thead>
                <tbody className="text-muted-foreground">
                  <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono">-f &lt;path&gt;</td><td className="py-2">{t("flagFile")}</td></tr>
                  <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono">-p</td><td className="py-2">{t("flagPassphrase")}</td></tr>
                  <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono">--burn</td><td className="py-2">{t("flagBurn")}</td></tr>
                  <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono">--ttl &lt;duration&gt;</td><td className="py-2">{t("flagTTL")}</td></tr>
                  <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono">--max-views &lt;n&gt;</td><td className="py-2">{t("flagMaxViews")}</td></tr>
                  <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono">--server &lt;url&gt;</td><td className="py-2">{t("flagServer")}</td></tr>
                  <tr><td className="py-2 pr-4 font-mono">--json</td><td className="py-2">{t("flagJSON")}</td></tr>
                </tbody>
              </table>
            </div>
          </div>

          <div className="rounded-xl border border-border p-6">
            <div className="flex items-center gap-3 mb-3">
              <span className="rounded bg-primary/20 text-primary px-2 py-1 text-xs font-mono font-bold">pull</span>
              <span className="text-sm text-muted-foreground">{t("pullDesc")}</span>
            </div>
            <pre className="rounded-lg bg-muted px-4 py-3 text-xs overflow-x-auto">{`# Retrieve a text secret
sharepwd pull https://sharepwd.io/s/abc123#key456

# Retrieve a file
sharepwd pull -o output.pdf https://sharepwd.io/f/abc123#key456

# With passphrase
sharepwd pull -p https://sharepwd.io/s/abc123`}</pre>
            <h4 className="text-sm font-medium mt-4 mb-2">{t("flags")}</h4>
            <div className="overflow-x-auto">
              <table className="w-full text-xs">
                <thead>
                  <tr className="border-b border-border">
                    <th className="text-left py-2 pr-4 font-medium">{t("flag")}</th>
                    <th className="text-left py-2 font-medium">{t("description")}</th>
                  </tr>
                </thead>
                <tbody className="text-muted-foreground">
                  <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono">-p</td><td className="py-2">{t("flagPassphraseDecrypt")}</td></tr>
                  <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono">-o &lt;path&gt;</td><td className="py-2">{t("flagOutput")}</td></tr>
                  <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono">--server &lt;url&gt;</td><td className="py-2">{t("flagServer")}</td></tr>
                  <tr><td className="py-2 pr-4 font-mono">--json</td><td className="py-2">{t("flagJSON")}</td></tr>
                </tbody>
              </table>
            </div>
          </div>

          <div className="rounded-xl border border-border p-6">
            <div className="flex items-center gap-3 mb-3">
              <span className="rounded bg-destructive/20 text-destructive px-2 py-1 text-xs font-mono font-bold">delete</span>
              <span className="text-sm text-muted-foreground">{t("deleteDesc")}</span>
            </div>
            <pre className="rounded-lg bg-muted px-4 py-3 text-xs overflow-x-auto">{`sharepwd delete --creator-token <token> https://sharepwd.io/s/abc123`}</pre>
          </div>

          <div className="rounded-xl border border-border p-6">
            <div className="flex items-center gap-3 mb-3">
              <span className="rounded bg-muted text-muted-foreground px-2 py-1 text-xs font-mono font-bold">version</span>
              <span className="text-sm text-muted-foreground">{t("versionDesc")}</span>
            </div>
            <pre className="rounded-lg bg-muted px-4 py-3 text-xs overflow-x-auto">sharepwd version</pre>
          </div>
        </div>
      </section>

      <section className="mb-12">
        <h2 className="text-xl font-semibold mb-4">{t("selfHost")}</h2>
        <p className="text-muted-foreground mb-4">{t("selfHostDesc")}</p>
        <pre className="rounded-lg bg-muted px-4 py-3 text-xs overflow-x-auto">{`# Point to your own instance
sharepwd push --server https://secrets.yourcompany.com "secret" --burn

sharepwd pull --server https://secrets.yourcompany.com https://secrets.yourcompany.com/s/abc123#key`}</pre>
      </section>

      <section className="mb-12">
        <h2 className="text-xl font-semibold mb-4">{t("exitCodes")}</h2>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-border">
                <th className="text-left py-2 pr-4 font-medium">{t("code")}</th>
                <th className="text-left py-2 font-medium">{t("meaning")}</th>
              </tr>
            </thead>
            <tbody className="text-muted-foreground">
              <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono">0</td><td className="py-2">{t("exit0")}</td></tr>
              <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono">1</td><td className="py-2">{t("exit1")}</td></tr>
              <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono">2</td><td className="py-2">{t("exit2")}</td></tr>
              <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono">3</td><td className="py-2">{t("exit3")}</td></tr>
              <tr><td className="py-2 pr-4 font-mono">4</td><td className="py-2">{t("exit4")}</td></tr>
            </tbody>
          </table>
        </div>
      </section>

      <section className="mb-12">
        <h2 className="text-xl font-semibold mb-4">{t("encryption")}</h2>
        <p className="text-muted-foreground mb-4">{t("encryptionDesc")}</p>
        <ul className="list-disc pl-6 space-y-2 text-sm text-muted-foreground">
          <li>{t("encCompat")}</li>
          <li>{t("encAlgo")}</li>
          <li>{t("encKdf")}</li>
          <li>{t("encKeyless")}</li>
        </ul>
      </section>
    </div>
  );
}
