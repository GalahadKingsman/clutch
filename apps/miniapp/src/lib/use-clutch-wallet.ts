import { useCallback } from 'react';
import { useConnection, useWallet } from '@solana/wallet-adapter-react';
import { type PublicKey, Transaction } from '@solana/web3.js';

function signatureToBase64(sig: Uint8Array): string {
  return btoa(String.fromCharCode(...sig));
}

export function useClutchWallet() {
  const { connection } = useConnection();
  const {
    publicKey,
    connect,
    connected,
    connecting,
    signMessage,
    sendTransaction,
    wallet,
  } = useWallet();

  const ensureConnected = useCallback(async (): Promise<PublicKey> => {
    if (publicKey) return publicKey;
    await connect();
    const pk = wallet?.adapter.publicKey;
    if (!pk) {
      throw new Error('Подключение отменено');
    }
    return pk;
  }, [connect, publicKey, wallet]);

  const signAuthMessage = useCallback(
    async (message: string): Promise<string> => {
      await ensureConnected();
      const sign = signMessage ?? wallet?.adapter.signMessage?.bind(wallet.adapter);
      if (!sign) {
        throw new Error('Кошелёк не поддерживает подпись сообщений');
      }
      const encoded = new TextEncoder().encode(message);
      const sig = await sign(encoded);
      return signatureToBase64(sig);
    },
    [ensureConnected, signMessage, wallet],
  );

  const sendBase64Transaction = useCallback(
    async (txBase64: string): Promise<string> => {
      await ensureConnected();
      if (!sendTransaction) {
        throw new Error('Кошелёк не поддерживает отправку транзакций');
      }
      const raw = Uint8Array.from(atob(txBase64), (c) => c.charCodeAt(0));
      const tx = Transaction.from(raw);
      return sendTransaction(tx, connection, { skipPreflight: false });
    },
    [connection, ensureConnected, sendTransaction],
  );

  return {
    publicKey,
    walletName: wallet?.adapter.name,
    connected,
    connecting,
    connect,
    ensureConnected,
    signAuthMessage,
    sendBase64Transaction,
  };
}
