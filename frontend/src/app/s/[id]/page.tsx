import RevealGate from "@/components/secret/reveal-gate";

interface PageProps {
  params: Promise<{ id: string }>;
}

export default async function SecretPage({ params }: PageProps) {
  const { id } = await params;
  return <RevealGate token={id} />;
}
