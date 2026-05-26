import UniversalProvider from '@walletconnect/universal-provider';
import bs58 from 'bs58';
import {
  metamaskWalletConnectUrl,
  openWalletHref,
  phantomWalletConnectUrl,
  trustWalletConnectUrl,
} from './telegram-wallet-bridge';

const SOLANA_DEVNET = 'solana:EtWTRABZaYq6iMfeYKouRuYVU6x099TC';
const SOLANA_MAINNET = 'solana:5eykt4UsFv8P8NJfTyeV4uEKm7aRfx5Fm7N8kTteij';

export type TelegramWalletTarget = 'phantom' | 'metamask' | 'trust';

let provider: UniversalProvider | null = null;

function solanaChainId(): string {
  return import.meta.env.VITE_SOLANA_NETWORK === 'mainnet-beta'
    ? SOLANA_MAINNET
    : SOLANA_DEVNET;
}

function appMetadata() {
  const appUrl = import.meta.env.VITE_APP_URL || 'https://clutch-duel.ru';
  return {
    name: 'CLUTCH',
    description: '1v1 crypto duels on Solana',
    url: appUrl,
    icons: [`${appUrl}/favicon.svg`],
  };
}

function walletDeepLink(target: TelegramWalletTarget, wcUri: string): string {
  switch (target) {
    case 'metamask':
      return metamaskWalletConnectUrl(wcUri);
    case 'trust':
      return trustWalletConnectUrl(wcUri);
    default:
      return phantomWalletConnectUrl(wcUri);
  }
}

function addressFromSession(
  session: UniversalProvider['session'],
): string | null {
  const accounts = session?.namespaces?.solana?.accounts;
  if (!accounts?.length) return null;
  const parts = accounts[0].split(':');
  return parts.length >= 3 ? parts[2] : null;
}

async function getProvider(): Promise<UniversalProvider> {
  if (provider) return provider;

  const projectId = import.meta.env.VITE_WALLETCONNECT_PROJECT_ID;
  if (!projectId) {
    throw new Error('VITE_WALLETCONNECT_PROJECT_ID не задан');
  }

  provider = await UniversalProvider.init({
    projectId,
    metadata: appMetadata(),
  });
  return provider;
}

/** Подключение в Telegram без модалки Reown «Not Detected». */
export async function connectTelegramSolanaWallet(
  target: TelegramWalletTarget,
): Promise<string> {
  const p = await getProvider();
  const chain = solanaChainId();

  const existing = addressFromSession(p.session);
  if (existing) return existing;

  return new Promise((resolve, reject) => {
    const onUri = (uri: string) => {
      openWalletHref(walletDeepLink(target, uri));
    };

    const cleanup = () => {
      if (typeof p.removeListener === 'function') {
        p.removeListener('display_uri', onUri);
      }
    };

    p.on('display_uri', onUri);

    p.connect({
      optionalNamespaces: {
        solana: {
          methods: ['solana_signMessage', 'solana_signTransaction'],
          chains: [chain],
          events: [],
        },
      },
    })
      .then((session) => {
        cleanup();
        const addr = addressFromSession(session);
        if (!addr) {
          reject(new Error('Solana-адрес не получен от кошелька'));
          return;
        }
        resolve(addr);
      })
      .catch((err) => {
        cleanup();
        reject(err instanceof Error ? err : new Error('Ошибка WalletConnect'));
      });
  });
}

function parseSignatureBytes(result: unknown): Uint8Array {
  if (result instanceof Uint8Array) return result;
  if (typeof result === 'string') {
    try {
      return bs58.decode(result);
    } catch {
      return Uint8Array.from(atob(result), (c) => c.charCodeAt(0));
    }
  }
  if (typeof result === 'object' && result !== null && 'signature' in result) {
    const sig = String((result as { signature: string }).signature);
    try {
      return bs58.decode(sig);
    } catch {
      return Uint8Array.from(atob(sig), (c) => c.charCodeAt(0));
    }
  }
  throw new Error('Не удалось прочитать подпись кошелька');
}

export async function signTelegramSolanaMessage(
  message: string,
): Promise<Uint8Array> {
  const p = await getProvider();
  const chain = solanaChainId();
  const addr = addressFromSession(p.session);
  if (!addr) throw new Error('Кошелёк не подключён');

  const encoded = new TextEncoder().encode(message);
  const result = await p.request(
    {
      method: 'solana_signMessage',
      params: {
        pubkey: addr,
        message: bs58.encode(encoded),
      },
    },
    chain,
  );

  return parseSignatureBytes(result);
}

export async function disconnectTelegramSolanaWallet(): Promise<void> {
  if (!provider) return;
  try {
    await provider.disconnect();
  } catch {
    /* ignore */
  }
  provider = null;
}
