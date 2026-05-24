import type { User } from '../lib/api';

type Props = { user: User };

export function FeedShell({ user }: Props) {
  return (
    <div className="flex min-h-screen flex-col px-4 pb-24 pt-6">
      <div className="flex items-center justify-between">
        <div>
          <p className="text-xs font-extrabold uppercase tracking-wide text-mut">
            Привет, {user.first_name}
          </p>
          <h1 className="font-display text-xl font-bold">Лента</h1>
        </div>
        <div className="rounded-full border border-white/10 bg-panel2 px-3 py-1 text-sm font-bold">
          💰 $0
        </div>
      </div>

      <div className="mt-10 flex flex-1 flex-col items-center justify-center text-center">
        <p className="text-lg font-bold">Пока тихо</p>
        <p className="mt-2 max-w-xs text-sm font-semibold text-mut">
          Phase 1: дуэли, друзья и ставки. Кошелёк привязан ✓
        </p>
        {user.wallet_address && (
          <p className="mt-4 break-all text-xs text-mut">
            {user.wallet_address.slice(0, 4)}…{user.wallet_address.slice(-4)}
          </p>
        )}
      </div>
    </div>
  );
}
