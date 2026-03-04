// SharePwd.io — Burn After Reading
// Zero-knowledge secret sharing. Client-side AES-256-GCM encryption, secrets self-destruct after viewing.
// Copyright (c) 2025-2026 Antonin HILY — CTO, Jizo AI

import { z } from "zod";

export const createSecretSchema = z.object({
  content: z.string().min(1, "Content is required").max(100_000, "Content too large (max 100KB)"),
  passphrase: z.string().optional(),
  expiresIn: z.string().optional(),
  maxViews: z.number().int().positive().optional(),
  burnAfterRead: z.boolean().default(false),
});

export type CreateSecretInput = z.infer<typeof createSecretSchema>;
