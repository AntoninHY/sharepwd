// SharePwd.io — Burn After Reading
// Zero-knowledge secret sharing. Client-side AES-256-GCM encryption, secrets self-destruct after viewing.
// Copyright (c) 2025-2026 Antonin HILY — CTO, Jizo AI

"use client";

import { Dices, Shuffle } from "lucide-react";
import { useTranslations } from "next-intl";
import { generateDicewarePassphrase, generateRandomPassphrase } from "@/lib/passphrase";

interface PassphraseGeneratorProps {
  onGenerate: (passphrase: string) => void;
  translationNamespace: string;
}

export default function PassphraseGenerator({ onGenerate, translationNamespace }: PassphraseGeneratorProps) {
  const t = useTranslations(translationNamespace);

  const handleWords = () => {
    onGenerate(generateDicewarePassphrase());
  };

  const handleRandom = () => {
    onGenerate(generateRandomPassphrase());
  };

  return (
    <div className="flex items-center gap-2 mt-1.5">
      <span className="text-xs text-muted-foreground">{t("generate")}:</span>
      <button
        type="button"
        onClick={handleWords}
        className="inline-flex items-center gap-1 rounded-md border border-border px-2 py-1 text-xs text-muted-foreground hover:bg-muted hover:text-foreground transition-colors"
      >
        <Dices className="h-3 w-3" />
        {t("generateWords")}
      </button>
      <button
        type="button"
        onClick={handleRandom}
        className="inline-flex items-center gap-1 rounded-md border border-border px-2 py-1 text-xs text-muted-foreground hover:bg-muted hover:text-foreground transition-colors"
      >
        <Shuffle className="h-3 w-3" />
        {t("generateRandom")}
      </button>
    </div>
  );
}
