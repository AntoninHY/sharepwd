// SharePwd.io — Burn After Reading
// Zero-knowledge secret sharing. Client-side AES-256-GCM encryption, secrets self-destruct after viewing.
// Copyright (c) 2025-2026 Antonin HILY — CTO, Jizo AI

interface MovementSample {
  x: number;
  y: number;
  t: number;
}

export class BehavioralCollector {
  private movements: MovementSample[] = [];
  private eventCount = 0;
  private startTime = 0;
  private hasTouch = false;
  private hasMouse = false;
  private active = false;

  private onPointerMove = (e: PointerEvent) => {
    this.eventCount++;
    if (e.pointerType === "touch") this.hasTouch = true;
    if (e.pointerType === "mouse") this.hasMouse = true;
    this.movements.push({ x: e.clientX, y: e.clientY, t: performance.now() });
  };

  private onClick = () => {
    this.eventCount++;
  };

  private onScroll = () => {
    this.eventCount++;
  };

  start(): void {
    if (this.active) return;
    this.active = true;
    this.startTime = performance.now();
    window.addEventListener("pointermove", this.onPointerMove, { passive: true });
    window.addEventListener("click", this.onClick, { passive: true });
    window.addEventListener("scroll", this.onScroll, { passive: true });
  }

  stop(): void {
    if (!this.active) return;
    this.active = false;
    window.removeEventListener("pointermove", this.onPointerMove);
    window.removeEventListener("click", this.onClick);
    window.removeEventListener("scroll", this.onScroll);
  }

  generateProof(): string {
    const mc = this.movements.length;
    const me = Math.round(this.computeEntropy() * 1000);
    const vv = Math.round(this.computeVelocityVariance());
    const sr = Math.round(this.computeStraightLineRatio());
    const ec = this.eventCount;
    const ts = Math.round(performance.now() - this.startTime);
    const ht = this.hasTouch ? 1 : 0;
    const hm = this.hasMouse ? 1 : 0;

    const proof = JSON.stringify({ mc, me, vv, sr, ec, ts, ht, hm });
    return btoa(proof);
  }

  private computeEntropy(): number {
    if (this.movements.length < 3) return 0;

    const bins = new Array(8).fill(0);
    let total = 0;

    for (let i = 1; i < this.movements.length; i++) {
      const dx = this.movements[i].x - this.movements[i - 1].x;
      const dy = this.movements[i].y - this.movements[i - 1].y;
      if (dx === 0 && dy === 0) continue;

      let angle = Math.atan2(dy, dx);
      if (angle < 0) angle += 2 * Math.PI;
      const bin = Math.floor((angle / (2 * Math.PI)) * 8) % 8;
      bins[bin]++;
      total++;
    }

    if (total === 0) return 0;

    let entropy = 0;
    for (const count of bins) {
      if (count === 0) continue;
      const p = count / total;
      entropy -= p * Math.log2(p);
    }

    return entropy;
  }

  private computeVelocityVariance(): number {
    if (this.movements.length < 3) return 0;

    const velocities: number[] = [];
    for (let i = 1; i < this.movements.length; i++) {
      const dx = this.movements[i].x - this.movements[i - 1].x;
      const dy = this.movements[i].y - this.movements[i - 1].y;
      const dt = this.movements[i].t - this.movements[i - 1].t;
      if (dt === 0) continue;
      velocities.push(Math.sqrt(dx * dx + dy * dy) / dt);
    }

    if (velocities.length < 2) return 0;

    const mean = velocities.reduce((a, b) => a + b, 0) / velocities.length;
    return velocities.reduce((sum, v) => sum + (v - mean) * (v - mean), 0) / velocities.length;
  }

  private computeStraightLineRatio(): number {
    if (this.movements.length < 3) return 0;

    let straightSegments = 0;
    let totalSegments = 0;
    const ANGLE_THRESHOLD = 0.05; // ~3 degrees

    for (let i = 2; i < this.movements.length; i++) {
      const dx1 = this.movements[i - 1].x - this.movements[i - 2].x;
      const dy1 = this.movements[i - 1].y - this.movements[i - 2].y;
      const dx2 = this.movements[i].x - this.movements[i - 1].x;
      const dy2 = this.movements[i].y - this.movements[i - 1].y;

      const len1 = Math.sqrt(dx1 * dx1 + dy1 * dy1);
      const len2 = Math.sqrt(dx2 * dx2 + dy2 * dy2);
      if (len1 === 0 || len2 === 0) continue;

      totalSegments++;
      const angle1 = Math.atan2(dy1, dx1);
      const angle2 = Math.atan2(dy2, dx2);
      let diff = Math.abs(angle2 - angle1);
      if (diff > Math.PI) diff = 2 * Math.PI - diff;

      if (diff < ANGLE_THRESHOLD) {
        straightSegments++;
      }
    }

    if (totalSegments === 0) return 0;
    return (straightSegments / totalSegments) * 100;
  }
}
