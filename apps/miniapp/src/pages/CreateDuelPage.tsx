import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  clarifyCondition,
  confirmDuelTxCreate,
  createDuel,
  fetchDuelTxCreate,
  fetchFriends,
  type FriendCard,
} from '../lib/api';
import { sendBase64Transaction } from '../lib/solana';

export function CreateDuelPage() {
  const nav = useNavigate();
  const [friends, setFriends] = useState<FriendCard[]>([]);
  const [opponentId, setOpponentId] = useState('');
  const [condition, setCondition] = useState('');
  const [sideCreator, setSideCreator] = useState('');
  const [sideOpponent, setSideOpponent] = useState('');
  const [stake, setStake] = useState('10');
  const [loading, setLoading] = useState(false);
  const [clarifyTips, setClarifyTips] = useState<string | null>(null);
  const [err, setErr] = useState<string | null>(null);

  async function clarify() {
    if (!condition.trim()) return;
    setErr(null);
    try {
      const out = await clarifyCondition({
        condition_text: condition,
        side_creator: sideCreator || 'Я',
        side_opponent: sideOpponent || 'Соперник',
      });
      setCondition(out.normalized_condition);
      setClarifyTips(`${out.win_criterion}\n${out.tips}`);
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'Ошибка AI');
    }
  }

  useEffect(() => {
    void fetchFriends().then(setFriends);
  }, []);

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setErr(null);
    try {
      const duel = await createDuel({
        opponent_id: opponentId,
        condition_text: condition,
        side_creator: sideCreator,
        side_opponent: sideOpponent,
        stake_usd_each: parseFloat(stake) || 10,
        deadline_hours: 24,
      });
      try {
        const { transaction } = await fetchDuelTxCreate(duel.id);
        const sig = await sendBase64Transaction(transaction);
        await confirmDuelTxCreate(duel.id, sig);
      } catch (chainErr) {
        console.warn('on-chain create skipped', chainErr);
      }
      nav(`/duel/${duel.id}`);
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'Ошибка');
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="px-4 pt-6">
      <h1 className="font-display text-xl font-bold">Новая дуэль</h1>
      <form onSubmit={submit} className="mt-4 space-y-3">
        <label className="block text-xs font-extrabold uppercase text-mut">
          Соперник
          <select
            required
            value={opponentId}
            onChange={(e) => setOpponentId(e.target.value)}
            className="mt-1 w-full rounded-xl border border-white/10 bg-panel2 px-3 py-3 text-sm font-semibold"
          >
            <option value="">Выбери друга</option>
            {friends.map((f) => (
              <option key={f.user.id} value={f.user.id}>
                {f.contact_alias || f.user.first_name}
              </option>
            ))}
          </select>
        </label>

        <label className="block text-xs font-extrabold uppercase text-mut">
          Условие
          <textarea
            required
            value={condition}
            onChange={(e) => setCondition(e.target.value)}
            className="mt-1 w-full rounded-xl border border-white/10 bg-panel2 px-3 py-3 text-sm font-semibold"
            rows={2}
            placeholder="Кто победит в матче…"
          />
        </label>
        <button
          type="button"
          onClick={() => void clarify()}
          className="w-full rounded-xl border border-blue/30 py-2 text-sm font-bold text-blue"
        >
          ✨ Уточнить условие (AI)
        </button>
        {clarifyTips && (
          <p className="whitespace-pre-wrap rounded-xl bg-panel2/80 p-3 text-xs text-mut">
            {clarifyTips}
          </p>
        )}

        <div className="grid grid-cols-2 gap-2">
          <label className="block text-xs font-extrabold uppercase text-mut">
            Твоя сторона
            <input
              required
              value={sideCreator}
              onChange={(e) => setSideCreator(e.target.value)}
              className="mt-1 w-full rounded-xl border border-white/10 bg-panel2 px-3 py-3 text-sm"
            />
          </label>
          <label className="block text-xs font-extrabold uppercase text-mut">
            Соперник
            <input
              required
              value={sideOpponent}
              onChange={(e) => setSideOpponent(e.target.value)}
              className="mt-1 w-full rounded-xl border border-white/10 bg-panel2 px-3 py-3 text-sm"
            />
          </label>
        </div>

        <label className="block text-xs font-extrabold uppercase text-mut">
          Ставка USD (с каждой стороны)
          <input
            type="number"
            min="1"
            step="1"
            value={stake}
            onChange={(e) => setStake(e.target.value)}
            className="mt-1 w-full rounded-xl border border-white/10 bg-panel2 px-3 py-3 text-sm font-semibold"
          />
        </label>

        {err && <p className="text-sm text-red">{err}</p>}

        <button
          type="submit"
          disabled={loading}
          className="w-full rounded-2xl bg-gradient-to-b from-[#5C88FF] to-[#4068E8] py-4 font-extrabold text-white shadow-[0_5px_0_#2E51C4] disabled:opacity-60"
        >
          {loading ? 'Создаём…' : 'Бросить вызов'}
        </button>
        <p className="text-center text-xs text-mut">
          После создания Phantom подпишет tx на devnet (регистрация дуэли on-chain).
        </p>
      </form>
    </div>
  );
}
