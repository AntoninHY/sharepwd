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
