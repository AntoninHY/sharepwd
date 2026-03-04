// SharePwd.io — Burn After Reading
// Zero-knowledge secret sharing. Client-side AES-256-GCM encryption, secrets self-destruct after viewing.
// Copyright (c) 2025-2026 Antonin HILY — CTO, Jizo AI

/**
 * Zero-knowledge encryption module using Web Crypto API.
 * All encryption/decryption happens client-side.
 * The server NEVER sees plaintext data.
 */

const PBKDF2_ITERATIONS = 600_000;
const AES_KEY_LENGTH = 256;
const IV_LENGTH = 12;
const SALT_LENGTH = 16;

export interface EncryptResult {
  encryptedData: string;
  iv: string;
  key: string;
}

export interface EncryptWithPassphraseResult {
  encryptedData: string;
  iv: string;
  salt: string;
}

function copyToArrayBuffer(src: Uint8Array): ArrayBuffer {
  const dst = new ArrayBuffer(src.byteLength);
  new Uint8Array(dst).set(src);
  return dst;
}

function toBase64(data: Uint8Array | ArrayBuffer): string {
  const bytes = data instanceof Uint8Array ? data : new Uint8Array(data);
  let binary = "";
  for (let i = 0; i < bytes.byteLength; i++) {
    binary += String.fromCharCode(bytes[i]);
  }
  return btoa(binary);
}

function fromBase64(base64: string): Uint8Array {
  const binary = atob(base64);
  const bytes = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i++) {
    bytes[i] = binary.charCodeAt(i);
  }
  return bytes;
}

function toBase64Url(data: Uint8Array | ArrayBuffer): string {
  return toBase64(data)
    .replace(/\+/g, "-")
    .replace(/\//g, "_")
    .replace(/=+$/, "");
}

function fromBase64Url(base64url: string): Uint8Array {
  let base64 = base64url.replace(/-/g, "+").replace(/_/g, "/");
  while (base64.length % 4) {
    base64 += "=";
  }
  return fromBase64(base64);
}

async function generateKey(): Promise<CryptoKey> {
  return crypto.subtle.generateKey(
    { name: "AES-GCM", length: AES_KEY_LENGTH },
    true,
    ["encrypt", "decrypt"]
  );
}

async function deriveKeyFromPassphrase(
  passphrase: string,
  salt: Uint8Array
): Promise<CryptoKey> {
  const encoder = new TextEncoder();
  const keyMaterial = await crypto.subtle.importKey(
    "raw",
    encoder.encode(passphrase),
    "PBKDF2",
    false,
    ["deriveKey"]
  );

  return crypto.subtle.deriveKey(
    {
      name: "PBKDF2",
      salt: copyToArrayBuffer(salt) as BufferSource,
      iterations: PBKDF2_ITERATIONS,
      hash: "SHA-256",
    },
    keyMaterial,
    { name: "AES-GCM", length: AES_KEY_LENGTH },
    false,
    ["encrypt", "decrypt"]
  );
}

export async function encryptText(plaintext: string): Promise<EncryptResult> {
  const key = await generateKey();
  const iv = crypto.getRandomValues(new Uint8Array(IV_LENGTH));
  const encoder = new TextEncoder();

  const ciphertext = await crypto.subtle.encrypt(
    { name: "AES-GCM", iv: copyToArrayBuffer(iv) as BufferSource },
    key,
    encoder.encode(plaintext)
  );

  const rawKey = await crypto.subtle.exportKey("raw", key);

  return {
    encryptedData: toBase64(new Uint8Array(ciphertext)),
    iv: toBase64(iv),
    key: toBase64Url(new Uint8Array(rawKey)),
  };
}

export async function decryptText(
  encryptedData: string,
  iv: string,
  keyBase64Url: string
): Promise<string> {
  const keyBytes = fromBase64Url(keyBase64Url);
  const key = await crypto.subtle.importKey(
    "raw",
    copyToArrayBuffer(keyBytes),
    { name: "AES-GCM", length: AES_KEY_LENGTH },
    false,
    ["decrypt"]
  );

  const ivBytes = fromBase64(iv);
  const cipherBytes = fromBase64(encryptedData);

  const decrypted = await crypto.subtle.decrypt(
    { name: "AES-GCM", iv: copyToArrayBuffer(ivBytes) as BufferSource },
    key,
    copyToArrayBuffer(cipherBytes)
  );

  return new TextDecoder().decode(decrypted);
}

export async function encryptWithPassphrase(
  plaintext: string,
  passphrase: string
): Promise<EncryptWithPassphraseResult> {
  const salt = crypto.getRandomValues(new Uint8Array(SALT_LENGTH));
  const key = await deriveKeyFromPassphrase(passphrase, salt);
  const iv = crypto.getRandomValues(new Uint8Array(IV_LENGTH));
  const encoder = new TextEncoder();

  const ciphertext = await crypto.subtle.encrypt(
    { name: "AES-GCM", iv: copyToArrayBuffer(iv) as BufferSource },
    key,
    encoder.encode(plaintext)
  );

  return {
    encryptedData: toBase64(new Uint8Array(ciphertext)),
    iv: toBase64(iv),
    salt: toBase64(salt),
  };
}

export async function decryptWithPassphrase(
  encryptedData: string,
  iv: string,
  salt: string,
  passphrase: string
): Promise<string> {
  const saltBytes = fromBase64(salt);
  const key = await deriveKeyFromPassphrase(passphrase, saltBytes);

  const ivBytes = fromBase64(iv);
  const cipherBytes = fromBase64(encryptedData);

  const decrypted = await crypto.subtle.decrypt(
    { name: "AES-GCM", iv: copyToArrayBuffer(ivBytes) as BufferSource },
    key,
    copyToArrayBuffer(cipherBytes)
  );

  return new TextDecoder().decode(decrypted);
}

export { toBase64, fromBase64, toBase64Url, fromBase64Url };
