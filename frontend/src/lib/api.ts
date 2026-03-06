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
};

export { APIError };
