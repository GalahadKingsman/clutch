import { useCallback, useEffect, useRef, useState } from 'react';
import { fetchMe, phantomBridgePrepare, walletLink, walletNonce } from '../lib/api';
import { SOLANA_WALLET_OPTIONS } from '../lib/appkit-init';
import { useAppKitInit } from './AppKitInitProvider';
import { useClutchWallet, useSolanaWalletConnect } from '../lib/use-clutch-wallet';
import { isTelegramWebApp } from '../lib/telegram';
import {
  connectTelegramSolanaWallet,
  signTelegramSolanaMessage,
  type TelegramWalletTarget,
} from '../lib/telegram-solana-wc';
import { waitFor } from '../lib/wallet-address';

const CONNECT_TIMEOUT_MS = 90_000;

type Props = {
  onLinked: () => void;
};

function toTelegramTarget(
  walletId: (typeof SOLANA_WALLET_OPTIONS)[number]['id'],
): TelegramWalletTarget {
  if (walletId === 'metamask') return 'metamask';
  if (walletId === 'trust') return 'trust';
  return 'phantom';
}

export function WalletGateConnect({ onLinked }: Props) {
  const inTelegram = isTelegramWebApp();
  const { configured, error: initError } = useAppKitInit();
  const { address, isConnected, walletProvider, closeWalletModal } =
    useClutchWallet();
  const [linking, setLinking] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [status, setStatus] = useState<string | null>(null);
  const [pendingAddr, setPendingAddr] = useState<string | null>(null);
  const [awaitingPhantom, setAwaitingPhantom] = useState(false);
  const finishing = useRef(false);
  const providerRef = useRef(walletProvider);
  providerRef.current = walletProvider;

  const cancelConnect = useCallback(() => {
    closeWalletModal();
    setLinking(false);
    setAwaitingPhantom(false);
    setStatus(null);
    setError(null);
  }, [closeWalletModal]);

  const completeLinkFlow = useCallback(
    async (overrideAddress?: string) => {
      const addr = overrideAddress ?? pendingAddr ?? address ?? undefined;
      if (!addr || finishing.current) {
        if (!addr && linking) return;
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
        let signature: string;

        if (inTelegram) {
          const { nonce, message } = await walletNonce();
          const sig = await signTelegramSolanaMessage(message);
          signature = btoa(String.fromCharCode(...sig));
          await walletLink({ wallet_address: addr, signature, nonce });
        } else {
          const provider = await waitFor(() => providerRef.current, 25_000);
          if (!provider.signMessage) {
            throw new Error('Кошелёк не поддерживает подпись сообщений');
          }
          const { nonce, message } = await walletNonce();
          const encoded = new TextEncoder().encode(message);
          const sig = await provider.signMessage(encoded);
          signature = btoa(String.fromCharCode(...sig));
          await walletLink({ wallet_address: addr, signature, nonce });
        }

        onLinked();
      } catch (e) {
        setError(e instanceof Error ? e.message : 'Ошибка привязки');
        setLinking(false);
      } finally {
        finishing.current = false;
      }
    },
    [address, closeWalletModal, inTelegram, linking, onLinked, pendingAddr],
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

  const checkPhantomLinked = useCallback(async () => {
    setError(null);
    try {
      const u = await fetchMe();
      if (u.wallet_linked) {
        setAwaitingPhantom(false);
        setLinking(false);
        onLinked();
        return;
      }
      setError('Кошелёк ещё не привязан. В Phantom нажми «Подключить» и подпиши сообщение.');
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Не удалось проверить');
    }
  }, [onLinked]);

  /** Telegram: Phantom — in-app браузер; остальные — WalletConnect deep link. */
  const connectInTelegram = useCallback(
    async (walletId: (typeof SOLANA_WALLET_OPTIONS)[number]['id']) => {
      closeWalletModal();
      setLinking(true);
      setError(null);
      setAwaitingPhantom(false);

      try {
        if (walletId === 'phantom') {
          setStatus('Открываем CLUTCH в Phantom…');
          const { phantom_url } = await phantomBridgePrepare();
          setAwaitingPhantom(true);
          setStatus(
            '1) Подтверди подключение в Phantom\n2) Подпиши сообщение\n3) Вернись в Telegram → «Проверить»',
          );
          openWalletHref(phantom_url);
          return;
        }

        const target = toTelegramTarget(walletId);
        setStatus(`Открываем ${walletId}…`);
        const addr = await connectTelegramSolanaWallet(target);
        setPendingAddr(addr);
        setStatus('Подпись в кошельке…');

        const { nonce, message } = await walletNonce();
        const sig = await signTelegramSolanaMessage(message);
        const signature = btoa(String.fromCharCode(...sig));

        await walletLink({ wallet_address: addr, signature, nonce });
        onLinked();
      } catch (e) {
        setError(
          e instanceof Error
            ? e.message
            : 'Не удалось подключить кошелёк.',
        );
        setLinking(false);
        setAwaitingPhantom(false);
      }
    },
    [closeWalletModal, onLinked],
  );

  useEffect(() => {
    if (!inTelegram && linking && isConnected && address && !finishing.current) {
      setPendingAddr(address);
      void completeLinkFlow(address);
    }
  }, [inTelegram, linking, isConnected, address, completeLinkFlow]);

  useEffect(() => {
    if (!linking || inTelegram) return;

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
  }, [linking, inTelegram, isConnected, address, completeLinkFlow]);

  useEffect(() => {
    if (!linking || inTelegram) return;
    if (!isPending) return;

    const timer = window.setTimeout(() => {
      closeWalletModal();
      setLinking(false);
      setError('Таймаут подключения. Попробуй снова.');
      setStatus(null);
    }, CONNECT_TIMEOUT_MS);

    return () => window.clearTimeout(timer);
  }, [linking, inTelegram, isPending, closeWalletModal]);

  function pickWallet(walletId: (typeof SOLANA_WALLET_OPTIONS)[number]['id']) {
    if (!configured) {
      setError(initError ?? 'WalletConnect не настроен');
      return;
    }

    if (inTelegram) {
      void connectInTelegram(walletId);
      return;
    }

    setError(null);
    setLinking(true);

    if (isConnected && address) {
      void completeLinkFlow(address);
      return;
    }

    if (walletId === 'trust') {
      setStatus('Trust может не поддержать Solana devnet — попробуй Phantom');
    } else {
      const label =
        SOLANA_WALLET_OPTIONS.find((w) => w.id === walletId)?.label ?? walletId;
      setStatus(`Открываем ${label}…`);
    }

    connectWallet(walletId);
  }

  const busy = linking || (!inTelegram && isPending);
  const showContinue =
    !inTelegram &&
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
            disabled={busy || !configured || (!inTelegram && !isReady)}
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

      {awaitingPhantom && (
        <button
          type="button"
          className="mt-4 w-full max-w-sm rounded-xl bg-green py-3 text-sm font-extrabold text-[#053022]"
          onClick={() => void checkPhantomLinked()}
        >
          Проверить привязку
        </button>
      )}

      <p className="mt-4 text-xs text-mut">
        {inTelegram ? (
          <>
            Phantom: откроется <strong className="text-ink">сайт CLUTCH внутри Phantom</strong>
            — там будет запрос на подключение (как в обычном dApp).
          </>
        ) : (
          <>
            Рекомендуем <strong className="text-ink">Phantom</strong> для devnet.
            Trust часто не поддерживает devnet.
          </>
        )}
      </p>

      {linking && (
        <button
          type="button"
          className="mt-4 text-sm font-semibold text-mut underline"
          onClick={cancelConnect}
        >
          Отменить
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
