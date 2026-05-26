import { useCallback, useRef } from 'react';
import {
  useAppKit,
  useAppKitAccount,
  useAppKitProvider,
} from '@reown/appkit/react';
import { useAppKitWallet } from '@reown/appkit-wallet-button/react';
import {
  useAppKitConnection,
  type Provider,
} from '@reown/appkit-adapter-solana/react';
import { Transaction } from '@solana/web3.js';
import type { SolanaWalletId } from './appkit-init';

function signatureToBase64(sig: Uint8Array): string {
  return btoa(String.fromCharCode(...sig));
}

export function useClutchWallet() {
  const { open } = useAppKit();
  const { address, isConnected } = useAppKitAccount();
  const { walletProvider } = useAppKitProvider<Provider>('solana');
  const { connection } = useAppKitConnection();

  const signAuthMessage = useCallback(
    async (message: string): Promise<string> => {
      if (!walletProvider) {
        throw new Error('Кошелёк не подключён');
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
        throw new Error('Кошелёк не подключён');
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
    openWalletModal: () => open({ view: 'Connect', namespace: 'solana' }),
    signAuthMessage,
    sendBase64Transaction,
  };
}

type WalletConnectCallbacks = {
  onSuccess: () => void;
  onError: (message: string) => void;
};

/** Прямое подключение к кошельку (как gmgn) — без общей модалки. */
export function useSolanaWalletConnect(callbacks: WalletConnectCallbacks) {
  const callbacksRef = useRef(callbacks);
  callbacksRef.current = callbacks;

  const { connect, isReady, isPending } = useAppKitWallet({
    namespace: 'solana',
    onSuccess: () => {
      callbacksRef.current.onSuccess();
    },
    onError: (err: Error) => {
      callbacksRef.current.onError(err?.message || 'Ошибка подключения');
    },
  });

  const connectWallet = useCallback(
    (walletId: SolanaWalletId) => {
      if (!isReady) {
        callbacksRef.current.onError(
          'WalletConnect загружается. Подожди 2–3 сек и нажми снова.',
        );
        return;
      }
      connect(walletId);
    },
    [connect, isReady],
  );

  return { connectWallet, isReady, isPending };
}
