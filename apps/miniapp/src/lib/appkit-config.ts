import { createAppKit } from '@reown/appkit/react';
import { SolanaAdapter } from '@reown/appkit-adapter-solana/react';
import { solanaDevnet } from '@reown/appkit/networks';

const projectId = import.meta.env.VITE_WALLETCONNECT_PROJECT_ID || '';
const appUrl = import.meta.env.VITE_APP_URL || 'https://clutch-duel.ru';

export const walletConnectConfigured = projectId.length > 0;

if (walletConnectConfigured) {
  const solanaAdapter = new SolanaAdapter();
  createAppKit({
    adapters: [solanaAdapter],
    networks: [solanaDevnet],
    projectId,
    metadata: {
      name: 'CLUTCH',
      description: '1v1 crypto duels on Solana',
      url: appUrl,
      icons: [`${appUrl}/favicon.svg`],
    },
    features: {
      analytics: false,
    },
    themeMode: 'dark',
    themeVariables: {
      '--w3m-z-index': '10000',
    },
  });
}

/** Solana wallets for direct connect (WalletConnect v2 under the hood). */
export const SOLANA_WALLET_OPTIONS = [
  {
    id: 'phantom' as const,
    label: 'Phantom',
    icon: 'https://avatars.githubusercontent.com/u/78782331?s=200&v=4',
  },
  {
    id: 'trust' as const,
    label: 'Trust',
    icon: 'https://avatars.githubusercontent.com/u/37784833?s=200&v=4',
  },
  {
    id: 'metamask' as const,
    label: 'MetaMask',
    icon: 'https://avatars.githubusercontent.com/u/11744586?s=200&v=4',
  },
  {
    id: 'walletConnect' as const,
    label: 'QR / Другие',
    icon: 'https://avatars.githubusercontent.com/u/179229932?s=200&v=4',
  },
];

export type SolanaWalletId = (typeof SOLANA_WALLET_OPTIONS)[number]['id'];
