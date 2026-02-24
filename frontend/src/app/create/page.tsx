import type { Metadata } from "next";
import CreatePageClient from "./client";

export const metadata: Metadata = {
  title: "Share a Secret",
  description:
    "Encrypt and share passwords, API keys, and secrets with a self-destructing link. Zero-knowledge AES-256-GCM encryption in your browser.",
  alternates: {
    canonical: "/create",
  },
};

export default function CreatePage() {
  return <CreatePageClient />;
}
