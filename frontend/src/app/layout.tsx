import type { Metadata } from "next";

import { Toaster } from "sonner";
import "./globals.css";

const SITE_URL = "https://sharepwd.io";

export const metadata: Metadata = {
  metadataBase: new URL(SITE_URL),
  title: {
    default: "SharePwd — Burn After Reading",
    template: "%s — SharePwd",
  },
  description:
    "Share passwords and secrets that self-destruct after reading. Zero-knowledge encryption, built by a cybersecurity team. Your data never touches our servers in plaintext.",
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
  authors: [{ name: "Jizô AI", url: "https://jizo.ai" }],
  creator: "Jizô AI",
  publisher: "Jizô AI",
  openGraph: {
    type: "website",
    locale: "en_US",
    url: SITE_URL,
    siteName: "SharePwd",
    title: "SharePwd — Burn After Reading",
    description:
      "Share passwords and secrets that self-destruct after reading. Zero-knowledge encryption, built by a cybersecurity team.",
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
    title: "SharePwd — Burn After Reading",
    description:
      "Share passwords and secrets that self-destruct after reading. Zero-knowledge encryption.",
    images: ["/og.png"],
  },
  robots: {
    index: true,
    follow: true,
  },
  alternates: {
    canonical: SITE_URL,
  },
};

const jsonLd = [
  {
    "@context": "https://schema.org",
    "@type": "Organization",
    name: "Jizô AI",
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
      name: "Jizô AI",
      url: "https://jizo.ai",
    },
  },
];

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const umamiWebsiteId = process.env.NEXT_PUBLIC_UMAMI_WEBSITE_ID;

  return (
    <html lang="en" className="dark">
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
        <header className="border-b border-border">
          <div className="mx-auto max-w-5xl px-4 py-4 flex items-center justify-between">
            <a href="/" className="flex items-center gap-3">
              <span className="text-xl font-bold text-primary">SharePwd</span>
              <span className="hidden sm:inline text-xs text-muted-foreground border-l border-border pl-3 uppercase tracking-widest">Burn After Reading</span>
            </a>
            <nav className="flex items-center gap-6 text-sm text-muted-foreground">
              <a href="/create" className="hover:text-foreground transition-colors">
                Share a Secret
              </a>
              <a href="/docs" className="hover:text-foreground transition-colors">
                API Docs
              </a>
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
              <span className="text-muted-foreground/60">© 2026</span>
            </div>
            <nav className="flex items-center gap-6">
              <a href="/create" className="hover:text-foreground transition-colors">Share a Secret</a>
              <a href="/docs" className="hover:text-foreground transition-colors">API Docs</a>
            </nav>
            <span>
              Built by{" "}
              <a href="https://jizo.ai" target="_blank" rel="noopener noreferrer" className="text-foreground hover:text-primary transition-colors">
                Jizô AI
              </a>
            </span>
          </div>
          <div className="mx-auto max-w-5xl px-4 pb-6 text-center text-xs text-muted-foreground/50">
            Zero-knowledge encryption — your data never touches our servers in plaintext.
          </div>
        </footer>
        <Toaster theme="dark" position="bottom-right" />
      </body>
    </html>
  );
}
