import { useEffect, useState } from 'react';
import {
  phantomBridgeLink,
  phantomBridgeSession,
} from '../lib/api';

type PhantomSolana = {
  isPhantom?: boolean;
  connect: () => Promise<{ publicKey: { toBase58: () => string } }>;
  signMessage: (
    message: Uint8Array,
    display?: string,
  ) => Promise<{ signature: Uint8Array }>;
};

function getPhantom(): PhantomSolana | null {
  const p = (
    window as Window & { phantom?: { solana?: PhantomSolana } }
  ).phantom?.solana;
  return p?.isPhantom ? p : p ?? null;
}

export function TgWalletPage() {
  const [status, setStatus] = useState('Подключение…');
  const [error, setError] = useState<string | null>(null);
  const [done, setDone] = useState(false);

  useEffect(() => {
    const token = new URLSearchParams(window.location.search).get('token');
    if (!token) {
      setError('Нет токена сессии. Открой ссылку из CLUTCH в Telegram.');
      return;
    }

    void (async () => {
      try {
        const phantom = getPhantom();
        if (!phantom) {
          setError(
            'Открой эту страницу через кнопку Phantom в CLUTCH (встроенный браузер Phantom).',
          );
          return;
        }

        setStatus('Запрос разрешения в Phantom…');
        const { publicKey } = await phantom.connect();
        const address = publicKey.toBase58();

        setStatus('Подпись сообщения…');
        const { nonce, message } = await phantomBridgeSession(token);
        const encoded = new TextEncoder().encode(message);
        const { signature } = await phantom.signMessage(encoded, 'utf8');
        const signatureB64 = btoa(String.fromCharCode(...signature));

        setStatus('Сохранение…');
        await phantomBridgeLink({
          token,
          wallet_address: address,
          signature: signatureB64,
          nonce,
        });

        setDone(true);
        setStatus(`Кошелёк ${address.slice(0, 4)}…${address.slice(-4)} привязан.`);
      } catch (e) {
        setError(e instanceof Error ? e.message : 'Ошибка привязки');
      }
    })();
  }, []);

  return (
    <div className="flex min-h-screen flex-col items-center justify-center bg-[#0f0e16] px-6 py-10 text-center">
      <p className="font-display text-2xl font-bold text-ink">clutch</p>
      <h1 className="mt-6 text-lg font-bold text-ink">Привязка в Phantom</h1>

      {error && (
        <p className="mt-4 rounded-xl border border-red/30 bg-red/10 px-4 py-3 text-sm text-red">
          {error}
        </p>
      )}

      {!error && (
        <p className="mt-4 text-sm font-semibold text-gold">{status}</p>
      )}

      {done && (
        <p className="mt-6 max-w-xs text-sm text-mut">
          Вернись в <strong className="text-ink">Telegram</strong>, открой CLUTCH и
          нажми «Проверить привязку».
        </p>
      )}
    </div>
  );
}
