// SharePwd.io — Burn After Reading
// Zero-knowledge secret sharing. Client-side AES-256-GCM encryption, secrets self-destruct after viewing.
// Copyright (c) 2025-2026 Antonin HILY — CTO, Jizo AI

"use client";

import { useState, useEffect, useRef } from "react";
import { toast } from "sonner";
import { useTranslations } from "next-intl";
import { Eye, Lock, AlertTriangle, Copy, Clock, Flame } from "lucide-react";
import { decryptText, decryptWithPassphrase } from "@/lib/crypto";
import { api, type SecretMetadata, type RevealSecretResponse } from "@/lib/api";
import { solvePoW, type PowResult } from "@/lib/pow";
import { BehavioralCollector } from "@/lib/behavioral";
import { collectEnvFingerprint } from "@/lib/env-fingerprint";
import { signProof } from "@/lib/hmac";

// Minimum delay (ms) between nonce issuance and reveal request.
// Must exceed the server-side ChallengeMinSolveTime (1500ms) plus network latency margin.
const MIN_CHALLENGE_DELAY = 2000;

interface RevealGateProps {
  token: string;
}

export default function RevealGate({ token }: RevealGateProps) {
  const t = useTranslations("reveal");

  const [metadata, setMetadata] = useState<SecretMetadata | null>(null);
  const [decryptedContent, setDecryptedContent] = useState<string | null>(null);
  const [passphrase, setPassphrase] = useState("");
  const [loading, setLoading] = useState(true);
  const [revealing, setRevealing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [passphraseError, setPassphraseError] = useState<string | null>(null);
  const [revealedData, setRevealedData] = useState<RevealSecretResponse | null>(null);

  const powPromiseRef = useRef<Promise<PowResult> | null>(null);
  const powAbortRef = useRef<(() => void) | null>(null);
  const behavioralRef = useRef<BehavioralCollector | null>(null);
  const challengeReceivedAt = useRef<number>(0);

  // Start behavioral collector on mount
  useEffect(() => {
    const collector = new BehavioralCollector();
    collector.start();
    behavioralRef.current = collector;
    return () => collector.stop();
  }, []);

  // Fetch metadata and start PoW solver
  useEffect(() => {
    async function fetchMetadata() {
      try {
        const meta = await api.getSecretMetadata(token);
        challengeReceivedAt.current = Date.now();
        if (meta.is_expired) {
          setError(t("errorExpired"));
          return;
        }
        setMetadata(meta);

        // Start PoW in background Web Worker
        if (meta.pow_challenge && meta.pow_difficulty) {
          const { promise, abort } = solvePoW(meta.pow_challenge, meta.pow_difficulty);
          powAbortRef.current = abort;
          powPromiseRef.current = promise;
        }
      } catch {
        setError(t("errorNotFound"));
      } finally {
        setLoading(false);
      }
    }
    fetchMetadata();

    return () => {
      powAbortRef.current?.();
    };
  }, [token, t]);

  const isDecryptionError = (err: unknown): boolean => {
    if (!(err instanceof Error)) return false;
    const msg = err.message.toLowerCase();
    return msg.includes("decrypt") || msg.includes("operation") || msg.includes("tag") || msg.includes("gcm");
  };

  const handleReveal = async () => {
    if (!metadata) return;

    if (metadata.has_passphrase && !passphrase) {
      toast.error(t("toastPassphraseRequired"));
      return;
    }

    setRevealing(true);
    setPassphraseError(null);
    try {
      // Wait for PoW to complete before sending reveal
      const powResult = powPromiseRef.current ? await powPromiseRef.current : null;

      // Ensure minimum delay since challenge was issued (server-side grace period)
      const elapsed = Date.now() - challengeReceivedAt.current;
      if (elapsed < MIN_CHALLENGE_DELAY) {
        await new Promise((r) => setTimeout(r, MIN_CHALLENGE_DELAY - elapsed));
      }

      // Collect defense proofs and sign them with HMAC
      const behavioralProof = behavioralRef.current?.generateProof();
      const envFp = collectEnvFingerprint();

      let behavioralSig: string | undefined;
      let envSig: string | undefined;

      if (metadata.hmac_key) {
        if (behavioralProof) {
          behavioralSig = await signProof(metadata.hmac_key, metadata.challenge_nonce, behavioralProof);
        }
        if (envFp) {
          envSig = await signProof(metadata.hmac_key, metadata.challenge_nonce, envFp);
        }
      }

      const revealed = revealedData || await api.revealSecret(
        token,
        metadata.challenge_nonce,
        powResult?.counter,
        behavioralProof,
        behavioralSig,
        envFp,
        envSig,
      );
      if (!revealedData) setRevealedData(revealed);

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
          setError(msg || t("errorDecrypt"));
        }
      }
    } finally {
      setRevealing(false);
    }
  };

  const copyContent = async () => {
    if (!decryptedContent) return;
    await navigator.clipboard.writeText(decryptedContent);
    toast.success(t("toastCopied"));
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center py-16">
        <div className="animate-pulse text-muted-foreground">{t("loading")}</div>
      </div>
    );
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

  if (decryptedContent !== null) {
    return (
      <div className="max-w-2xl mx-auto space-y-4">
        <div className="rounded-xl border border-success/50 bg-card p-6">
          <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
            <Eye className="h-5 w-5 text-success" />
            {t("revealedTitle")}
          </h2>
          {metadata?.burn_after_read && (
            <p className="text-sm text-warning mb-4 flex items-center gap-2">
              <Flame className="h-4 w-4" />
              {t("burnWarning")}
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
          {t("sharedTitle")}
        </h2>

        <div className="mt-4 space-y-3 text-sm text-muted-foreground">
          {metadata?.burn_after_read && (
            <p className="flex items-center gap-2 text-warning">
              <Flame className="h-4 w-4" />
              {t("burnNotice")}
            </p>
          )}
          {metadata?.max_views && (
            <p className="flex items-center gap-2">
              <Eye className="h-4 w-4" />
              {t("viewsUsed", { current: metadata.current_views, max: metadata.max_views })}
            </p>
          )}
          {metadata?.expires_at && (
            <p className="flex items-center gap-2">
              <Clock className="h-4 w-4" />
              {t("expires", { date: new Date(metadata.expires_at).toLocaleString() })}
            </p>
          )}
        </div>

        {metadata?.has_passphrase && (
          <div className="mt-4">
            <label htmlFor="passphrase" className="block text-sm font-medium mb-2">
              {t("enterPassphrase")}
            </label>
            <input
              id="passphrase"
              type="password"
              value={passphrase}
              onChange={(e) => { setPassphrase(e.target.value); setPassphraseError(null); }}
              placeholder={t("passphrasePlaceholder")}
              className={`w-full rounded-lg border px-4 py-3 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary/50 ${
                passphraseError ? "border-destructive bg-destructive/5" : "border-border bg-background"
              }`}
              onKeyDown={(e) => e.key === "Enter" && handleReveal()}
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
          onClick={handleReveal}
          disabled={revealing || (metadata?.has_passphrase && !passphrase)}
          className="mt-6 w-full rounded-lg bg-primary px-6 py-3 text-sm font-medium text-primary-foreground hover:bg-primary/90 transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
        >
          {revealing ? (
            <span className="animate-pulse">{t("decrypting")}</span>
          ) : (
            <>
              <Eye className="h-4 w-4" />
              {t("viewButton")}
            </>
          )}
        </button>
      </div>
    </div>
  );
}
