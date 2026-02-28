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
