import type { Metadata } from "next";
import { Toaster } from "sonner";
import "./globals.css";

export const metadata: Metadata = {
  title: "SharePwd - Secure Secret Sharing",
  description: "Share passwords and secrets securely with zero-knowledge encryption. Your data never touches our servers in plaintext.",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" className="dark">
      <body className="min-h-screen bg-background antialiased">
        <header className="border-b border-border">
          <div className="mx-auto max-w-5xl px-4 py-4 flex items-center justify-between">
            <a href="/" className="text-xl font-bold text-primary">
              SharePwd
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
        <main className="mx-auto max-w-5xl px-4 py-8">
          {children}
        </main>
        <Toaster theme="dark" position="bottom-right" />
      </body>
    </html>
  );
}
