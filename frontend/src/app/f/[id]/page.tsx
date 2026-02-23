import FileDownload from "@/components/secret/file-download";

interface PageProps {
  params: Promise<{ id: string }>;
}

export default async function FilePage({ params }: PageProps) {
  const { id } = await params;
  return <FileDownload token={id} />;
}
