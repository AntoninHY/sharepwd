"use client";

import { useState, useEffect, useRef } from "react";
import { toast } from "sonner";
import { Download, Lock, AlertTriangle, Clock, Flame, Eye } from "lucide-react";
import { decryptText, decryptWithPassphrase, fromBase64 } from "@/lib/crypto";
import { api, type SecretMetadata, type RevealSecretResponse } from "@/lib/api";

interface FileDownloadProps {
  token: string;
}

export default function FileDownload({ token }: FileDownloadProps) {
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
          setError("This file has expired or been deleted.");
          return;
        }
        setMetadata(meta);
      } catch {
        setError("File not found or has expired.");
      } finally {
        setLoading(false);
      }
    }
    fetchMetadata();
  }, [token]);

  const isDecryptionError = (err: unknown): boolean => {
    if (!(err instanceof Error)) return false;
    const msg = err.message.toLowerCase();
    return msg.includes("decrypt") || msg.includes("operation") || msg.includes("tag") || msg.includes("gcm");
  };

  const handleDownload = async () => {
    if (!metadata) return;

    const timeOnPage = Date.now() - pageLoadTime.current;
    if (timeOnPage < 500 || !hasInteracted) {
      setError("Please wait a moment and interact with the page first.");
      return;
    }

    if (metadata.has_passphrase && !passphrase) {
      toast.error("Please enter the passphrase");
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

      toast.success("File downloaded and decrypted");
    } catch (err) {
      if (isDecryptionError(err) && metadata.has_passphrase) {
        setPassphraseError("Wrong passphrase. Please try again.");
        setPassphrase("");
      } else {
        const msg = err instanceof Error ? err.message : "Failed to download file";
        if (msg.includes("expired") || msg.includes("Gone")) {
          setError("This file has expired or been deleted.");
        } else if (isDecryptionError(err)) {
          setError("Failed to decrypt. The link may be corrupted.");
        } else {
          setError(msg);
        }
      }
    } finally {
      setDownloading(false);
    }
  };

  if (loading) {
    return <div className="flex items-center justify-center py-16"><div className="animate-pulse text-muted-foreground">Loading...</div></div>;
  }

  if (error) {
    return (
      <div className="max-w-lg mx-auto">
        <div className="rounded-xl border border-destructive/50 bg-card p-6 text-center">
          <AlertTriangle className="h-12 w-12 text-destructive mx-auto mb-4" />
          <h2 className="text-lg font-semibold mb-2">File Unavailable</h2>
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
          Someone shared a file with you
        </h2>
        <div className="mt-4 space-y-3 text-sm text-muted-foreground">
          {metadata?.burn_after_read && (
            <p className="flex items-center gap-2 text-warning"><Flame className="h-4 w-4" /> This file will be destroyed after download</p>
          )}
          {metadata?.max_views && (
            <p className="flex items-center gap-2"><Eye className="h-4 w-4" /> {metadata.current_views} of {metadata.max_views} views used</p>
          )}
          {metadata?.expires_at && (
            <p className="flex items-center gap-2"><Clock className="h-4 w-4" /> Expires: {new Date(metadata.expires_at).toLocaleString()}</p>
          )}
        </div>
        {metadata?.has_passphrase && (
          <div className="mt-4">
            <label htmlFor="file-passphrase" className="block text-sm font-medium mb-2">Enter Passphrase</label>
            <input
              id="file-passphrase"
              type="password"
              value={passphrase}
              onChange={(e) => { setPassphrase(e.target.value); setPassphraseError(null); }}
              placeholder="Enter the passphrase"
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
          {downloading ? <span className="animate-pulse">Decrypting...</span> : <><Download className="h-4 w-4" /> Download & Decrypt</>}
        </button>
      </div>
    </div>
  );
}
