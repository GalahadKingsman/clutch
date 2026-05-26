import { createAppKit } from '@reown/appkit/react';
import { SolanaAdapter } from '@reown/appkit-adapter-solana/react';
import { solana, solanaDevnet } from '@reown/appkit/networks';
import { isTelegramWebApp } from './telegram';

const projectId = import.meta.env.VITE_WALLETCONNECT_PROJECT_ID || '';
const appUrl = import.meta.env.VITE_APP_URL || 'https://clutch-duel.ru';
const useMainnet = import.meta.env.VITE_SOLANA_NETWORK === 'mainnet-beta';

/** Trust часто не поддерживает devnet — для devnet основной кошелёк Phantom. */
export const appKitNetworks = useMainnet ? [solana] : [solanaDevnet, solana];
export const appKitDefaultNetwork = useMainnet ? solana : solanaDevnet;

export const walletConnectConfigured = projectId.length > 0;

let initialized = false;
let initError: string | null = null;

/** Вызвать один раз перед хуками Reown. Не импортировать на уровне main.tsx. */
export function initAppKit(): { ok: boolean; error?: string } {
  if (!walletConnectConfigured) {
    return { ok: false, error: 'VITE_WALLETCONNECT_PROJECT_ID не задан' };
  }
  if (initialized) {
    return { ok: true };
  }
  if (initError) {
    return { ok: false, error: initError };
  }

  try {
    const solanaAdapter = new SolanaAdapter();
    createAppKit({
      adapters: [solanaAdapter],
      networks: appKitNetworks,
      defaultNetwork: appKitDefaultNetwork,
      projectId,
      metadata: {
        name: 'CLUTCH',
        description: '1v1 crypto duels on Solana',
        url: appUrl,
        icons: [`${appUrl}/favicon.svg`],
      },
      features: {
        analytics: false,
        email: false,
        socials: false,
      },
      enableWalletConnect: true,
      ...(isTelegramWebApp() && {
        enableMobileWalletSelection: true,
      }),
      themeMode: 'dark',
      themeVariables: {
        '--w3m-z-index': '10000',
      },
    });
    initialized = true;
    return { ok: true };
  } catch (e) {
    initError = e instanceof Error ? e.message : 'init AppKit failed';
    return { ok: false, error: initError };
  }
}

export const SOLANA_WALLET_OPTIONS = [
  {
    id: 'phantom' as const,
    label: 'Phantom',
    icon: 'https://avatars.githubusercontent.com/u/78782331?s=200&v=4',
  },
  {
    id: 'trust' as const,
    label: 'Trust',
    hint: useMainnet ? undefined : 'devnet?',
    icon: 'https://avatars.githubusercontent.com/u/37784833?s=200&v=4',
  },
  {
    id: 'metamask' as const,
    label: 'MetaMask',
    hint: 'Solana',
    icon: 'https://avatars.githubusercontent.com/u/11744586?s=200&v=4',
  },
  {
    id: 'walletConnect' as const,
    label: 'QR / Другие',
    icon: 'https://avatars.githubusercontent.com/u/179229932?s=200&v=4',
  },
] as const;

export type SolanaWalletId = (typeof SOLANA_WALLET_OPTIONS)[number]['id'];
