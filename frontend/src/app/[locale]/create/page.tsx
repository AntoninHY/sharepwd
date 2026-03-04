// SharePwd.io — Burn After Reading
// Zero-knowledge secret sharing. Client-side AES-256-GCM encryption, secrets self-destruct after viewing.
// Copyright (c) 2025-2026 Antonin HILY — CTO, Jizo AI

import type { Metadata } from "next";
import { getTranslations } from "next-intl/server";
import CreatePageClient from "./client";

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations("create");
  return {
    title: t("title"),
    description: t("subtitle"),
  };
}

export default function CreatePage() {
  return <CreatePageClient />;
}
