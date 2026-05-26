import { useCallback, useEffect, useRef, useState } from 'react';
import { walletLink, walletNonce } from '../lib/api';
import {
  SOLANA_WALLET_OPTIONS,
  walletConnectConfigured,
} from '../lib/appkit-config';
import { useClutchWallet, useSolanaWalletConnect } from '../lib/use-clutch-wallet';

type Props = {
  onLinked: () => void;
};

export function WalletGate({ onLinked }: Props) {
  const { address, isConnected, signAuthMessage, walletProvider } =
    useClutchWallet();
  const [linking, setLinking] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [status, setStatus] = useState<string | null>(null);
  const finishing = useRef(false);

  const finishLink = useCallback(async () => {
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
  }, [address, onLinked, signAuthMessage, walletProvider]);

  const onWalletConnected = useCallback(() => {
    setStatus('Подпись в кошельке…');
  }, []);

  const onWalletError = useCallback((msg: string) => {
    setError(msg);
    setLinking(false);
    setStatus(null);
  }, []);

  const { connectWallet, isReady, isPending } = useSolanaWalletConnect({
    onSuccess: onWalletConnected,
    onError: onWalletError,
  });

  useEffect(() => {
    if (linking && isConnected && address && walletProvider) {
      void finishLink();
    }
  }, [linking, isConnected, address, walletProvider, finishLink]);

  function pickWallet(walletId: (typeof SOLANA_WALLET_OPTIONS)[number]['id']) {
    if (!walletConnectConfigured()) {
      setError(
        'Нет VITE_WALLETCONNECT_PROJECT_ID в .env. Пересобери nginx после правки.',
      );
      return;
    }
    setError(null);
    setLinking(true);
    if (isConnected && address) {
      void finishLink();
      return;
    }
    const label =
      SOLANA_WALLET_OPTIONS.find((w) => w.id === walletId)?.label ?? walletId;
    setStatus(`Открываем ${label}…`);
    connectWallet(walletId);
  }

  const busy = linking || isPending;

  return (
    <div className="flex min-h-screen flex-col items-center justify-center px-6 py-10 text-center">
      <p className="font-display text-2xl font-bold">clutch</p>
      <h1 className="mt-6 text-xl font-bold">Привязать кошелёк</h1>
      <p className="mt-3 max-w-xs text-sm font-semibold text-mut">
        Выбери кошелёк Solana. Подключение через WalletConnect — как на gmgn.
      </p>

      {!walletConnectConfigured() && (
        <p className="mt-4 rounded-xl border border-red/30 bg-red/10 px-3 py-2 text-xs text-red">
          Project ID не в сборке. Добавь VITE_WALLETCONNECT_PROJECT_ID в .env и
          выполни build nginx.
        </p>
      )}

      {walletConnectConfigured() && !isReady && (
        <p className="mt-4 text-xs text-gold">Загрузка WalletConnect…</p>
      )}

      <div className="mt-8 grid w-full max-w-sm grid-cols-4 gap-3">
        {SOLANA_WALLET_OPTIONS.map((w) => (
          <button
            key={w.id}
            type="button"
            disabled={busy || !walletConnectConfigured()}
            onClick={() => pickWallet(w.id)}
            className="flex flex-col items-center gap-2 disabled:opacity-50"
          >
            <span className="flex h-14 w-14 items-center justify-center overflow-hidden rounded-2xl border border-white/10 bg-panel2">
              <img
                src={w.icon}
                alt=""
                className="h-10 w-10 rounded-xl object-cover"
              />
            </span>
            <span className="text-[10px] font-bold leading-tight text-mut">
              {w.label}
            </span>
          </button>
        ))}
      </div>

      <p className="mt-6 text-xs text-mut">
        Phantom / Trust откроются в приложении кошелька.
        <br />
        «QR / Другие» — список всех кошельков WalletConnect.
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
