export interface PowResult {
  counter: number;
  iterations: number;
}

export function solvePoW(
  prefix: string,
  difficulty: number
): { promise: Promise<PowResult>; abort: () => void } {
  const worker = new Worker("/pow-worker.js");

  const promise = new Promise<PowResult>((resolve, reject) => {
    worker.onmessage = (e: MessageEvent) => {
      if (e.data.counter !== undefined) {
        worker.terminate();
        resolve({ counter: e.data.counter, iterations: e.data.iterations });
      }
    };

    worker.onerror = (err) => {
      worker.terminate();
      reject(new Error(`PoW worker error: ${err.message}`));
    };
  });

  worker.postMessage({ prefix, difficulty });

  return {
    promise,
    abort: () => worker.terminate(),
  };
}
