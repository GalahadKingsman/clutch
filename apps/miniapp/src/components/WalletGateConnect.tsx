import { useCallback, useEffect, useRef, useState } from 'react';
import { walletLink, walletNonce } from '../lib/api';
import { SOLANA_WALLET_OPTIONS } from '../lib/appkit-init';
import { useAppKitInit } from './AppKitInitProvider';
import { useClutchWallet, useSolanaWalletConnect } from '../lib/use-clutch-wallet';
import { waitFor } from '../lib/wallet-address';

type Props = {
  onLinked: () => void;
};

export function WalletGateConnect({ onLinked }: Props) {
  const { configured, error: initError } = useAppKitInit();
  const { address, isConnected, walletProvider } = useClutchWallet();
  const [linking, setLinking] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [status, setStatus] = useState<string | null>(null);
  const [pendingAddr, setPendingAddr] = useState<string | null>(null);
  const finishing = useRef(false);
  const providerRef = useRef(walletProvider);
  providerRef.current = walletProvider;

  const completeLinkFlow = useCallback(
    async (overrideAddress?: string) => {
      const addr = overrideAddress ?? pendingAddr ?? address ?? undefined;
      if (!addr || finishing.current) {
        if (!addr) {
          setError(
            'Адрес Solana не получен. Подключи кошелёк с поддержкой Solana (лучше Phantom).',
          );
          setLinking(false);
        }
        return;
      }

      finishing.current = true;
      setError(null);
      setStatus('Подпись в кошельке…');

      try {
        const provider = await waitFor(() => providerRef.current, 25_000);
        if (!provider.signMessage) {
          throw new Error('Кошелёк не поддерживает подпись сообщений');
        }

        const { nonce, message } = await walletNonce();
        const encoded = new TextEncoder().encode(message);
        const sig = await provider.signMessage(encoded);
        const signature = btoa(String.fromCharCode(...sig));

        await walletLink({
          wallet_address: addr,
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
    },
    [address, onLinked, pendingAddr],
  );

  const onWalletConnected = useCallback(
    (addr: string | null) => {
      if (addr) setPendingAddr(addr);
      setStatus('Подпись в кошельке…');
      void completeLinkFlow(addr ?? undefined);
    },
    [completeLinkFlow],
  );

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
    if (linking && isConnected && address && !finishing.current) {
      setPendingAddr(address);
      void completeLinkFlow(address);
    }
  }, [linking, isConnected, address, completeLinkFlow]);

  function pickWallet(walletId: (typeof SOLANA_WALLET_OPTIONS)[number]['id']) {
    if (!configured) {
      setError(initError ?? 'WalletConnect не настроен');
      return;
    }
    if (walletId === 'trust') {
      setStatus('Trust может не поддержать Solana devnet — попробуй Phantom');
    }
    setError(null);
    setLinking(true);
    if (isConnected && address) {
      void completeLinkFlow(address);
      return;
    }
    const label =
      SOLANA_WALLET_OPTIONS.find((w) => w.id === walletId)?.label ?? walletId;
    setStatus(`Открываем ${label}…`);
    connectWallet(walletId);
  }

  const busy = linking || isPending;
  const showContinue =
    linking &&
    (pendingAddr || address) &&
    !finishing.current &&
    status?.includes('Подпись');

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
            {'hint' in w && w.hint && (
              <span className="text-[9px] text-gold/80">{w.hint}</span>
            )}
          </button>
        ))}
      </div>

      <p className="mt-4 text-xs text-mut">
        Рекомендуем <strong className="text-ink">Phantom</strong> для devnet.
        Trust часто не поддерживает devnet (ошибка chains).
      </p>

      {showContinue && (
        <button
          type="button"
          className="mt-4 w-full max-w-sm rounded-xl bg-green py-3 text-sm font-extrabold text-[#053022]"
          onClick={() => void completeLinkFlow()}
        >
          Продолжить привязку
        </button>
      )}

      {status && !error && (
        <p className="mt-3 text-sm font-semibold text-gold">{status}</p>
      )}

      {(pendingAddr || address) && linking && (
        <p className="mt-2 break-all text-[10px] text-mut">
          {pendingAddr ?? address}
        </p>
      )}

      {error && (
        <p className="mt-4 rounded-xl border border-red/30 bg-red/10 px-4 py-3 text-sm text-red">
          {error}
        </p>
      )}
    </>
  );
}
