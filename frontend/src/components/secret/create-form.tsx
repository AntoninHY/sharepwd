"use client";

import { useState } from "react";
import { toast } from "sonner";
import { useTranslations } from "next-intl";
import { Lock, Eye, EyeOff, Clock, Flame, Copy, ExternalLink } from "lucide-react";
import PassphraseGenerator from "@/components/ui/passphrase-generator";
import { encryptText, encryptWithPassphrase } from "@/lib/crypto";
import { api } from "@/lib/api";
import { EXPIRATION_OPTIONS, VIEW_OPTIONS } from "@/lib/types";

const APP_URL = process.env.NEXT_PUBLIC_APP_URL || "http://localhost:3000";

export default function CreateForm() {
  const t = useTranslations("createForm");
  const tExp = useTranslations("expiration");
  const tViews = useTranslations("views");

  const [content, setContent] = useState("");
  const [passphrase, setPassphrase] = useState("");
  const [expiresIn, setExpiresIn] = useState("24h");
  const [maxViews, setMaxViews] = useState<number | undefined>(undefined);
  const [burnAfterRead, setBurnAfterRead] = useState(false);
  const [showPassphrase, setShowPassphrase] = useState(false);
  const [loading, setLoading] = useState(false);
  const [shareUrl, setShareUrl] = useState<string | null>(null);
  const [creatorToken, setCreatorToken] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!content.trim()) {
      toast.error(t("emptyError"));
      return;
    }

    setLoading(true);
    try {
      let encryptedData: string;
      let iv: string;
      let salt: string | null = null;
      let keyFragment: string | null = null;

      if (passphrase) {
        const result = await encryptWithPassphrase(content, passphrase);
        encryptedData = result.encryptedData;
        iv = result.iv;
        salt = result.salt;
      } else {
        const result = await encryptText(content);
        encryptedData = result.encryptedData;
        iv = result.iv;
        keyFragment = result.key;
      }

      const response = await api.createSecret({
        encrypted_data: encryptedData,
        iv,
        salt,
        max_views: maxViews || null,
        expires_in: expiresIn || null,
        burn_after_read: burnAfterRead,
        content_type: "text",
      });

      const url = keyFragment
        ? `${APP_URL}/s/${response.access_token}#${keyFragment}`
        : `${APP_URL}/s/${response.access_token}`;

      setShareUrl(url);
      setCreatorToken(response.creator_token);
      toast.success(t("toastSuccess"));
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t("toastError"));
    } finally {
      setLoading(false);
    }
  };

  const copyUrl = async () => {
    if (!shareUrl) return;
    await navigator.clipboard.writeText(shareUrl);
    toast.success(t("toastCopied"));
  };

  if (shareUrl) {
    return (
      <div className="space-y-6">
        <div className="rounded-xl border border-border bg-card p-6">
          <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
            <Lock className="h-5 w-5 text-success" />
            {t("successTitle")}
          </h2>
          <p className="text-sm text-muted-foreground mb-4">
            {passphrase ? t("successDescPassphrase") : t("successDescLink")}
          </p>
          <div className="flex items-center gap-2">
            <code className="flex-1 rounded-lg bg-muted px-4 py-3 text-sm break-all font-mono">
              {shareUrl}
            </code>
            <button
              onClick={copyUrl}
              className="shrink-0 rounded-lg bg-primary p-3 hover:bg-primary/90 transition-colors"
              title="Copy link"
            >
              <Copy className="h-4 w-4" />
            </button>
          </div>
          {passphrase && (
            <p className="mt-4 text-sm text-warning">{t("passphraseWarning")}</p>
          )}
          <div className="mt-6 flex gap-3">
            <button
              onClick={() => {
                setShareUrl(null);
                setCreatorToken(null);
                setContent("");
                setPassphrase("");
              }}
              className="rounded-lg border border-border px-4 py-2 text-sm hover:bg-muted transition-colors"
            >
              {t("createAnother")}
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <div>
        <label htmlFor="content" className="block text-sm font-medium mb-2">
          {t("contentLabel")}
        </label>
        <textarea
          id="content"
          value={content}
          onChange={(e) => setContent(e.target.value)}
          placeholder={t("contentPlaceholder")}
          className="w-full rounded-lg border border-border bg-card px-4 py-3 text-sm font-mono min-h-[160px] placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary/50 resize-y"
          maxLength={100_000}
        />
        <p className="mt-1 text-xs text-muted-foreground">
          {t("charCount", { count: content.length, max: 100000 })}
        </p>
      </div>

      <div>
        <label htmlFor="passphrase" className="block text-sm font-medium mb-2 flex items-center gap-2">
          <Lock className="h-4 w-4" />
          {t("passphraseLabel")}
        </label>
        <div className="relative">
          <input
            id="passphrase"
            type={showPassphrase ? "text" : "password"}
            value={passphrase}
            onChange={(e) => setPassphrase(e.target.value)}
            placeholder={t("passphrasePlaceholder")}
            className="w-full rounded-lg border border-border bg-card px-4 py-3 pr-10 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary/50"
          />
          <button
            type="button"
            onClick={() => setShowPassphrase(!showPassphrase)}
            className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors"
            title={t(showPassphrase ? "hidePassphrase" : "showPassphrase")}
          >
            {showPassphrase ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
          </button>
        </div>
        <PassphraseGenerator
          translationNamespace="createForm"
          onGenerate={(value) => {
            setPassphrase(value);
            setShowPassphrase(true);
          }}
        />
        <p className="mt-1 text-xs text-muted-foreground">
          {passphrase ? t("passphraseHintSet") : t("passphraseHintNone")}
        </p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div>
          <label htmlFor="expires" className="block text-sm font-medium mb-2 flex items-center gap-2">
            <Clock className="h-4 w-4" />
            {t("expiresLabel")}
          </label>
          <select
            id="expires"
            value={expiresIn}
            onChange={(e) => setExpiresIn(e.target.value)}
            className="w-full rounded-lg border border-border bg-card px-4 py-3 text-sm focus:outline-none focus:ring-2 focus:ring-primary/50"
          >
            {EXPIRATION_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {tExp(opt.value)}
              </option>
            ))}
          </select>
        </div>

        <div>
          <label htmlFor="maxViews" className="block text-sm font-medium mb-2 flex items-center gap-2">
            <Eye className="h-4 w-4" />
            {t("maxViewsLabel")}
          </label>
          <select
            id="maxViews"
            value={burnAfterRead ? "" : (maxViews ?? "")}
            onChange={(e) => setMaxViews(e.target.value ? Number(e.target.value) : undefined)}
            disabled={burnAfterRead}
            className="w-full rounded-lg border border-border bg-card px-4 py-3 text-sm focus:outline-none focus:ring-2 focus:ring-primary/50 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <option value="">{burnAfterRead ? t("burnSingleView") : t("unlimited")}</option>
            {VIEW_OPTIONS.map((n) => (
              <option key={n} value={n}>
                {tViews("count", { count: n })}
              </option>
            ))}
          </select>
        </div>
      </div>

      <label className="flex items-center gap-3 cursor-pointer">
        <input
          type="checkbox"
          checked={burnAfterRead}
          onChange={(e) => {
            setBurnAfterRead(e.target.checked);
            if (e.target.checked) setMaxViews(undefined);
          }}
          className="rounded border-border bg-card h-4 w-4 accent-primary"
        />
        <span className="text-sm flex items-center gap-2">
          <Flame className="h-4 w-4 text-destructive" />
          {t("burnLabel")}
        </span>
      </label>

      <button
        type="submit"
        disabled={loading || !content.trim()}
        className="w-full rounded-lg bg-primary px-6 py-3 text-sm font-medium text-primary-foreground hover:bg-primary/90 transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
      >
        {loading ? (
          <span className="animate-pulse">{t("submitting")}</span>
        ) : (
          <>
            <Lock className="h-4 w-4" />
            {t("submitButton")}
          </>
        )}
      </button>
    </form>
  );
}
