import { Transaction } from '@solana/web3.js';

type PhantomProvider = {
  isPhantom?: boolean;
  publicKey: { toString(): string } | null;
  connect(): Promise<{ publicKey: { toString(): string } }>;
  signMessage(
    message: Uint8Array,
    display?: string,
  ): Promise<{ signature: Uint8Array }>;
  signAndSendTransaction(
    transaction: Transaction,
  ): Promise<{ signature: string }>;
};

declare global {
  interface Window {
    phantom?: { solana?: PhantomProvider };
    solana?: PhantomProvider;
  }
}

export function getPhantom(): PhantomProvider | null {
  return window.phantom?.solana ?? window.solana ?? null;
}

export async function connectPhantom(): Promise<string> {
  const provider = getPhantom();
  if (!provider) {
    throw new Error(
      'Phantom не найден. Открой Mini App через Phantom Browser или установи кошелёк.',
    );
  }
  const res = await provider.connect();
  return res.publicKey.toString();
}

export async function signMessagePhantom(message: string): Promise<string> {
  const provider = getPhantom();
  if (!provider?.publicKey) {
    throw new Error('Сначала подключи Phantom');
  }
  const encoded = new TextEncoder().encode(message);
  const { signature } = await provider.signMessage(encoded, 'utf8');
  return btoa(String.fromCharCode(...signature));
}

export async function sendBase64Transaction(txBase64: string): Promise<string> {
  const provider = getPhantom();
  if (!provider?.publicKey) {
    throw new Error('Подключи Phantom');
  }
  const raw = Uint8Array.from(atob(txBase64), (c) => c.charCodeAt(0));
  const tx = Transaction.from(raw);
  const { signature } = await provider.signAndSendTransaction(tx);
  return signature;
}
