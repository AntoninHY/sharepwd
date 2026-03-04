// SharePwd.io — Burn After Reading
// Zero-knowledge secret sharing. Client-side AES-256-GCM encryption, secrets self-destruct after viewing.
// Copyright (c) 2025-2026 Antonin HILY — CTO, Jizo AI

export type ExpirationOption = {
  value: string;
};

export const EXPIRATION_OPTIONS: ExpirationOption[] = [
  { value: "5m" },
  { value: "30m" },
  { value: "1h" },
  { value: "4h" },
  { value: "24h" },
  { value: "168h" },
  { value: "720h" },
];

export const VIEW_OPTIONS = [1, 2, 3, 5, 10, 25, 50, 100] as const;
