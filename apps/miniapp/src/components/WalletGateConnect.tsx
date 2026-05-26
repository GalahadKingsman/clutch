import { useCallback, useEffect, useRef, useState } from 'react';
import { walletLink, walletNonce } from '../lib/api';
import { SOLANA_WALLET_OPTIONS } from '../lib/appkit-init';
import { useAppKitInit } from './AppKitInitProvider';
import { useClutchWallet, useSolanaWalletConnect } from '../lib/use-clutch-wallet';

type Props = {
  onLinked: () => void;
};

export function WalletGateConnect({ onLinked }: Props) {
  const { configured, error: initError } = useAppKitInit();
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
    if (!configured) {
      setError(initError ?? 'WalletConnect не настроен');
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
    <>
      <div className="mt-8 grid w-full max-w-sm grid-cols-4 gap-3">
        {SOLANA_WALLET_OPTIONS.map((w) => (
          <button
            key={w.id}
            type="button"
            disabled={busy || !configured || !isReady}
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
      </p>

      {status && !error && (
        <p className="mt-3 text-sm font-semibold text-gold">{status}</p>
      )}

      {error && (
        <p className="mt-4 rounded-xl border border-red/30 bg-red/10 px-4 py-3 text-sm text-red">
          {error}
        </p>
      )}
    </>
  );
}
