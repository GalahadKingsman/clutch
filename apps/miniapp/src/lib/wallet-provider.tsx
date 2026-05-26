import { type ReactNode, useMemo } from 'react';
import { ConnectionProvider, WalletProvider } from '@solana/wallet-adapter-react';
import { WalletAdapterNetwork } from '@solana/wallet-adapter-base';
import { WalletConnectWalletAdapter } from '@solana/wallet-adapter-walletconnect';
import { clusterApiUrl } from '@solana/web3.js';

const network =
  import.meta.env.VITE_SOLANA_NETWORK === 'mainnet-beta'
    ? WalletAdapterNetwork.Mainnet
    : WalletAdapterNetwork.Devnet;

const endpoint =
  import.meta.env.VITE_SOLANA_RPC_URL || clusterApiUrl(network);

const appUrl = import.meta.env.VITE_APP_URL || 'https://clutch-duel.ru';
const projectId = import.meta.env.VITE_WALLETCONNECT_PROJECT_ID || '';

export function SolanaWalletProvider({ children }: { children: ReactNode }) {
  const wallets = useMemo(
    () => [
      new WalletConnectWalletAdapter({
        network,
        options: {
          projectId,
          metadata: {
            name: 'CLUTCH',
            description: '1v1 crypto duels on Solana',
            url: appUrl,
            icons: [`${appUrl}/favicon.svg`],
          },
        },
      }),
    ],
    [],
  );

  return (
    <ConnectionProvider endpoint={endpoint}>
      <WalletProvider wallets={wallets} autoConnect>
        {children}
      </WalletProvider>
    </ConnectionProvider>
  );
}

export function walletConnectConfigured(): boolean {
  return projectId.length > 0;
}
