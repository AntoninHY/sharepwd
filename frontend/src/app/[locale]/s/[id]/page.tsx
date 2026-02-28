import type { Metadata } from "next";
import RevealGate from "@/components/secret/reveal-gate";

export const metadata: Metadata = {
  robots: { index: false, follow: false },
};

interface PageProps {
  params: Promise<{ id: string }>;
}

export default async function SecretPage({ params }: PageProps) {
  const { id } = await params;
  return <RevealGate token={id} />;
}
