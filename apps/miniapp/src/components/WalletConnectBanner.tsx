import { useClutchWallet } from '../lib/use-clutch-wallet';

type Props = {
  className?: string;
};

/** Напоминание переподключить WC перед on-chain tx (сессия могла сброситься). */
export function WalletConnectBanner({ className = '' }: Props) {
  const { connected, connecting, connect, walletName } = useClutchWallet();

  if (connected) return null;

  return (
    <div
      className={`rounded-xl border border-gold/30 bg-gold/10 px-3 py-2 text-xs ${className}`}
    >
      <p className="font-semibold text-gold">Кошелёк не подключён</p>
      <p className="mt-1 text-mut">
        Для транзакций на devnet снова подключи WalletConnect.
      </p>
      <button
        type="button"
        disabled={connecting}
        onClick={() => void connect()}
        className="mt-2 font-bold text-blue"
      >
        {connecting ? 'Подключение…' : 'Подключить'}
        {walletName ? ` (${walletName})` : ''}
      </button>
    </div>
  );
}
