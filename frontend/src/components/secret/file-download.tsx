"use client";

import { useState, useEffect, useRef } from "react";
import { toast } from "sonner";
import { useTranslations } from "next-intl";
import { Download, Lock, AlertTriangle, Clock, Flame, Eye } from "lucide-react";
import { decryptText, decryptWithPassphrase, fromBase64 } from "@/lib/crypto";
import { api, type SecretMetadata, type RevealSecretResponse } from "@/lib/api";

interface FileDownloadProps {
  token: string;
}

export default function FileDownload({ token }: FileDownloadProps) {
  const t = useTranslations("fileDownload");

  const [metadata, setMetadata] = useState<SecretMetadata | null>(null);
  const [passphrase, setPassphrase] = useState("");
  const [loading, setLoading] = useState(true);
  const [downloading, setDownloading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [passphraseError, setPassphraseError] = useState<string | null>(null);
  const [hasInteracted, setHasInteracted] = useState(false);
  const [revealedData, setRevealedData] = useState<RevealSecretResponse | null>(null);
  const pageLoadTime = useRef(Date.now());

  useEffect(() => {
    const handler = () => setHasInteracted(true);
    window.addEventListener("pointermove", handler, { once: true });
    window.addEventListener("pointerdown", handler, { once: true });
    return () => {
      window.removeEventListener("pointermove", handler);
      window.removeEventListener("pointerdown", handler);
    };
  }, []);

  useEffect(() => {
    async function fetchMetadata() {
      try {
        const meta = await api.getSecretMetadata(token);
        if (meta.is_expired) {
          setError(t("errorExpired"));
          return;
        }
        setMetadata(meta);
      } catch {
        setError(t("errorNotFound"));
      } finally {
        setLoading(false);
      }
    }
    fetchMetadata();
  }, [token, t]);

  const isDecryptionError = (err: unknown): boolean => {
    if (!(err instanceof Error)) return false;
    const msg = err.message.toLowerCase();
    return msg.includes("decrypt") || msg.includes("operation") || msg.includes("tag") || msg.includes("gcm");
  };

  const handleDownload = async () => {
    if (!metadata) return;

    const timeOnPage = Date.now() - pageLoadTime.current;
    if (timeOnPage < 500 || !hasInteracted) {
      setError(t("errorWait"));
      return;
    }

    if (metadata.has_passphrase && !passphrase) {
      toast.error(t("toastPassphraseRequired"));
      return;
    }

    setDownloading(true);
    setPassphraseError(null);
    try {
      const revealed = revealedData || await api.revealSecret(token, metadata.challenge_nonce);
      if (!revealedData) setRevealedData(revealed);

      const keyFragment = window.location.hash.slice(1);

      let decryptedPayload: string;
      if (metadata.has_passphrase) {
        if (!revealed.salt) throw new Error("Missing salt");
        decryptedPayload = await decryptWithPassphrase(revealed.encrypted_data, revealed.iv, revealed.salt, passphrase);
      } else {
        if (!keyFragment) throw new Error("Missing encryption key in URL");
        decryptedPayload = await decryptText(revealed.encrypted_data, revealed.iv, keyFragment);
      }

      let fileName = "downloaded-file";
      let fileBytes: Uint8Array;

      try {
        const parsed = JSON.parse(decryptedPayload);
        if (parsed.name && parsed.data) {
          fileName = parsed.name;
          fileBytes = fromBase64(parsed.data);
        } else {
          fileBytes = fromBase64(decryptedPayload);
        }
      } catch {
        fileBytes = fromBase64(decryptedPayload);
      }

      const blob = new Blob([fileBytes as BlobPart]);
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = fileName;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);

      toast.success(t("toastSuccess"));
    } catch (err) {
      if (isDecryptionError(err) && metadata.has_passphrase) {
        setPassphraseError(t("errorWrongPassphrase"));
        setPassphrase("");
      } else {
        const msg = err instanceof Error ? err.message : "";
        if (msg.includes("expired") || msg.includes("Gone")) {
          setError(t("errorExpired"));
        } else if (isDecryptionError(err)) {
          setError(t("errorDecrypt"));
        } else {
          setError(msg || t("errorGeneric"));
        }
      }
    } finally {
      setDownloading(false);
    }
  };

  if (loading) {
    return <div className="flex items-center justify-center py-16"><div className="animate-pulse text-muted-foreground">{t("loading")}</div></div>;
  }

  if (error) {
    return (
      <div className="max-w-lg mx-auto">
        <div className="rounded-xl border border-destructive/50 bg-card p-6 text-center">
          <AlertTriangle className="h-12 w-12 text-destructive mx-auto mb-4" />
          <h2 className="text-lg font-semibold mb-2">{t("unavailableTitle")}</h2>
          <p className="text-sm text-muted-foreground">{error}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-lg mx-auto">
      <div className="rounded-xl border border-border bg-card p-6">
        <h2 className="text-lg font-semibold mb-2 flex items-center gap-2">
          <Lock className="h-5 w-5 text-primary" />
          {t("sharedTitle")}
        </h2>
        <div className="mt-4 space-y-3 text-sm text-muted-foreground">
          {metadata?.burn_after_read && (
            <p className="flex items-center gap-2 text-warning"><Flame className="h-4 w-4" /> {t("burnNotice")}</p>
          )}
          {metadata?.max_views && (
            <p className="flex items-center gap-2"><Eye className="h-4 w-4" /> {t("viewsUsed", { current: metadata.current_views, max: metadata.max_views })}</p>
          )}
          {metadata?.expires_at && (
            <p className="flex items-center gap-2"><Clock className="h-4 w-4" /> {t("expires", { date: new Date(metadata.expires_at).toLocaleString() })}</p>
          )}
        </div>
        {metadata?.has_passphrase && (
          <div className="mt-4">
            <label htmlFor="file-passphrase" className="block text-sm font-medium mb-2">{t("enterPassphrase")}</label>
            <input
              id="file-passphrase"
              type="password"
              value={passphrase}
              onChange={(e) => { setPassphrase(e.target.value); setPassphraseError(null); }}
              placeholder={t("passphrasePlaceholder")}
              className={`w-full rounded-lg border px-4 py-3 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary/50 ${
                passphraseError ? "border-destructive bg-destructive/5" : "border-border bg-background"
              }`}
              onKeyDown={(e) => e.key === "Enter" && handleDownload()}
            />
            {passphraseError && (
              <p className="mt-2 text-sm text-destructive flex items-center gap-1.5">
                <AlertTriangle className="h-3.5 w-3.5" />
                {passphraseError}
              </p>
            )}
          </div>
        )}
        <button
          onClick={handleDownload}
          disabled={downloading || (metadata?.has_passphrase && !passphrase)}
          className="mt-6 w-full rounded-lg bg-primary px-6 py-3 text-sm font-medium text-primary-foreground hover:bg-primary/90 transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
        >
          {downloading ? <span className="animate-pulse">{t("decrypting")}</span> : <><Download className="h-4 w-4" /> {t("downloadButton")}</>}
        </button>
      </div>
    </div>
  );
}
