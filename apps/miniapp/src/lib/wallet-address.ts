/** Из CAIP-адреса Reown / строки WC достаём base58 Solana pubkey. */
export function parseSolanaAddress(input: unknown): string | null {
  if (!input) return null;
  if (typeof input === 'string') {
    const trimmed = input.trim();
    if (!trimmed) return null;
    if (trimmed.includes(':')) {
      const part = trimmed.split(':').pop();
      return part && part.length >= 32 ? part : null;
    }
    return trimmed.length >= 32 ? trimmed : null;
  }
  if (typeof input === 'object' && input !== null && 'address' in input) {
    const addr = String((input as { address: string }).address);
    return addr.length >= 32 ? addr : null;
  }
  return null;
}

export async function waitFor<T>(
  getter: () => T | null | undefined,
  timeoutMs = 20_000,
  intervalMs = 250,
): Promise<T> {
  const start = Date.now();
  while (Date.now() - start < timeoutMs) {
    const value = getter();
    if (value) return value;
    await new Promise((r) => window.setTimeout(r, intervalMs));
  }
  throw new Error('Таймаут ожидания кошелька');
}
