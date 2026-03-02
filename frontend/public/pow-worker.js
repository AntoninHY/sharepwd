// Proof-of-Work Web Worker — SHA-256 Hashcash
// Finds counter such that SHA-256(prefix + ":" + counter) has N leading zero bits.

self.onmessage = async function (e) {
  const { prefix, difficulty } = e.data;
  const target = difficulty;
  let counter = 1;

  while (true) {
    const input = prefix + ":" + counter;
    const buffer = new TextEncoder().encode(input);
    const hashBuffer = await crypto.subtle.digest("SHA-256", buffer);
    const hash = new Uint8Array(hashBuffer);

    if (hasLeadingZeroBits(hash, target)) {
      self.postMessage({ counter, iterations: counter });
      return;
    }

    counter++;

    if (counter % 10000 === 0) {
      self.postMessage({ progress: counter });
    }
  }
};

function hasLeadingZeroBits(hash, bits) {
  const fullBytes = Math.floor(bits / 8);
  const remaining = bits % 8;

  for (let i = 0; i < fullBytes; i++) {
    if (hash[i] !== 0) return false;
  }

  if (remaining > 0) {
    const mask = 0xff << (8 - remaining);
    if ((hash[fullBytes] & mask) !== 0) return false;
  }

  return true;
}
