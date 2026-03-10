// SharePwd.io — Burn After Reading
// Zero-knowledge secret sharing. Client-side AES-256-GCM encryption, secrets self-destruct after viewing.
// Copyright (c) 2025-2026 Antonin HILY — CTO, Jizo AI

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export interface SecretMetadata {
  access_token: string;
  has_passphrase: boolean;
  content_type: "text" | "file";
  max_views: number | null;
  current_views: number;
  expires_at: string | null;
  burn_after_read: boolean;
  is_expired: boolean;
  created_at: string;
  challenge_nonce: string;
  pow_challenge: string;
  pow_difficulty: number;
  hmac_key: string;
}

export interface CreateSecretPayload {
  encrypted_data: string;
  iv: string;
  salt?: string | null;
  max_views?: number | null;
  expires_in?: string | null;
  burn_after_read: boolean;
  content_type: string;
}

export interface CreateSecretResponse {
  access_token: string;
  creator_token: string;
  expires_at: string | null;
}

export interface RevealSecretResponse {
  encrypted_data: string;
  iv: string;
  salt?: string | null;
}


export interface InitFileUploadPayload {
  encrypted_name: string;
  original_size: number;
  chunk_count: number;
  max_views?: number | null;
  expires_in?: string | null;
  burn_after_read: boolean;
  iv: string;
  salt?: string | null;
}

export interface InitFileUploadResponse {
  access_token: string;
  file_id: string;
  creator_token: string;
}

export interface FileInfo {
  file_id: string;
  encrypted_name: string;
  original_size: number;
  chunk_count: number;
}

class APIError extends Error {
  constructor(public status: number, message: string) {
    super(message);
    this.name = "APIError";
  }
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${API_URL}${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...options?.headers,
    },
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: "Unknown error" }));
    throw new APIError(res.status, body.error || `HTTP ${res.status}`);
  }

  if (res.status === 204) return undefined as T;
  return res.json();
}

export const api = {
  createSecret(payload: CreateSecretPayload): Promise<CreateSecretResponse> {
    return request("/v1/secrets", {
      method: "POST",
      body: JSON.stringify(payload),
    });
  },

  getSecretMetadata(token: string): Promise<SecretMetadata> {
    return request(`/v1/secrets/${token}`);
  },

  revealSecret(
    token: string,
    challengeNonce: string,
    powSolution?: number,
    behavioralProof?: string,
    behavioralSig?: string,
    envFingerprint?: string,
    envSig?: string,
  ): Promise<RevealSecretResponse> {
    return request(`/v1/secrets/${token}/reveal`, {
      method: "POST",
      body: JSON.stringify({
        challenge_nonce: challengeNonce,
        ...(powSolution && { pow_solution: powSolution }),
        ...(behavioralProof && { behavioral_proof: behavioralProof }),
        ...(behavioralSig && { behavioral_sig: behavioralSig }),
        ...(envFingerprint && { env_fingerprint: envFingerprint }),
        ...(envSig && { env_sig: envSig }),
      }),
    });
  },

  deleteSecret(token: string, creatorToken: string): Promise<void> {
    return request(`/v1/secrets/${token}`, {
      method: "DELETE",
      body: JSON.stringify({ creator_token: creatorToken }),
    });
  },

  initFileUpload(payload: InitFileUploadPayload): Promise<InitFileUploadResponse> {
    return request("/v1/secrets/file", {
      method: "POST",
      body: JSON.stringify(payload),
    });
  },

  async uploadChunk(fileId: string, chunkN: number, data: Uint8Array): Promise<void> {
    const res = await fetch(`${API_URL}/v1/secrets/file/${fileId}/chunk/${chunkN}`, {
      method: "PUT",
      headers: { "Content-Type": "application/octet-stream" },
      body: data.buffer as ArrayBuffer,
    });
    if (!res.ok) {
      const body = await res.json().catch(() => ({ error: "Upload failed" }));
      throw new APIError(res.status, body.error || `HTTP ${res.status}`);
    }
  },

  completeUpload(fileId: string): Promise<void> {
    return request(`/v1/secrets/file/${fileId}/complete`, {
      method: "POST",
    });
  },

  getFileInfo(token: string): Promise<FileInfo> {
    return request(`/v1/secrets/${token}/file`);
  },

  async downloadChunk(fileId: string, chunkN: number, token: string): Promise<Uint8Array> {
    const res = await fetch(
      `${API_URL}/v1/secrets/file/${fileId}/chunk/${chunkN}?token=${token}`,
    );
    if (!res.ok) {
      throw new APIError(res.status, "Failed to download chunk");
    }
    return new Uint8Array(await res.arrayBuffer());
  },
};

export { APIError };
