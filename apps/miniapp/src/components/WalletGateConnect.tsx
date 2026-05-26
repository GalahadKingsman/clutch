import { useCallback, useEffect, useRef, useState } from 'react';
import { walletLink, walletNonce } from '../lib/api';
import { SOLANA_WALLET_OPTIONS } from '../lib/appkit-init';
import { useAppKitInit } from './AppKitInitProvider';
import { useClutchWallet, useSolanaWalletConnect } from '../lib/use-clutch-wallet';
import { isTelegramWebApp } from '../lib/telegram';
import {
  openWalletHref,
  useTelegramDirectWalletConnect,
} from '../lib/telegram-wallet-bridge';
import { watchAndOpenPhantom } from '../lib/telegram-wc-watcher';
import { useTelegramWalletUriRelay } from '../lib/use-telegram-wc-relay';
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
  const [relayWallet, setRelayWallet] = useState<'phantom' | 'metamask' | null>(
    null,
  );
  const finishing = useRef(false);
  const stopWcWatch = useRef<(() => void) | null>(null);
  const providerRef = useRef(walletProvider);
  providerRef.current = walletProvider;

  const cancelConnect = useCallback(() => {
    stopWcWatch.current?.();
    stopWcWatch.current = null;
    closeWalletModal();
    setLinking(false);
    setRelayWallet(null);
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
      setRelayWallet(null);
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

  useTelegramWalletUriRelay(linking || isPending, relayWallet);

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
    setError(null);
    setLinking(true);
    setRelayWallet(
      walletId === 'phantom' || walletId === 'metamask' ? walletId : null,
    );

    if (isConnected && address) {
      void completeLinkFlow(address);
      return;
    }

    if (walletId === 'trust') {
      setStatus('Trust может не поддержать Solana devnet — попробуй Phantom');
    }

    const label =
      SOLANA_WALLET_OPTIONS.find((w) => w.id === walletId)?.label ?? walletId;
    if (walletId !== 'trust') {
      setStatus(`Открываем ${label}…`);
    }

    if (walletId === 'walletConnect' && isTelegramWebApp()) {
      setStatus('Сканируй QR в Phantom (Настройки → WalletConnect)');
    }

    /** В Telegram connect('phantom') → ложный «Not Detected» и App Store. */
    if (useTelegramDirectWalletConnect(walletId)) {
      setStatus(
        walletId === 'phantom'
          ? 'Открываем Phantom… (установленный кошелёк)'
          : `Открываем ${label}…`,
      );
      stopWcWatch.current?.();
      if (walletId === 'phantom') {
        stopWcWatch.current = watchAndOpenPhantom();
      }
      connectWallet('walletConnect');
      return;
    }

    connectWallet(walletId);
  }

  const busy = linking || isPending;
  const showContinue =
    linking &&
    (pendingAddr || address) &&
    !finishing.current &&
    !isPending;

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
        {isTelegramWebApp() ? (
          <>
            В Telegram нажми <strong className="text-ink">Phantom</strong> — откроется
            установленное приложение (не App Store). Если не сработало — кнопка ниже
            или QR.
          </>
        ) : (
          <>
            Рекомендуем <strong className="text-ink">Phantom</strong> для devnet.
            Trust часто не поддерживает devnet.
          </>
        )}
      </p>

      {isTelegramWebApp() && (linking || isPending) && (
        <button
          type="button"
          className="mt-4 w-full max-w-sm rounded-xl border border-gold/40 bg-gold/10 py-3 text-sm font-bold text-gold"
          onClick={() => {
            stopWcWatch.current?.();
            stopWcWatch.current = watchAndOpenPhantom();
            const anchors = document.querySelectorAll<HTMLAnchorElement>(
              'a[href*="wc"], a[href*="phantom.app"]',
            );
            for (const a of anchors) {
              if (a.href) {
                openWalletHref(a.href);
                return;
              }
            }
            setStatus('Ищем сессию… откроется Phantom через 1–2 сек');
          }}
        >
          Открыть установленный Phantom
        </button>
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
