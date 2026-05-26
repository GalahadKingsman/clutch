import { useAppKitInit } from './AppKitInitProvider';
import { WalletGateConnect } from './WalletGateConnect';
import { ErrorBoundary } from './ErrorBoundary';

type Props = {
  onLinked: () => void;
};

export function WalletGate({ onLinked }: Props) {
  const { ready, configured, error: initError } = useAppKitInit();

  return (
    <div className="flex min-h-screen flex-col items-center justify-center px-6 py-10 text-center">
      <p className="font-display text-2xl font-bold">clutch</p>
      <h1 className="mt-6 text-xl font-bold">Привязать кошелёк</h1>
      <p className="mt-3 max-w-xs text-sm font-semibold text-mut">
        Выбери кошелёк Solana (WalletConnect).
      </p>

      {!ready && (
        <p className="mt-8 text-sm text-gold">Загрузка WalletConnect…</p>
      )}

      {ready && !configured && (
        <p className="mt-6 rounded-xl border border-red/30 bg-red/10 px-4 py-3 text-sm text-red">
          {initError}
          <br />
          <span className="text-xs text-mut">
            Добавь VITE_WALLETCONNECT_PROJECT_ID в .env и пересобери nginx.
          </span>
        </p>
      )}

      {ready && configured && (
        <ErrorBoundary fallbackTitle="Ошибка кошелька">
          <WalletGateConnect onLinked={onLinked} />
        </ErrorBoundary>
      )}
    </div>
  );
}
