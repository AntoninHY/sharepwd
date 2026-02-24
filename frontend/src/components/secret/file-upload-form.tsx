"use client";

import { useState, useRef, useCallback } from "react";
import { toast } from "sonner";
import { Upload, Lock, Clock, Eye, Flame, Copy, X, File as FileIcon } from "lucide-react";
import { encryptText, encryptWithPassphrase, toBase64 } from "@/lib/crypto";
import { api, type CreateSecretResponse } from "@/lib/api";
import { EXPIRATION_OPTIONS, VIEW_OPTIONS } from "@/lib/types";

const APP_URL = process.env.NEXT_PUBLIC_APP_URL || "http://localhost:3000";
const MAX_FILE_SIZE = 5 * 1024 * 1024; // 5MB

export default function FileUploadForm() {
  const [file, setFile] = useState<File | null>(null);
  const [passphrase, setPassphrase] = useState("");
  const [expiresIn, setExpiresIn] = useState("24h");
  const [maxViews, setMaxViews] = useState<number | undefined>(undefined);
  const [burnAfterRead, setBurnAfterRead] = useState(false);
  const [loading, setLoading] = useState(false);
  const [progress, setProgress] = useState(0);
  const [shareUrl, setShareUrl] = useState<string | null>(null);
  const [isDragging, setIsDragging] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
    const dropped = e.dataTransfer.files[0];
    if (dropped) {
      if (dropped.size > MAX_FILE_SIZE) {
        toast.error("File too large (max 5MB)");
        return;
      }
      setFile(dropped);
    }
  }, []);

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const selected = e.target.files?.[0];
    if (selected) {
      if (selected.size > MAX_FILE_SIZE) {
        toast.error("File too large (max 5MB)");
        return;
      }
      setFile(selected);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!file) return;

    setLoading(true);
    setProgress(0);

    try {
      const arrayBuffer = await file.arrayBuffer();
      const fileBytes = new Uint8Array(arrayBuffer);

      // Pack filename + file content together before encryption
      const filePayload = JSON.stringify({
        name: file.name,
        data: toBase64(fileBytes),
      });

      let encryptedData: string;
      let iv: string;
      let salt: string | null = null;
      let keyFragment: string | null = null;

      setProgress(30);

      if (passphrase) {
        const result = await encryptWithPassphrase(filePayload, passphrase);
        encryptedData = result.encryptedData;
        iv = result.iv;
        salt = result.salt;
      } else {
        const result = await encryptText(filePayload);
        encryptedData = result.encryptedData;
        iv = result.iv;
        keyFragment = result.key;
      }

      setProgress(60);

      const response = await api.createSecret({
        encrypted_data: encryptedData,
        iv,
        salt,
        max_views: maxViews || null,
        expires_in: expiresIn || null,
        burn_after_read: burnAfterRead,
        content_type: "file",
      });

      setProgress(100);

      const url = keyFragment
        ? `${APP_URL}/f/${response.access_token}#${keyFragment}`
        : `${APP_URL}/f/${response.access_token}`;

      setShareUrl(url);
      toast.success("File encrypted and uploaded");
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to upload file");
    } finally {
      setLoading(false);
    }
  };

  const copyUrl = async () => {
    if (!shareUrl) return;
    await navigator.clipboard.writeText(shareUrl);
    toast.success("Link copied to clipboard");
  };

  if (shareUrl) {
    return (
      <div className="rounded-xl border border-border bg-card p-6">
        <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
          <Lock className="h-5 w-5 text-success" />
          File Encrypted & Shared
        </h2>
        <div className="flex items-center gap-2">
          <code className="flex-1 rounded-lg bg-muted px-4 py-3 text-sm break-all font-mono">
            {shareUrl}
          </code>
          <button onClick={copyUrl} className="shrink-0 rounded-lg bg-primary p-3 hover:bg-primary/90 transition-colors">
            <Copy className="h-4 w-4" />
          </button>
        </div>
        {passphrase && (
          <p className="mt-4 text-sm text-warning">
            Remember to share the passphrase via a separate channel.
          </p>
        )}
        <button
          onClick={() => { setShareUrl(null); setFile(null); setPassphrase(""); }}
          className="mt-4 rounded-lg border border-border px-4 py-2 text-sm hover:bg-muted transition-colors"
        >
          Upload Another
        </button>
      </div>
    );
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <div
        onDragOver={(e) => { e.preventDefault(); setIsDragging(true); }}
        onDragLeave={() => setIsDragging(false)}
        onDrop={handleDrop}
        onClick={() => inputRef.current?.click()}
        className={`relative rounded-xl border-2 border-dashed p-12 text-center cursor-pointer transition-colors ${
          isDragging ? "border-primary bg-primary/5" : "border-border hover:border-muted-foreground"
        }`}
      >
        <input ref={inputRef} type="file" className="hidden" onChange={handleFileSelect} />
        {file ? (
          <div className="flex items-center justify-center gap-3">
            <FileIcon className="h-8 w-8 text-primary" />
            <div className="text-left">
              <p className="font-medium">{file.name}</p>
              <p className="text-sm text-muted-foreground">{(file.size / 1024 / 1024).toFixed(2)} MB</p>
            </div>
            <button type="button" onClick={(e) => { e.stopPropagation(); setFile(null); }} className="ml-4 p-1 rounded hover:bg-muted">
              <X className="h-4 w-4" />
            </button>
          </div>
        ) : (
          <>
            <Upload className="h-10 w-10 text-muted-foreground mx-auto mb-3" />
            <p className="text-sm text-muted-foreground">
              Drop a file here or click to browse
            </p>
            <p className="text-xs text-muted-foreground mt-1">Max 5MB — need more? <a href="mailto:contact@jizo.ai" className="text-primary hover:underline">Contact us</a></p>
          </>
        )}
      </div>

      <div>
        <label htmlFor="file-passphrase" className="block text-sm font-medium mb-2 flex items-center gap-2">
          <Lock className="h-4 w-4" /> Passphrase (optional)
        </label>
        <input
          id="file-passphrase"
          type="password"
          value={passphrase}
          onChange={(e) => setPassphrase(e.target.value)}
          placeholder="Additional passphrase for extra security"
          className="w-full rounded-lg border border-border bg-card px-4 py-3 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary/50"
        />
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div>
          <label className="block text-sm font-medium mb-2 flex items-center gap-2">
            <Clock className="h-4 w-4" /> Expires After
          </label>
          <select value={expiresIn} onChange={(e) => setExpiresIn(e.target.value)} className="w-full rounded-lg border border-border bg-card px-4 py-3 text-sm focus:outline-none focus:ring-2 focus:ring-primary/50">
            {EXPIRATION_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>{opt.label}</option>
            ))}
          </select>
        </div>
        <div>
          <label className="block text-sm font-medium mb-2 flex items-center gap-2">
            <Eye className="h-4 w-4" /> Max Views
          </label>
          <select value={maxViews ?? ""} onChange={(e) => setMaxViews(e.target.value ? Number(e.target.value) : undefined)} className="w-full rounded-lg border border-border bg-card px-4 py-3 text-sm focus:outline-none focus:ring-2 focus:ring-primary/50">
            <option value="">Unlimited</option>
            {VIEW_OPTIONS.map((n) => (
              <option key={n} value={n}>{n} view{n > 1 ? "s" : ""}</option>
            ))}
          </select>
        </div>
      </div>

      <label className="flex items-center gap-3 cursor-pointer">
        <input type="checkbox" checked={burnAfterRead} onChange={(e) => setBurnAfterRead(e.target.checked)} className="rounded border-border bg-card h-4 w-4 accent-primary" />
        <span className="text-sm flex items-center gap-2">
          <Flame className="h-4 w-4 text-destructive" /> Burn after reading
        </span>
      </label>

      {loading && (
        <div className="w-full rounded-full bg-muted h-2">
          <div className="bg-primary h-2 rounded-full transition-all duration-300" style={{ width: `${progress}%` }} />
        </div>
      )}

      <button
        type="submit"
        disabled={loading || !file}
        className="w-full rounded-lg bg-primary px-6 py-3 text-sm font-medium text-primary-foreground hover:bg-primary/90 transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
      >
        {loading ? <span className="animate-pulse">Encrypting & Uploading...</span> : <><Lock className="h-4 w-4" /> Encrypt & Share File</>}
      </button>
    </form>
  );
}
