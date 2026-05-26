import { useEffect } from 'react';
import { useAppKitEvents } from '@reown/appkit/react';
import { isTelegramWebApp } from './telegram';
import {
  metamaskWalletConnectUrl,
  openWalletHref,
  phantomWalletConnectUrl,
} from './telegram-wallet-bridge';

type RelayTarget = 'phantom' | 'metamask' | null;

/** В Telegram: при появлении WC URI — сразу openLink в кошелёк (обход «Not Detected»). */
export function useTelegramWalletUriRelay(
  active: boolean,
  target: RelayTarget,
): void {
  const events = useAppKitEvents();

  useEffect(() => {
    if (!active || !isTelegramWebApp() || !target) return;

    const props = events?.data?.properties as Record<string, unknown> | undefined;
    const uri =
      (typeof props?.uri === 'string' && props.uri) ||
      (typeof props?.walletConnectUri === 'string' && props.walletConnectUri);

    if (!uri || !uri.includes('wc:')) return;

    const url =
      target === 'phantom'
        ? phantomWalletConnectUrl(uri)
        : metamaskWalletConnectUrl(uri);

    openWalletHref(url);
  }, [active, target, events]);
}
