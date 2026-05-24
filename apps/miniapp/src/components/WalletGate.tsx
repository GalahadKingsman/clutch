import { useState } from 'react';
import { walletLink, walletNonce } from '../lib/api';
import { connectPhantom, signMessagePhantom } from '../lib/solana';

type Props = {
  onLinked: () => void;
};

export function WalletGate({ onLinked }: Props) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function linkPhantom() {
    setLoading(true);
    setError(null);
    try {
      const address = await connectPhantom();
      const { nonce, message } = await walletNonce();
      const signature = await signMessagePhantom(message);
      await walletLink({
        wallet_address: address,
        signature,
        nonce,
      });
      onLinked();
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Ошибка привязки');
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="flex min-h-screen flex-col items-center justify-center px-6 py-10 text-center">
      <p className="font-display text-2xl font-bold">clutch</p>
      <h1 className="mt-6 text-xl font-bold">Привязать кошелёк</h1>
      <p className="mt-3 max-w-xs text-sm font-semibold text-mut">
        Без кошелька CLUTCH недоступен. Подключи Phantom (Solana) для дуэлей и
        ставок.
      </p>

      <button
        type="button"
        disabled={loading}
        onClick={linkPhantom}
        className="mt-8 w-full max-w-sm rounded-2xl bg-gradient-to-b from-[#5C88FF] to-[#4068E8] py-4 text-base font-extrabold text-white shadow-[0_5px_0_#2E51C4] disabled:opacity-60"
      >
        {loading ? 'Подключение…' : 'Phantom'}
      </button>

      <p className="mt-4 text-xs text-mut">
        Trust / MetaMask — в Phase 1 (WalletConnect)
      </p>

      {error && (
        <p className="mt-4 rounded-xl border border-red/30 bg-red/10 px-4 py-3 text-sm text-red">
          {error}
        </p>
      )}
    </div>
  );
}
