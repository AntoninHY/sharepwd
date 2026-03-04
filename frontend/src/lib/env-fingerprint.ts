// SharePwd.io — Burn After Reading
// Zero-knowledge secret sharing. Client-side AES-256-GCM encryption, secrets self-destruct after viewing.
// Copyright (c) 2025-2026 Antonin HILY — CTO, Jizo AI

interface EnvFingerprintData {
  wd: boolean; // navigator.webdriver
  pc: number; // navigator.plugins.length
  lc: number; // navigator.languages.length
  sw: number; // screen.width
  sh: number; // screen.height
  cd: number; // screen.colorDepth
  hc: number; // navigator.hardwareConcurrency
  hn: boolean; // has Notification API
  dm: number; // navigator.deviceMemory
  pt: boolean; // has real performance navigation timing
}

export function collectEnvFingerprint(): string {
  const nav = navigator as Navigator & { deviceMemory?: number };

  const data: EnvFingerprintData = {
    wd: !!(nav as Navigator & { webdriver?: boolean }).webdriver,
    pc: nav.plugins?.length ?? 0,
    lc: nav.languages?.length ?? 0,
    sw: screen.width ?? 0,
    sh: screen.height ?? 0,
    cd: screen.colorDepth ?? 0,
    hc: nav.hardwareConcurrency ?? 0,
    hn: typeof Notification !== "undefined",
    dm: nav.deviceMemory ?? 0,
    pt: hasRealPerfTiming(),
  };

  return btoa(JSON.stringify(data));
}

function hasRealPerfTiming(): boolean {
  try {
    const entries = performance.getEntriesByType("navigation");
    if (entries.length === 0) return false;
    const nav = entries[0] as PerformanceNavigationTiming;
    return nav.domContentLoadedEventStart > 0;
  } catch {
    return false;
  }
}
