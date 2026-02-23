"use client";

import { useState, useEffect, useRef } from "react";
import { toast } from "sonner";
import { Eye, Lock, AlertTriangle, Copy, Clock, Flame } from "lucide-react";
import { decryptText, decryptWithPassphrase } from "@/lib/crypto";
import { api, type SecretMetadata, type RevealSecretResponse } from "@/lib/api";

interface RevealGateProps {
  token: string;
}

export default function RevealGate({ token }: RevealGateProps) {
  const [metadata, setMetadata] = useState<SecretMetadata | null>(null);
  const [decryptedContent, setDecryptedContent] = useState<string | null>(null);
  const [passphrase, setPassphrase] = useState("");
  const [loading, setLoading] = useState(true);
  const [revealing, setRevealing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [hasInteracted, setHasInteracted] = useState(false);
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
          setError("This secret has expired or been deleted.");
          return;
        }
        setMetadata(meta);
      } catch (err) {
        setError("Secret not found or has expired.");
      } finally {
        setLoading(false);
      }
    }
    fetchMetadata();
  }, [token]);

  const handleReveal = async () => {
    if (!metadata) return;

    const timeOnPage = Date.now() - pageLoadTime.current;
    if (timeOnPage < 500) {
      setError("Please wait a moment before revealing the secret.");
      return;
    }

    if (!hasInteracted) {
      setError("Please move your mouse or interact with the page first.");
      return;
    }

    if (metadata.has_passphrase && !passphrase) {
      toast.error("Please enter the passphrase");
      return;
    }

    setRevealing(true);
    try {
      const revealed = await api.revealSecret(token, metadata.challenge_nonce);

      const keyFragment = window.location.hash.slice(1);

      let plaintext: string;
      if (metadata.has_passphrase) {
        if (!revealed.salt) throw new Error("Missing salt for passphrase decryption");
        plaintext = await decryptWithPassphrase(
          revealed.encrypted_data,
          revealed.iv,
          revealed.salt,
          passphrase
        );
      } else {
        if (!keyFragment) throw new Error("Missing encryption key in URL");
        plaintext = await decryptText(revealed.encrypted_data, revealed.iv, keyFragment);
      }

      setDecryptedContent(plaintext);
    } catch (err) {
      const msg = err instanceof Error ? err.message : "Failed to reveal secret";
      if (msg.includes("expired") || msg.includes("Gone")) {
        setError("This secret has expired or been deleted.");
      } else if (msg.includes("Decryption") || msg.includes("operation")) {
        setError("Failed to decrypt. Wrong passphrase or corrupted data.");
      } else {
        setError(msg);
      }
    } finally {
      setRevealing(false);
    }
  };

  const copyContent = async () => {
    if (!decryptedContent) return;
    await navigator.clipboard.writeText(decryptedContent);
    toast.success("Copied to clipboard");
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center py-16">
        <div className="animate-pulse text-muted-foreground">Loading...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="max-w-lg mx-auto">
        <div className="rounded-xl border border-destructive/50 bg-card p-6 text-center">
          <AlertTriangle className="h-12 w-12 text-destructive mx-auto mb-4" />
          <h2 className="text-lg font-semibold mb-2">Secret Unavailable</h2>
          <p className="text-sm text-muted-foreground">{error}</p>
        </div>
      </div>
    );
  }

  if (decryptedContent !== null) {
    return (
      <div className="max-w-2xl mx-auto space-y-4">
        <div className="rounded-xl border border-success/50 bg-card p-6">
          <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
            <Eye className="h-5 w-5 text-success" />
            Secret Revealed
          </h2>
          {metadata?.burn_after_read && (
            <p className="text-sm text-warning mb-4 flex items-center gap-2">
              <Flame className="h-4 w-4" />
              This secret has been destroyed and cannot be viewed again.
            </p>
          )}
          <div className="relative">
            <pre className="rounded-lg bg-muted px-4 py-3 text-sm font-mono whitespace-pre-wrap break-all max-h-[400px] overflow-y-auto">
              {decryptedContent}
            </pre>
            <button
              onClick={copyContent}
              className="absolute top-2 right-2 rounded-md bg-secondary p-2 hover:bg-secondary/80 transition-colors"
              title="Copy to clipboard"
            >
              <Copy className="h-4 w-4" />
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-lg mx-auto">
      <div className="rounded-xl border border-border bg-card p-6">
        <h2 className="text-lg font-semibold mb-2 flex items-center gap-2">
          <Lock className="h-5 w-5 text-primary" />
          Someone shared a secret with you
        </h2>

        <div className="mt-4 space-y-3 text-sm text-muted-foreground">
          {metadata?.burn_after_read && (
            <p className="flex items-center gap-2 text-warning">
              <Flame className="h-4 w-4" />
              This secret will be destroyed after viewing
            </p>
          )}
          {metadata?.max_views && (
            <p className="flex items-center gap-2">
              <Eye className="h-4 w-4" />
              {metadata.current_views} of {metadata.max_views} views used
            </p>
          )}
          {metadata?.expires_at && (
            <p className="flex items-center gap-2">
              <Clock className="h-4 w-4" />
              Expires: {new Date(metadata.expires_at).toLocaleString()}
            </p>
          )}
        </div>

        {metadata?.has_passphrase && (
          <div className="mt-4">
            <label htmlFor="passphrase" className="block text-sm font-medium mb-2">
              Enter Passphrase
            </label>
            <input
              id="passphrase"
              type="password"
              value={passphrase}
              onChange={(e) => setPassphrase(e.target.value)}
              placeholder="Enter the passphrase shared with you"
              className="w-full rounded-lg border border-border bg-background px-4 py-3 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary/50"
              onKeyDown={(e) => e.key === "Enter" && handleReveal()}
            />
          </div>
        )}

        <button
          onClick={handleReveal}
          disabled={revealing || (metadata?.has_passphrase && !passphrase)}
          className="mt-6 w-full rounded-lg bg-primary px-6 py-3 text-sm font-medium text-primary-foreground hover:bg-primary/90 transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
        >
          {revealing ? (
            <span className="animate-pulse">Decrypting...</span>
          ) : (
            <>
              <Eye className="h-4 w-4" />
              View Secret
            </>
          )}
        </button>
      </div>
    </div>
  );
}
