// SharePwd.io — Burn After Reading
// Zero-knowledge secret sharing. Client-side AES-256-GCM encryption, secrets self-destruct after viewing.
// Copyright (c) 2025-2026 Antonin HILY — CTO, Jizo AI

import type { Metadata } from "next";
import { NextIntlClientProvider } from "next-intl";
import { getMessages, getTranslations } from "next-intl/server";
import { notFound } from "next/navigation";
import { Toaster } from "sonner";
import { routing } from "@/i18n/navigation";
import { Link } from "@/i18n/navigation";
import LanguageSwitcher from "@/components/language-switcher";
import "../globals.css";

const SITE_URL = "https://sharepwd.io";

export async function generateMetadata({
  params,
}: {
  params: Promise<{ locale: string }>;
}): Promise<Metadata> {
  const { locale } = await params;
  const t = await getTranslations({ locale, namespace: "metadata" });

  return {
    metadataBase: new URL(SITE_URL),
    title: {
      default: t("title"),
      template: "%s — SharePwd",
    },
    description: t("description"),
    keywords: [
      "password sharing",
      "secret sharing",
      "burn after reading",
      "zero-knowledge encryption",
      "self-destructing links",
      "secure file sharing",
      "AES-256-GCM",
      "encrypted secrets",
      "one-time secret",
      "sharepwd",
    ],
    authors: [{ name: "Jizo AI", url: "https://jizo.ai" }],
    creator: "Jizo AI",
    publisher: "Jizo AI",
    openGraph: {
      type: "website",
      locale: locale === "fr" ? "fr_FR" : "en_US",
      url: SITE_URL,
      siteName: "SharePwd",
      title: t("title"),
      description: t("description"),
      images: [
        {
          url: "/og.png",
          width: 1200,
          height: 630,
          alt: "SharePwd — Burn After Reading",
        },
      ],
    },
    twitter: {
      card: "summary_large_image",
      title: t("title"),
      description: t("description"),
      images: ["/og.png"],
    },
    robots: {
      index: true,
      follow: true,
    },
    alternates: {
      canonical: `${SITE_URL}/${locale}`,
      languages: {
        en: `${SITE_URL}/en`,
        fr: `${SITE_URL}/fr`,
      },
    },
  };
}

const jsonLd = [
  {
    "@context": "https://schema.org",
    "@type": "Organization",
    name: "Jizo AI",
    url: "https://jizo.ai",
    logo: "https://jizo.ai/logo.png",
  },
  {
    "@context": "https://schema.org",
    "@type": "WebApplication",
    name: "SharePwd",
    url: SITE_URL,
    description:
      "Share passwords and secrets that self-destruct after reading. Zero-knowledge AES-256-GCM encryption in your browser.",
    applicationCategory: "SecurityApplication",
    operatingSystem: "Any",
    offers: {
      "@type": "Offer",
      price: "0",
      priceCurrency: "USD",
    },
    author: {
      "@type": "Organization",
      name: "Jizo AI",
      url: "https://jizo.ai",
    },
  },
];

export default async function LocaleLayout({
  children,
  params,
}: {
  children: React.ReactNode;
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;

  if (!routing.locales.includes(locale as "en" | "fr")) {
    notFound();
  }

  const messages = await getMessages();
  const t = await getTranslations("nav");
  const tFooter = await getTranslations("footer");
  const umamiWebsiteId = process.env.NEXT_PUBLIC_UMAMI_WEBSITE_ID;

  return (
    <html lang={locale} className="dark">
      <head>
        <script
          type="application/ld+json"
          dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }}
        />
      </head>
      <body className="min-h-screen flex flex-col bg-background antialiased">
        {umamiWebsiteId && (
          <script
            src="/analytics/script.js"
            data-website-id={umamiWebsiteId}
            data-domains="sharepwd.io"
            defer
          />
        )}
        <NextIntlClientProvider messages={messages}>
          <header className="border-b border-border">
            <div className="mx-auto max-w-5xl px-4 py-4 flex items-center justify-between">
              <Link href="/" className="flex items-center gap-3">
                <span className="text-xl font-bold text-primary">SharePwd</span>
                <span className="hidden sm:inline text-xs text-muted-foreground border-l border-border pl-3 uppercase tracking-widest">
                  {t("tagline")}
                </span>
              </Link>
              <nav className="flex items-center gap-6 text-sm text-muted-foreground">
                <Link href="/create" className="hover:text-foreground transition-colors">
                  {t("shareSecret")}
                </Link>
                <Link href="/docs" className="hover:text-foreground transition-colors">
                  {t("apiDocs")}
                </Link>
                <Link href="/docs/cli" className="hover:text-foreground transition-colors">
                  {t("cliDocs")}
                </Link>
                <LanguageSwitcher />
              </nav>
            </div>
          </header>
          <main className="mx-auto max-w-5xl px-4 py-8 flex-1 w-full">
            {children}
          </main>
          <footer className="border-t border-border mt-auto">
            <div className="mx-auto max-w-5xl px-4 py-6 flex flex-col sm:flex-row items-center justify-between gap-4 text-sm text-muted-foreground">
              <div className="flex items-center gap-2">
                <span className="font-semibold text-foreground">SharePwd</span>
                <span className="text-muted-foreground/60">&copy; 2026</span>
              </div>
              <nav className="flex items-center gap-6">
                <Link href="/create" className="hover:text-foreground transition-colors">
                  {t("shareSecret")}
                </Link>
                <Link href="/docs" className="hover:text-foreground transition-colors">
                  {t("apiDocs")}
                </Link>
                <Link href="/docs/cli" className="hover:text-foreground transition-colors">
                  {t("cliDocs")}
                </Link>
                <a
                  href="https://github.com/AntoninHY/sharepwd"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="hover:text-foreground transition-colors"
                  title="GitHub"
                >
                  <svg className="h-5 w-5" fill="currentColor" viewBox="0 0 24 24" aria-hidden="true">
                    <path fillRule="evenodd" d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z" clipRule="evenodd" />
                  </svg>
                </a>
              </nav>
              <span>
                {tFooter("builtBy")}{" "}
                <a
                  href="https://jizo.ai"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-foreground hover:text-primary transition-colors"
                >
                  Jizo AI
                </a>
              </span>
            </div>
            <div className="mx-auto max-w-5xl px-4 pb-4 text-center text-xs text-muted-foreground/50">
              {tFooter("zeroKnowledge")}
            </div>
            <div className="mx-auto max-w-5xl px-4 pb-6 text-center text-xs text-muted-foreground/50">
              {tFooter("hosting")}
            </div>
          </footer>
        </NextIntlClientProvider>
        <Toaster theme="dark" position="bottom-right" />
      </body>
    </html>
  );
}
