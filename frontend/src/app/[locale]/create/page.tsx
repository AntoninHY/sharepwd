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
