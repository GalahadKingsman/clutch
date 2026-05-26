import { useState } from 'react';
import { walletLink, walletNonce } from '../lib/api';
import { walletConnectConfigured } from '../lib/wallet-provider';
import { useClutchWallet } from '../lib/use-clutch-wallet';

type Props = {
  onLinked: () => void;
};

export function WalletGate({ onLinked }: Props) {
  const { signAuthMessage, ensureConnected, connecting } = useClutchWallet();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function linkWallet() {
    if (!walletConnectConfigured()) {
      setError(
        'Не задан VITE_WALLETCONNECT_PROJECT_ID. Создай проект на cloud.reown.com',
      );
      return;
    }

    setLoading(true);
    setError(null);
    try {
      const pk = await ensureConnected();
      const { nonce, message } = await walletNonce();
      const signature = await signAuthMessage(message);
      await walletLink({
        wallet_address: pk.toBase58(),
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

  const busy = loading || connecting;

  return (
    <div className="flex min-h-screen flex-col items-center justify-center px-6 py-10 text-center">
      <p className="font-display text-2xl font-bold">clutch</p>
      <h1 className="mt-6 text-xl font-bold">Привязать кошелёк</h1>
      <p className="mt-3 max-w-xs text-sm font-semibold text-mut">
        Без кошелька CLUTCH недоступен. Подключи Solana-кошелёк через
        WalletConnect — Phantom, Trust, MetaMask и другие.
      </p>

      <button
        type="button"
        disabled={busy}
        onClick={() => void linkWallet()}
        className="mt-8 w-full max-w-sm rounded-2xl bg-gradient-to-b from-[#5C88FF] to-[#4068E8] py-4 text-base font-extrabold text-white shadow-[0_5px_0_#2E51C4] disabled:opacity-60"
      >
        {busy ? 'Подключение…' : 'Подключить кошелёк'}
      </button>

      <p className="mt-4 text-xs text-mut">
        Откроется выбор кошелька (WalletConnect). На телефоне — переход в
        приложение кошелька для подписи.
      </p>

      {error && (
        <p className="mt-4 rounded-xl border border-red/30 bg-red/10 px-4 py-3 text-sm text-red">
          {error}
        </p>
      )}
    </div>
  );
}
