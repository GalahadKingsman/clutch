import { useCallback, useEffect, useRef, useState } from 'react';
import { walletLink, walletNonce } from '../lib/api';
import { SOLANA_WALLET_OPTIONS } from '../lib/appkit-init';
import { useAppKitInit } from './AppKitInitProvider';
import { useClutchWallet, useSolanaWalletConnect } from '../lib/use-clutch-wallet';
import { isTelegramMobile, isTelegramWebApp } from '../lib/telegram';
import { shouldHideWalletInTelegram } from '../lib/telegram-wallet-bridge';
import { waitFor } from '../lib/wallet-address';

const CONNECT_TIMEOUT_MS = 90_000;

type Props = {
  onLinked: () => void;
};

export function WalletGateConnect({ onLinked }: Props) {
  const { configured, error: initError } = useAppKitInit();
  const { address, isConnected, walletProvider, closeWalletModal } =
    useClutchWallet();
  const [linking, setLinking] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [status, setStatus] = useState<string | null>(null);
  const [pendingAddr, setPendingAddr] = useState<string | null>(null);
  const finishing = useRef(false);
  const providerRef = useRef(walletProvider);
  providerRef.current = walletProvider;

  const walletOptions = SOLANA_WALLET_OPTIONS.filter(
    (w) => !shouldHideWalletInTelegram(w.id),
  );

  const cancelConnect = useCallback(() => {
    closeWalletModal();
    setLinking(false);
    setStatus(null);
    setError(null);
  }, [closeWalletModal]);

  const completeLinkFlow = useCallback(
    async (overrideAddress?: string) => {
      const addr = overrideAddress ?? pendingAddr ?? address ?? undefined;
      if (!addr || finishing.current) {
        if (!addr && linking) {
          return;
        }
        if (!addr) {
          setError(
            'Адрес Solana не получен. Подключи Phantom или включи Solana в MetaMask.',
          );
          setLinking(false);
        }
        return;
      }

      finishing.current = true;
      setError(null);
      setStatus('Подпись в кошельке…');
      closeWalletModal();

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
    [address, closeWalletModal, linking, onLinked, pendingAddr],
  );

  const onWalletConnected = useCallback(
    (addr: string | null) => {
      if (addr) setPendingAddr(addr);
      setStatus('Подпись в кошельке…');
      void completeLinkFlow(addr ?? undefined);
    },
    [completeLinkFlow],
  );

  const onWalletError = useCallback(
    (msg: string) => {
      closeWalletModal();
      setError(msg);
      setLinking(false);
      setStatus(null);
    },
    [closeWalletModal],
  );

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

  /** После возврата из кошелька в Telegram — догоняем сессию WC. */
  useEffect(() => {
    if (!linking) return;

    const onVisible = () => {
      if (document.visibilityState !== 'visible') return;
      window.setTimeout(() => {
        if (isConnected && address && !finishing.current) {
          void completeLinkFlow(address);
        }
      }, 600);
    };

    document.addEventListener('visibilitychange', onVisible);
    return () => document.removeEventListener('visibilitychange', onVisible);
  }, [linking, isConnected, address, completeLinkFlow]);

  /** Таймаут «бесконечной» модалки Reown. */
  useEffect(() => {
    if (!linking && !isPending) return;

    const timer = window.setTimeout(() => {
      closeWalletModal();
      setLinking(false);
      setError(
        isTelegramWebApp()
          ? 'Не удалось подключить кошелёк. Полностью закрой приложение кошелька, открой CLUTCH снова и нажми Phantom. MetaMask в Telegram часто зависает.'
          : 'Таймаут подключения. Попробуй снова или выбери другой кошелёк.',
      );
      setStatus(null);
    }, CONNECT_TIMEOUT_MS);

    return () => window.clearTimeout(timer);
  }, [linking, isPending, closeWalletModal]);

  function pickWallet(walletId: (typeof SOLANA_WALLET_OPTIONS)[number]['id']) {
    if (!configured) {
      setError(initError ?? 'WalletConnect не настроен');
      return;
    }
    if (shouldHideWalletInTelegram(walletId)) {
      setError('В Telegram используй Phantom или QR.');
      return;
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

    if (walletId === 'walletConnect' && isTelegramWebApp()) {
      setStatus('Сканируй QR в Phantom (Настройки → WalletConnect)');
    }

    connectWallet(walletId);
  }

  const busy = linking || isPending;
  const showContinue =
    linking &&
    (pendingAddr || address) &&
    !finishing.current &&
    !isPending;

  const inTgMobile = isTelegramWebApp() && isTelegramMobile();

  return (
    <>
      <div
        className={`mt-8 grid w-full max-w-sm gap-3 ${
          walletOptions.length <= 2 ? 'grid-cols-2' : 'grid-cols-4'
        }`}
      >
        {walletOptions.map((w) => (
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

      {inTgMobile && (
        <p className="mt-4 max-w-sm text-xs text-mut">
          В Telegram Mini App надёжно работает{' '}
          <strong className="text-ink">Phantom</strong>. MetaMask и Trust здесь
          скрыты — у них ломается возврат из приложения кошелька (бесконечная
          загрузка).
        </p>
      )}

      {!inTgMobile && (
        <p className="mt-4 text-xs text-mut">
          Для devnet лучше Phantom. Trust часто не поддерживает devnet.
        </p>
      )}

      {(linking || isPending) && (
        <button
          type="button"
          className="mt-4 text-sm font-semibold text-mut underline"
          onClick={cancelConnect}
        >
          Отменить подключение
        </button>
      )}

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
