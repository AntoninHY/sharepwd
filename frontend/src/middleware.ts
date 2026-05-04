// SharePwd.io — Burn After Reading
// Zero-knowledge secret sharing. Client-side AES-256-GCM encryption, secrets self-destruct after viewing.
// Copyright (c) 2025-2026 Antonin HILY — CTO, Jizo AI

import createMiddleware from "next-intl/middleware";
import type { NextRequest } from "next/server";
import { routing } from "./i18n/navigation";

const intlMiddleware = createMiddleware(routing);

/**
 * Trust X-Forwarded-Host and X-Forwarded-Proto when generating redirects.
 *
 * When SharePwd is deployed behind a reverse proxy that terminates TLS
 * (nginx, Caddy, Traefik, etc.) on a port different from Next.js's listen
 * port, the next-intl locale middleware constructs absolute redirect URLs
 * using Next.js's internal listen port (e.g. 3000) rather than the public
 * port. The result is unreachable redirects like
 * https://example.com:3000/<locale>/path.
 *
 * This wrapper inspects the response: if next-intl returns a 3xx redirect
 * with an absolute Location header, we rewrite the host and protocol to
 * match the X-Forwarded-Host and X-Forwarded-Proto headers from the proxy.
 *
 * If those headers are absent (no proxy in front), behaviour is unchanged.
 */
export default function middleware(request: NextRequest) {
  const response = intlMiddleware(request);

  if (response.status >= 300 && response.status < 400) {
    const location = response.headers.get("location");
    if (location && /^https?:\/\//.test(location)) {
      const forwardedHost = request.headers.get("x-forwarded-host");
      const forwardedProto = request.headers.get("x-forwarded-proto");

      if (forwardedHost || forwardedProto) {
        try {
          const url = new URL(location);
          if (forwardedProto) {
            url.protocol = `${forwardedProto}:`;
          }
          if (forwardedHost) {
            // X-Forwarded-Host may include a non-default port (e.g. ":8080")
            url.host = forwardedHost;
          }
          response.headers.set("location", url.toString());
        } catch {
          // Invalid URL in Location header: leave it untouched
        }
      }
    }
  }

  return response;
}

export const config = {
  matcher: ["/((?!_next|v1|analytics|.*\\..*).*)"],
};
