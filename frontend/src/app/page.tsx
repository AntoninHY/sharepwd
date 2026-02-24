import { Shield, Clock, Eye, Lock, Terminal, Zap, Code, ArrowRight, CheckCircle2, XCircle, Flame } from "lucide-react";

export default function Home() {
  return (
    <div className="flex flex-col items-center gap-20 py-16">
      {/* Hero */}
      <section className="text-center max-w-3xl">
        <div className="inline-flex items-center gap-2 rounded-full border border-primary/30 bg-primary/5 px-4 py-1.5 text-xs text-primary mb-8">
          <Shield className="h-3.5 w-3.5" />
          Zero-Knowledge Encryption
        </div>
        <h1 className="text-5xl md:text-7xl font-bold tracking-tight mb-6 leading-tight">
          <span className="text-primary">Burn</span> After Reading.<br /><span className="text-2xl md:text-3xl text-muted-foreground font-medium">Share secrets. Not risks.</span>
        </h1>
        <p className="text-lg text-muted-foreground mb-10 max-w-xl mx-auto">
          Share passwords and secrets that self-destruct.
          Encrypted in your browser, unreadable on our servers, gone after viewing.
        </p>
        <div className="flex flex-col sm:flex-row items-center justify-center gap-4">
          <a
            href="/create"
            className="inline-flex items-center gap-2 rounded-lg bg-primary px-8 py-3.5 text-sm font-medium text-primary-foreground hover:bg-primary/90 transition-colors"
          >
            <Flame className="h-4 w-4" />
            Share a Secret
            <ArrowRight className="h-4 w-4" />
          </a>
          <a
            href="/docs"
            className="inline-flex items-center gap-2 rounded-lg border border-border px-8 py-3.5 text-sm font-medium hover:bg-muted transition-colors"
          >
            <Code className="h-4 w-4" />
            API Documentation
          </a>
        </div>
      </section>

      {/* Features */}
      <section className="grid grid-cols-1 md:grid-cols-3 gap-6 w-full max-w-4xl">
        <div className="rounded-xl border border-border bg-card p-6 hover:border-primary/30 transition-colors">
          <Shield className="h-8 w-8 text-primary mb-4" />
          <h3 className="font-semibold mb-2">Zero-Knowledge</h3>
          <p className="text-sm text-muted-foreground leading-relaxed">
            AES-256-GCM encryption in your browser. The key stays in the URL fragment — never sent to our servers.
          </p>
        </div>
        <div className="rounded-xl border border-border bg-card p-6 hover:border-primary/30 transition-colors">
          <Flame className="h-8 w-8 text-primary mb-4" />
          <h3 className="font-semibold mb-2">Burn After Reading</h3>
          <p className="text-sm text-muted-foreground leading-relaxed">
            Secrets self-destruct after viewing. Set view limits, time-based expiration, or one-time read. No traces left behind.
          </p>
        </div>
        <div className="rounded-xl border border-border bg-card p-6 hover:border-primary/30 transition-colors">
          <Eye className="h-8 w-8 text-primary mb-4" />
          <h3 className="font-semibold mb-2">Anti-Bot Protection</h3>
          <p className="text-sm text-muted-foreground leading-relaxed">
            5-layer defense: grace period, challenge gate, nonces, behavioral analysis, and UA detection.
          </p>
        </div>
        <div className="rounded-xl border border-border bg-card p-6 hover:border-primary/30 transition-colors">
          <Terminal className="h-8 w-8 text-primary mb-4" />
          <h3 className="font-semibold mb-2">CLI & API</h3>
          <p className="text-sm text-muted-foreground leading-relaxed">
            Full REST API and CLI tool. Integrate secret sharing into your CI/CD pipelines and scripts.
          </p>
        </div>
        <div className="rounded-xl border border-border bg-card p-6 hover:border-primary/30 transition-colors">
          <Zap className="h-8 w-8 text-primary mb-4" />
          <h3 className="font-semibold mb-2">File Sharing</h3>
          <p className="text-sm text-muted-foreground leading-relaxed">
            Share encrypted files up to 5MB. Need more? Paid plans support up to 100MB+.
          </p>
        </div>
        <div className="rounded-xl border border-border bg-card p-6 hover:border-primary/30 transition-colors">
          <Code className="h-8 w-8 text-primary mb-4" />
          <h3 className="font-semibold mb-2">Open Source</h3>
          <p className="text-sm text-muted-foreground leading-relaxed">
            AGPLv3 licensed. Audit the code, self-host it, contribute. Full transparency.
          </p>
        </div>
      </section>

      {/* How it works */}
      <section className="w-full max-w-3xl">
        <h2 className="text-2xl font-bold text-center mb-12">How It Works</h2>
        <div className="space-y-8">
          <div className="flex gap-6 items-start">
            <div className="shrink-0 w-10 h-10 rounded-full bg-primary/10 text-primary flex items-center justify-center font-bold text-sm">1</div>
            <div>
              <h3 className="font-semibold mb-1">Encrypt in your browser</h3>
              <p className="text-sm text-muted-foreground">Your secret is encrypted with AES-256-GCM using the Web Crypto API. The encryption key is generated randomly and never leaves your device.</p>
            </div>
          </div>
          <div className="flex gap-6 items-start">
            <div className="shrink-0 w-10 h-10 rounded-full bg-primary/10 text-primary flex items-center justify-center font-bold text-sm">2</div>
            <div>
              <h3 className="font-semibold mb-1">Store the encrypted blob</h3>
              <p className="text-sm text-muted-foreground">Only the ciphertext is sent to our server. We store it with your chosen expiration rules. We cannot decrypt it.</p>
            </div>
          </div>
          <div className="flex gap-6 items-start">
            <div className="shrink-0 w-10 h-10 rounded-full bg-primary/10 text-primary flex items-center justify-center font-bold text-sm">3</div>
            <div>
              <h3 className="font-semibold mb-1">Share the link — it burns after reading</h3>
              <p className="text-sm text-muted-foreground">The encryption key is embedded in the URL fragment (#), never sent to the server. The recipient decrypts locally, then the secret self-destructs.</p>
            </div>
          </div>
        </div>
      </section>

      {/* CLI example */}
      <section className="w-full max-w-3xl">
        <h2 className="text-2xl font-bold text-center mb-8">Works Everywhere</h2>
        <div className="rounded-xl border border-border bg-card overflow-hidden">
          <div className="flex items-center gap-2 px-4 py-3 border-b border-border bg-muted/30">
            <div className="w-3 h-3 rounded-full bg-destructive/50" />
            <div className="w-3 h-3 rounded-full bg-warning/50" />
            <div className="w-3 h-3 rounded-full bg-success/50" />
            <span className="text-xs text-muted-foreground ml-2 font-mono">terminal</span>
          </div>
          <pre className="px-6 py-4 text-sm font-mono text-muted-foreground overflow-x-auto leading-relaxed">
<span className="text-success">$</span> sharepwd push &quot;db_password=S3cureP@ss!&quot; --burn --ttl 1h{"\n"}
<span className="text-foreground">Secret created successfully!</span>{"\n"}
<span className="text-foreground">URL: https://sharepwd.io/s/abc123...#key456...</span>{"\n"}
{"\n"}
<span className="text-success">$</span> sharepwd pull https://sharepwd.io/s/abc123...#key456...{"\n"}
<span className="text-foreground">db_password=S3cureP@ss!</span>{"\n"}
<span className="text-destructive">Secret burned. This link is now dead.</span>
          </pre>
        </div>
      </section>

      {/* Comparison */}
      <section className="w-full max-w-3xl">
        <h2 className="text-2xl font-bold text-center mb-8">Why SharePwd?</h2>
        <div className="rounded-xl border border-border overflow-hidden">
          <table className="w-full text-sm">
            <thead>
              <tr className="bg-muted/30 border-b border-border">
                <th className="text-left px-6 py-3 font-medium">Feature</th>
                <th className="text-center px-4 py-3 font-medium text-primary">SharePwd</th>
                <th className="text-center px-4 py-3 font-medium text-muted-foreground">Others</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              <tr>
                <td className="px-6 py-3">Zero-knowledge encryption</td>
                <td className="text-center px-4 py-3"><CheckCircle2 className="h-5 w-5 text-success mx-auto" /></td>
                <td className="text-center px-4 py-3"><XCircle className="h-5 w-5 text-muted-foreground/40 mx-auto" /></td>
              </tr>
              <tr>
                <td className="px-6 py-3">Key never sent to server</td>
                <td className="text-center px-4 py-3"><CheckCircle2 className="h-5 w-5 text-success mx-auto" /></td>
                <td className="text-center px-4 py-3"><XCircle className="h-5 w-5 text-muted-foreground/40 mx-auto" /></td>
              </tr>
              <tr>
                <td className="px-6 py-3">Anti-bot protection</td>
                <td className="text-center px-4 py-3"><CheckCircle2 className="h-5 w-5 text-success mx-auto" /></td>
                <td className="text-center px-4 py-3"><XCircle className="h-5 w-5 text-muted-foreground/40 mx-auto" /></td>
              </tr>
              <tr>
                <td className="px-6 py-3">CLI tool</td>
                <td className="text-center px-4 py-3"><CheckCircle2 className="h-5 w-5 text-success mx-auto" /></td>
                <td className="text-center px-4 py-3"><CheckCircle2 className="h-5 w-5 text-success mx-auto" /></td>
              </tr>
              <tr>
                <td className="px-6 py-3">File sharing</td>
                <td className="text-center px-4 py-3"><CheckCircle2 className="h-5 w-5 text-success mx-auto" /></td>
                <td className="text-center px-4 py-3"><CheckCircle2 className="h-5 w-5 text-success mx-auto" /></td>
              </tr>
              <tr>
                <td className="px-6 py-3">Open source (AGPLv3)</td>
                <td className="text-center px-4 py-3"><CheckCircle2 className="h-5 w-5 text-success mx-auto" /></td>
                <td className="text-center px-4 py-3"><CheckCircle2 className="h-5 w-5 text-success mx-auto" /></td>
              </tr>
              <tr>
                <td className="px-6 py-3">Self-hostable</td>
                <td className="text-center px-4 py-3"><CheckCircle2 className="h-5 w-5 text-success mx-auto" /></td>
                <td className="text-center px-4 py-3"><CheckCircle2 className="h-5 w-5 text-success mx-auto" /></td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      {/* CTA */}
      <section className="text-center">
        <a
          href="/create"
          className="inline-flex items-center gap-2 rounded-lg bg-primary px-8 py-3.5 text-sm font-medium text-primary-foreground hover:bg-primary/90 transition-colors"
        >
          <Flame className="h-4 w-4" />
          Start Sharing Securely
        </a>
      </section>
    </div>
  );
}
