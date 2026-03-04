// SharePwd.io — Burn After Reading
// Zero-knowledge secret sharing. Client-side AES-256-GCM encryption, secrets self-destruct after viewing.
// Copyright (c) 2025-2026 Antonin HILY — CTO, Jizo AI

import createMiddleware from "next-intl/middleware";
import { routing } from "./i18n/navigation";

export default createMiddleware(routing);

export const config = {
  matcher: ["/((?!_next|v1|analytics|.*\\..*).*)"],
};
