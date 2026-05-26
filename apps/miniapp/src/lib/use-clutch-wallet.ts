import { useCallback } from 'react';
import {
  useAppKit,
  useAppKitAccount,
  useAppKitProvider,
} from '@reown/appkit/react';
import {
  useAppKitConnection,
  type Provider,
} from '@reown/appkit-adapter-solana/react';
import { Transaction } from '@solana/web3.js';

function signatureToBase64(sig: Uint8Array): string {
  return btoa(String.fromCharCode(...sig));
}

export function useClutchWallet() {
  const { open } = useAppKit();
  const { address, isConnected } = useAppKitAccount();
  const { walletProvider } = useAppKitProvider<Provider>('solana');
  const { connection } = useAppKitConnection();

  const openWalletModal = useCallback(() => {
    open({ view: 'Connect' });
  }, [open]);

  const signAuthMessage = useCallback(
    async (message: string): Promise<string> => {
      if (!walletProvider) {
        throw new Error('Сначала подключи кошелёк');
      }
      const encoded = new TextEncoder().encode(message);
      const sig = await walletProvider.signMessage(encoded);
      return signatureToBase64(sig);
    },
    [walletProvider],
  );

  const sendBase64Transaction = useCallback(
    async (txBase64: string): Promise<string> => {
      if (!walletProvider) {
        throw new Error('Сначала подключи кошелёк');
      }
      const raw = Uint8Array.from(atob(txBase64), (c) => c.charCodeAt(0));
      const tx = Transaction.from(raw);
      if (!tx.recentBlockhash) {
        const { blockhash } = await connection.getLatestBlockhash();
        tx.recentBlockhash = blockhash;
      }
      if (!tx.feePayer && walletProvider.publicKey) {
        tx.feePayer = walletProvider.publicKey;
      }
      return walletProvider.signAndSendTransaction(tx);
    },
    [connection, walletProvider],
  );

  return {
    address,
    isConnected,
    walletProvider,
    openWalletModal,
    signAuthMessage,
    sendBase64Transaction,
  };
}
