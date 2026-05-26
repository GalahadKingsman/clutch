import { useEffect, useRef, useState } from 'react';
import { walletLink, walletNonce } from '../lib/api';
import { walletConnectConfigured } from '../lib/appkit-config';
import { useClutchWallet } from '../lib/use-clutch-wallet';

type Props = {
  onLinked: () => void;
};

export function WalletGate({ onLinked }: Props) {
  const { address, isConnected, openWalletModal, signAuthMessage, walletProvider } =
    useClutchWallet();
  const [linking, setLinking] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [status, setStatus] = useState<string | null>(null);
  const finishing = useRef(false);

  async function finishLink() {
    if (finishing.current || !address || !walletProvider) return;
    finishing.current = true;
    setStatus('Подпись в кошельке…');
    try {
      const { nonce, message } = await walletNonce();
      const signature = await signAuthMessage(message);
      await walletLink({
        wallet_address: address,
        signature,
        nonce,
      });
      onLinked();
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Ошибка привязки');
      setLinking(false);
    } finally {
      finishing.current = false;
    }
  }

  useEffect(() => {
    if (linking && isConnected && address && walletProvider) {
      void finishLink();
    }
  }, [linking, isConnected, address, walletProvider]); // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    if (!linking || isConnected) return;
    const t = window.setTimeout(() => {
      setLinking(false);
      setStatus(null);
      setError('Не удалось подключить кошелёк. Попробуй ещё раз.');
    }, 120_000);
    return () => window.clearTimeout(t);
  }, [linking, isConnected]);

  function startLink() {
    if (!walletConnectConfigured()) {
      setError(
        'Не задан VITE_WALLETCONNECT_PROJECT_ID. Добавь в .env и пересобери nginx.',
      );
      return;
    }
    setError(null);
    setLinking(true);
    if (isConnected && address) {
      setStatus('Подпись в кошельке…');
      void finishLink();
      return;
    }
    setStatus('Выбери кошелёк в окне WalletConnect…');
    openWalletModal();
  }

  return (
    <div className="flex min-h-screen flex-col items-center justify-center px-6 py-10 text-center">
      <p className="font-display text-2xl font-bold">clutch</p>
      <h1 className="mt-6 text-xl font-bold">Привязать кошелёк</h1>
      <p className="mt-3 max-w-xs text-sm font-semibold text-mut">
        Без кошелька CLUTCH недоступен. Подключи Solana-кошелёк через
        WalletConnect — Phantom, Trust и другие.
      </p>

      <button
        type="button"
        disabled={linking && !error}
        onClick={() => startLink()}
        className="mt-8 w-full max-w-sm rounded-2xl bg-gradient-to-b from-[#5C88FF] to-[#4068E8] py-4 text-base font-extrabold text-white shadow-[0_5px_0_#2E51C4] disabled:opacity-60"
      >
        {linking ? 'Подключение…' : 'Подключить кошелёк'}
      </button>

      <p className="mt-4 text-xs text-mut">
        1) Откроется окно WalletConnect → выбери Phantom (или другой)
        <br />
        2) Подтверди подключение и подпись в кошельке
      </p>

      {status && !error && (
        <p className="mt-3 text-sm font-semibold text-gold">{status}</p>
      )}

      {error && (
        <p className="mt-4 rounded-xl border border-red/30 bg-red/10 px-4 py-3 text-sm text-red">
          {error}
        </p>
      )}
    </div>
  );
}
