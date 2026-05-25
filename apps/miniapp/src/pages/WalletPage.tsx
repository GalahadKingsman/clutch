import { useEffect, useState } from 'react';
import { fetchWalletBalances, type User, type WalletBalances } from '../lib/api';

type Props = { user: User };

export function WalletPage({ user }: Props) {
  const [bal, setBal] = useState<WalletBalances | null>(null);
  const [err, setErr] = useState<string | null>(null);

  useEffect(() => {
    void fetchWalletBalances()
      .then(setBal)
      .catch((e) => setErr(e instanceof Error ? e.message : 'Ошибка'));
  }, []);

  return (
    <div className="px-4 pt-6">
      <h1 className="font-display text-xl font-bold">Кошелёк</h1>
      <div className="mt-6 rounded-2xl border border-white/10 bg-panel p-4">
        <p className="text-xs font-extrabold uppercase text-mut">Адрес</p>
        <p className="mt-2 break-all text-sm font-semibold">
          {user.wallet_address || '—'}
        </p>
        {bal && (
          <div className="mt-4 space-y-2 text-sm font-bold">
            <p>SOL: {bal.sol.toFixed(4)}</p>
            <p>USDC (devnet): {bal.usdc.toFixed(2)}</p>
            <p className="text-xs font-semibold text-mut">Сеть: {bal.network}</p>
          </div>
        )}
        {err && <p className="mt-4 text-sm text-red">{err}</p>}
        <p className="mt-4 text-xs text-mut">
          USDC — с faucet.circle.com (Solana Devnet). Эскроу USDC в программе — Phase 2.
        </p>
      </div>
    </div>
  );
}
