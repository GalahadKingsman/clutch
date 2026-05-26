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
  });
}
