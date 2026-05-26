import { isTelegramWebApp } from './telegram';

/** Префиксы native scheme → universal link (для openLink в Telegram). */
const SCHEME_MAP: Array<{ scheme: string; universal: string }> = [
  { scheme: 'metamask://', universal: 'https://metamask.app.link/' },
  { scheme: 'phantom://', universal: 'https://phantom.app/' },
  { scheme: 'trust://', universal: 'https://link.trustwallet.com/' },
];

function isWalletHref(href: string): boolean {
  return (
    href.startsWith('phantom://') ||
    href.startsWith('metamask://') ||
    href.startsWith('trust://') ||
    href.includes('phantom.app') ||
    href.includes('metamask.app.link') ||
    href.includes('link.trustwallet.com') ||
    href.includes('wc?uri=')
  );
}

function isAppStorePhantomLink(href: string): boolean {
  return (
    (href.includes('apps.apple.com') || href.includes('play.google.com')) &&
    href.toLowerCase().includes('phantom')
  );
}

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
 * Reown в Telegram иногда double-encode WC uri → кошелёк не открывается.
 * @see https://github.com/reown-com/appkit/issues/5605
 */
export function fixWalletConnectHref(href: string): string {
  let url = toUniversalWalletUrl(href);
  if (!url.includes('uri=')) return url;

  const idx = url.indexOf('uri=');
  const prefix = url.slice(0, idx + 4);
  let encoded = url.slice(idx + 4);
  const cut = encoded.indexOf('&');
  if (cut >= 0) {
    encoded = encoded.slice(0, cut);
  }

  try {
    let decoded = decodeURIComponent(encoded);
    if (!decoded.startsWith('wc:')) {
      decoded = decodeURIComponent(decoded);
    }
    if (decoded.startsWith('wc:')) {
      return prefix + encodeURIComponent(decoded);
    }
  } catch {
    /* keep original */
  }
  return url;
}

/** Universal link Phantom + WalletConnect (iOS/Android). */
export function phantomWalletConnectUrl(wcUri: string): string {
  const uri = wcUri.startsWith('wc:') ? wcUri : decodeURIComponent(wcUri);
  const appUrl = import.meta.env.VITE_APP_URL || 'https://clutch-duel.ru';
  return `https://phantom.app/ul/browse/${encodeURIComponent(uri)}?ref=${encodeURIComponent(appUrl)}`;
}

export function metamaskWalletConnectUrl(wcUri: string): string {
  const uri = wcUri.startsWith('wc:') ? wcUri : decodeURIComponent(wcUri);
  return `https://metamask.app.link/wc?uri=${encodeURIComponent(uri)}`;
}

export function trustWalletConnectUrl(wcUri: string): string {
  const uri = wcUri.startsWith('wc:') ? wcUri : decodeURIComponent(wcUri);
  return `https://link.trustwallet.com/wc?uri=${encodeURIComponent(uri)}`;
}

/** Открыть ссылку кошелька в Telegram (или fallback). */
export function openWalletHref(href: string): void {
  const url = fixWalletConnectHref(href);
  const tg = window.Telegram?.WebApp;

  if (tg?.openLink) {
    tg.openLink(url, { try_instant_view: false });
    return;
  }

  window.location.assign(url);
}

/**
 * В Telegram Mini App window.open('metamask://…') не работает — сессия WC зависает.
 * Перехватываем open / клики по ссылкам и открываем через Telegram.WebApp.openLink.
 * Вызвать до createAppKit (см. main.tsx).
 */
export function installTelegramWalletBridge(): void {
  if (!isTelegramWebApp()) return;
  const w = window as Window & { __clutchTgWalletBridge?: boolean };
  if (w.__clutchTgWalletBridge) return;
  w.__clutchTgWalletBridge = true;

  const nativeOpen = window.open.bind(window);

  window.open = (
    url?: string | URL,
    target?: string,
    features?: string,
  ): Window | null => {
    if (url == null || url === '') {
      return nativeOpen(url as string, target, features);
    }

    const href =
      typeof url === 'string' ? fixWalletConnectHref(url) : fixWalletConnectHref(url.toString());

    openWalletHref(href);
    return null;
  };

  document.addEventListener(
    'click',
    (e) => {
      const anchor = (e.target as HTMLElement).closest('a');
      if (!anchor?.href) return;

      if (isAppStorePhantomLink(anchor.href)) {
        e.preventDefault();
        e.stopPropagation();
        return;
      }

      const label = (e.target as HTMLElement).textContent?.trim().toLowerCase();
      if (
        (label === 'get' || label === 'get >' || label?.startsWith('get ')) &&
        (e.target as HTMLElement).closest('w3m-modal, w3m-connecting-wc-mobile')
      ) {
        e.preventDefault();
        e.stopPropagation();
        return;
      }

      if (!isWalletHref(anchor.href)) return;
      e.preventDefault();
      e.stopPropagation();
      openWalletHref(anchor.href);
    },
    true,
  );
}

/** В Telegram нельзя полагаться на detect installed — только WC + deep link. */
export function useTelegramDirectWalletConnect(walletId: string): boolean {
  if (!isTelegramWebApp()) return false;
  return walletId === 'phantom' || walletId === 'metamask' || walletId === 'trust';
}

/** Кошельки для UI: в Telegram mobile MetaMask/Trust почти всегда зависают. */
export function shouldHideWalletInTelegram(walletId: string): boolean {
  return false;
}
