// SharePwd.io — Burn After Reading
// Zero-knowledge secret sharing. Client-side AES-256-GCM encryption, secrets self-destruct after viewing.
// Copyright (c) 2025-2026 Antonin HILY — CTO, Jizo AI

import type { Metadata } from "next";
import FileDownload from "@/components/secret/file-download";

export const metadata: Metadata = {
  robots: { index: false, follow: false },
};

interface PageProps {
  params: Promise<{ id: string }>;
}

export default async function FilePage({ params }: PageProps) {
  const { id } = await params;
  return <FileDownload token={id} />;
}
