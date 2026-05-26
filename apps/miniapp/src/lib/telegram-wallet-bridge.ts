import { isTelegramMobile, isTelegramWebApp } from './telegram';

/** Префиксы native scheme → universal link (для openLink в Telegram). */
const SCHEME_MAP: Array<{ scheme: string; universal: string }> = [
  { scheme: 'metamask://', universal: 'https://metamask.app.link/' },
  { scheme: 'phantom://', universal: 'https://phantom.app/' },
  { scheme: 'trust://', universal: 'https://link.trustwallet.com/' },
  { scheme: 'wc://', universal: 'https://walletconnect.com/wc?' },
];

function toUniversalWalletUrl(raw: string): string {
  if (raw.startsWith('https://') || raw.startsWith('http://')) {
    return raw;
  }
  for (const { scheme, universal } of SCHEME_MAP) {
    if (raw.startsWith(scheme)) {
      return universal + raw.slice(scheme.length);
    }
  }
  return raw;
}

/**
 * В Telegram Mini App window.open('metamask://…') не работает — сессия WC зависает.
 * Перехватываем open и открываем universal link через Telegram.WebApp.openLink.
 * Вызвать до createAppKit (см. main.tsx).
 */
export function installTelegramWalletBridge(): void {
  if (!isTelegramWebApp()) return;
  const w = window as Window & { __clutchTgWalletBridge?: boolean };
  if (w.__clutchTgWalletBridge) return;
  w.__clutchTgWalletBridge = true;

  const tg = window.Telegram?.WebApp;
  const nativeOpen = window.open.bind(window);

  window.open = (
    url?: string | URL,
    target?: string,
    features?: string,
  ): Window | null => {
    if (url == null || url === '') {
      return nativeOpen(url as string, target, features);
    }

    let href = typeof url === 'string' ? url : url.toString();
    href = toUniversalWalletUrl(href);

    if (tg?.openLink) {
      tg.openLink(href, { try_instant_view: false });
      return null;
    }

    return nativeOpen(href, target, features);
  };
}

/** Кошельки для UI: в Telegram mobile MetaMask/Trust почти всегда зависают. */
export function shouldHideWalletInTelegram(walletId: string): boolean {
  if (!isTelegramWebApp() || !isTelegramMobile()) return false;
  return walletId === 'metamask' || walletId === 'trust';
}
