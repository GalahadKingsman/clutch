import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  acceptDuel,
  fetchFeed,
  fetchPrices,
  type DuelCard,
  type FeedResponse,
  type User,
} from '../lib/api';
import { DuelCardView } from '../components/DuelCardView';

type Props = { user: User };

export function FeedPage({ user }: Props) {
  const nav = useNavigate();
  const [feed, setFeed] = useState<FeedResponse | null>(null);
  const [balance, setBalance] = useState('$—');
  const [err, setErr] = useState<string | null>(null);

  async function load() {
    try {
      const [f, prices] = await Promise.all([fetchFeed(), fetchPrices()]);
      setFeed(f);
      const sol = prices.SOL ?? 0;
      setBalance(sol > 0 ? `$${Math.round(sol)} SOL` : '💰 $—');
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'Ошибка');
    }
  }

  useEffect(() => {
    void load();
  }, []);

  async function handleAccept(d: DuelCard) {
    await acceptDuel(d.id);
    nav(`/duel/${d.id}`);
  }

  const empty =
    !feed?.incoming_challenges.length && !feed?.active_duels.length;

  return (
    <div className="px-4 pt-6">
      <div className="flex items-center justify-between">
        <div>
          <p className="text-xs font-extrabold uppercase tracking-wide text-mut">
            Привет, {user.first_name}
          </p>
          <h1 className="font-display text-xl font-bold">Лента</h1>
        </div>
        <div className="rounded-full border border-white/10 bg-panel2 px-3 py-1 text-sm font-bold">
          {balance}
        </div>
      </div>

      {err && (
        <p className="mt-4 text-sm text-red">{err}</p>
      )}

      <div className="mt-4 flex rounded-2xl border border-white/10 bg-panel2 p-1">
        <span className="flex-1 rounded-xl bg-blue py-2 text-center text-sm font-extrabold text-white">
          👥 Друзья
        </span>
        <span className="flex-1 py-2 text-center text-sm font-bold text-mut">
          🌐 Арена
        </span>
      </div>

      <div className="mt-6 space-y-3">
        {feed?.incoming_challenges.map((d) => (
          <DuelCardView
            key={d.id}
            duel={d}
            actionLabel="Принять вызов"
            onAction={() => void handleAccept(d)}
          />
        ))}
        {feed?.active_duels.map((d) => (
          <DuelCardView key={d.id} duel={d} />
        ))}
      </div>

      {empty && !err && (
        <div className="mt-16 text-center">
          <p className="text-lg font-bold">Пока тихо</p>
          <p className="mt-2 text-sm font-semibold text-mut">
            Создай дуэль с другом или прими входящий вызов
          </p>
        </div>
      )}
    </div>
  );
}
