// SharePwd.io — Burn After Reading
// Zero-knowledge secret sharing. Client-side AES-256-GCM encryption, secrets self-destruct after viewing.
// Copyright (c) 2025-2026 Antonin HILY — CTO, Jizo AI

import { Shield, Clock, Eye, Lock, Terminal, Zap, Code, ArrowRight, CheckCircle2, XCircle, Flame, MapPin } from "lucide-react";
import { getTranslations } from "next-intl/server";
import { Link } from "@/i18n/navigation";

export default async function Home() {
  const t = await getTranslations("home");

  return (
    <div className="flex flex-col items-center gap-20 py-16">
      {/* Hero */}
      <section className="text-center max-w-3xl">
        <div className="inline-flex items-center gap-2 rounded-full border border-primary/30 bg-primary/5 px-4 py-1.5 text-xs text-primary mb-8">
          <Shield className="h-3.5 w-3.5" />
          {t("badge")}
        </div>
        <h1 className="text-5xl md:text-7xl font-bold tracking-tight mb-6 leading-tight">
          <span className="text-primary">{t("headlineBurn")}</span> {t("headlineAfter")}<br />
          <span className="text-2xl md:text-3xl text-muted-foreground font-medium">{t("subheadline")}</span>
        </h1>
        <p className="text-lg text-muted-foreground mb-10 max-w-xl mx-auto">
          {t("description")}
        </p>
        <div className="flex flex-col sm:flex-row items-center justify-center gap-4">
          <Link
            href="/create"
            className="inline-flex items-center gap-2 rounded-lg bg-primary px-8 py-3.5 text-sm font-medium text-primary-foreground hover:bg-primary/90 transition-colors"
          >
            <Flame className="h-4 w-4" />
            {t("ctaShare")}
            <ArrowRight className="h-4 w-4" />
          </Link>
          <Link
            href="/docs"
            className="inline-flex items-center gap-2 rounded-lg border border-border px-8 py-3.5 text-sm font-medium hover:bg-muted transition-colors"
          >
            <Code className="h-4 w-4" />
            {t("ctaApi")}
          </Link>
        </div>
      </section>

      {/* Features */}
      <section className="grid grid-cols-1 md:grid-cols-3 gap-6 w-full max-w-4xl">
        <div className="rounded-xl border border-border bg-card p-6 hover:border-primary/30 transition-colors">
          <Shield className="h-8 w-8 text-primary mb-4" />
          <h3 className="font-semibold mb-2">{t("featZkTitle")}</h3>
          <p className="text-sm text-muted-foreground leading-relaxed">{t("featZkDesc")}</p>
        </div>
        <div className="rounded-xl border border-border bg-card p-6 hover:border-primary/30 transition-colors">
          <Flame className="h-8 w-8 text-primary mb-4" />
          <h3 className="font-semibold mb-2">{t("featBurnTitle")}</h3>
          <p className="text-sm text-muted-foreground leading-relaxed">{t("featBurnDesc")}</p>
        </div>
        <div className="rounded-xl border border-border bg-card p-6 hover:border-primary/30 transition-colors">
          <Eye className="h-8 w-8 text-primary mb-4" />
          <h3 className="font-semibold mb-2">{t("featAntiBotTitle")}</h3>
          <p className="text-sm text-muted-foreground leading-relaxed">{t("featAntiBotDesc")}</p>
        </div>
        <div className="rounded-xl border border-border bg-card p-6 hover:border-primary/30 transition-colors">
          <Terminal className="h-8 w-8 text-primary mb-4" />
          <h3 className="font-semibold mb-2">{t("featCliTitle")}</h3>
          <p className="text-sm text-muted-foreground leading-relaxed">{t("featCliDesc")}</p>
        </div>
        <div className="rounded-xl border border-border bg-card p-6 hover:border-primary/30 transition-colors">
          <Zap className="h-8 w-8 text-primary mb-4" />
          <h3 className="font-semibold mb-2">{t("featFileTitle")}</h3>
          <p className="text-sm text-muted-foreground leading-relaxed">{t("featFileDesc")}</p>
        </div>
        <div className="rounded-xl border border-border bg-card p-6 hover:border-primary/30 transition-colors">
          <Code className="h-8 w-8 text-primary mb-4" />
          <h3 className="font-semibold mb-2">{t("featOssTitle")}</h3>
          <p className="text-sm text-muted-foreground leading-relaxed">{t("featOssDesc")}</p>
        </div>
      </section>

      {/* How it works */}
      <section className="w-full max-w-3xl">
        <h2 className="text-2xl font-bold text-center mb-12">{t("howTitle")}</h2>
        <div className="space-y-8">
          <div className="flex gap-6 items-start">
            <div className="shrink-0 w-10 h-10 rounded-full bg-primary/10 text-primary flex items-center justify-center font-bold text-sm">1</div>
            <div>
              <h3 className="font-semibold mb-1">{t("howStep1Title")}</h3>
              <p className="text-sm text-muted-foreground">{t("howStep1Desc")}</p>
            </div>
          </div>
          <div className="flex gap-6 items-start">
            <div className="shrink-0 w-10 h-10 rounded-full bg-primary/10 text-primary flex items-center justify-center font-bold text-sm">2</div>
            <div>
              <h3 className="font-semibold mb-1">{t("howStep2Title")}</h3>
              <p className="text-sm text-muted-foreground">{t("howStep2Desc")}</p>
            </div>
          </div>
          <div className="flex gap-6 items-start">
            <div className="shrink-0 w-10 h-10 rounded-full bg-primary/10 text-primary flex items-center justify-center font-bold text-sm">3</div>
            <div>
              <h3 className="font-semibold mb-1">{t("howStep3Title")}</h3>
              <p className="text-sm text-muted-foreground">{t("howStep3Desc")}</p>
            </div>
          </div>
        </div>
      </section>

      {/* CLI example */}
      <section className="w-full max-w-3xl">
        <h2 className="text-2xl font-bold text-center mb-8">{t("cliTitle")}</h2>
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
        <h2 className="text-2xl font-bold text-center mb-8">{t("compTitle")}</h2>
        <div className="rounded-xl border border-border overflow-hidden">
          <table className="w-full text-sm">
            <thead>
              <tr className="bg-muted/30 border-b border-border">
                <th className="text-left px-6 py-3 font-medium">{t("compFeature")}</th>
                <th className="text-center px-4 py-3 font-medium text-primary">{t("compSharePwd")}</th>
                <th className="text-center px-4 py-3 font-medium text-muted-foreground">{t("compOthers")}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              <tr>
                <td className="px-6 py-3">{t("compZk")}</td>
                <td className="text-center px-4 py-3"><CheckCircle2 className="h-5 w-5 text-success mx-auto" /></td>
                <td className="text-center px-4 py-3"><XCircle className="h-5 w-5 text-muted-foreground/40 mx-auto" /></td>
              </tr>
              <tr>
                <td className="px-6 py-3">{t("compKey")}</td>
                <td className="text-center px-4 py-3"><CheckCircle2 className="h-5 w-5 text-success mx-auto" /></td>
                <td className="text-center px-4 py-3"><XCircle className="h-5 w-5 text-muted-foreground/40 mx-auto" /></td>
              </tr>
              <tr>
                <td className="px-6 py-3">{t("compAntiBot")}</td>
                <td className="text-center px-4 py-3"><CheckCircle2 className="h-5 w-5 text-success mx-auto" /></td>
                <td className="text-center px-4 py-3"><XCircle className="h-5 w-5 text-muted-foreground/40 mx-auto" /></td>
              </tr>
              <tr>
                <td className="px-6 py-3">{t("compCli")}</td>
                <td className="text-center px-4 py-3"><CheckCircle2 className="h-5 w-5 text-success mx-auto" /></td>
                <td className="text-center px-4 py-3"><CheckCircle2 className="h-5 w-5 text-success mx-auto" /></td>
              </tr>
              <tr>
                <td className="px-6 py-3">{t("compFile")}</td>
                <td className="text-center px-4 py-3"><CheckCircle2 className="h-5 w-5 text-success mx-auto" /></td>
                <td className="text-center px-4 py-3"><CheckCircle2 className="h-5 w-5 text-success mx-auto" /></td>
              </tr>
              <tr>
                <td className="px-6 py-3">{t("compOss")}</td>
                <td className="text-center px-4 py-3"><CheckCircle2 className="h-5 w-5 text-success mx-auto" /></td>
                <td className="text-center px-4 py-3"><CheckCircle2 className="h-5 w-5 text-success mx-auto" /></td>
              </tr>
              <tr>
                <td className="px-6 py-3">{t("compSelfHost")}</td>
                <td className="text-center px-4 py-3"><CheckCircle2 className="h-5 w-5 text-success mx-auto" /></td>
                <td className="text-center px-4 py-3"><CheckCircle2 className="h-5 w-5 text-success mx-auto" /></td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      {/* Sovereignty */}
      <section className="w-full max-w-3xl">
        <div className="rounded-xl border border-border bg-card p-8 text-center">
          <div className="inline-flex items-center gap-2 rounded-full border border-primary/30 bg-primary/5 px-4 py-1.5 text-xs text-primary mb-6">
            <MapPin className="h-3.5 w-3.5" />
            {t("sovereigntyBadge")}
          </div>
          <h2 className="text-2xl font-bold mb-4">{t("sovereigntyTitle")}</h2>
          <p className="text-sm text-muted-foreground leading-relaxed max-w-2xl mx-auto">
            {t("sovereigntyDesc")}
          </p>
        </div>
      </section>

      {/* CTA */}
      <section className="text-center">
        <Link
          href="/create"
          className="inline-flex items-center gap-2 rounded-lg bg-primary px-8 py-3.5 text-sm font-medium text-primary-foreground hover:bg-primary/90 transition-colors"
        >
          <Flame className="h-4 w-4" />
          {t("ctaStart")}
        </Link>
      </section>
    </div>
  );
}
