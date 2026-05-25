import { Link } from 'react-router-dom';
import type { DuelCard } from '../lib/api';

type Props = {
  duel: DuelCard;
  actionLabel?: string;
  onAction?: () => void;
};

export function DuelCardView({ duel, actionLabel, onAction }: Props) {
  const opp = duel.opponent;
  const name = opp?.first_name || 'Соперник';
  return (
    <div className="rounded-2xl border border-white/10 bg-panel p-4">
      <div className="flex items-start justify-between gap-2">
        <div>
          <p className="text-xs font-extrabold uppercase tracking-wide text-mut">
            {duel.status === 'pending_opponent' ? 'Входящий вызов' : 'Дуэль'}
          </p>
          <p className="mt-1 text-sm font-bold leading-snug">
            {duel.condition_text}
          </p>
        </div>
        <div className="rounded-xl bg-gold/15 px-2 py-1 text-xs font-extrabold text-gold">
          ${duel.bank_usd}
        </div>
      </div>
      <p className="mt-2 text-xs font-semibold text-mut">
        {duel.side_creator} vs {duel.side_opponent} · с {name}
      </p>
      <p className="mt-1 text-xs text-mut">
        Ставка ${duel.stake_usd_each} с каждой стороны
      </p>
      {actionLabel && onAction ? (
        <button
          type="button"
          onClick={onAction}
          className="mt-3 w-full rounded-xl bg-gradient-to-b from-[#5C88FF] to-[#4068E8] py-3 text-sm font-extrabold text-white shadow-[0_4px_0_#2E51C4]"
        >
          {actionLabel}
        </button>
      ) : (
        <Link
          to={`/duel/${duel.id}`}
          className="mt-3 block w-full rounded-xl border border-white/10 bg-panel2 py-3 text-center text-sm font-extrabold"
        >
          Открыть комнату
        </Link>
      )}
    </div>
  );
}
