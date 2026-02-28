"use client";

import { useState } from "react";
import { FileText, Upload } from "lucide-react";
import { useTranslations } from "next-intl";
import CreateForm from "@/components/secret/create-form";
import FileUploadForm from "@/components/secret/file-upload-form";

export default function CreatePageClient() {
  const [tab, setTab] = useState<"text" | "file">("text");
  const t = useTranslations("create");

  return (
    <div className="max-w-2xl mx-auto">
      <h1 className="text-2xl font-bold mb-2">{t("title")}</h1>
      <p className="text-muted-foreground mb-6">{t("subtitle")}</p>

      <div className="flex gap-1 mb-8 rounded-lg bg-muted p-1">
        <button
          onClick={() => setTab("text")}
          className={`flex-1 flex items-center justify-center gap-2 rounded-md px-4 py-2.5 text-sm font-medium transition-colors ${
            tab === "text" ? "bg-card text-foreground shadow-sm" : "text-muted-foreground hover:text-foreground"
          }`}
        >
          <FileText className="h-4 w-4" /> {t("tabText")}
        </button>
        <button
          onClick={() => setTab("file")}
          className={`flex-1 flex items-center justify-center gap-2 rounded-md px-4 py-2.5 text-sm font-medium transition-colors ${
            tab === "file" ? "bg-card text-foreground shadow-sm" : "text-muted-foreground hover:text-foreground"
          }`}
        >
          <Upload className="h-4 w-4" /> {t("tabFile")}
        </button>
      </div>

      {tab === "text" ? <CreateForm /> : <FileUploadForm />}
    </div>
  );
}
