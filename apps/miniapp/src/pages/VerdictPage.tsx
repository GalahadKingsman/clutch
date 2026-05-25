import { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import {
  appealVerdict,
  fetchDuel,
  fetchVerdict,
  finalizeVerdict,
  type AIVerdict,
  type DuelCard,
} from '../lib/api';

function msLeft(endsAt?: string) {
  if (!endsAt) return 0;
  return Math.max(0, new Date(endsAt).getTime() - Date.now());
}

export function VerdictPage() {
  const { id } = useParams<{ id: string }>();
  const nav = useNavigate();
  const [duel, setDuel] = useState<DuelCard | null>(null);
  const [verdict, setVerdict] = useState<AIVerdict | null>(null);
  const [leftMs, setLeftMs] = useState(0);
  const [loading, setLoading] = useState(false);
  const [err, setErr] = useState<string | null>(null);

  async function load() {
    if (!id) return;
    const [d, v] = await Promise.all([fetchDuel(id), fetchVerdict(id)]);
    setDuel(d);
    setVerdict(v);
    setLeftMs(msLeft(v.appeal_window_ends_at));
  }

  useEffect(() => {
    void load().catch((e) =>
      setErr(e instanceof Error ? e.message : 'Ошибка'),
    );
  }, [id]);

  useEffect(() => {
    if (!verdict?.appeal_window_ends_at) return;
    const t = setInterval(() => {
      setLeftMs(msLeft(verdict.appeal_window_ends_at));
    }, 1000);
    return () => clearInterval(t);
  }, [verdict?.appeal_window_ends_at]);

  async function appeal() {
    if (!id) return;
    setLoading(true);
    setErr(null);
    try {
      await appealVerdict(id);
      nav(`/duel/${id}`);
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'Ошибка апелляции');
    } finally {
      setLoading(false);
    }
  }

  async function finalize() {
    if (!id) return;
    setLoading(true);
    try {
      await finalizeVerdict(id);
      nav('/feed');
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'Ошибка');
    } finally {
      setLoading(false);
    }
  }

  if (!verdict && !err) {
    return (
      <div className="flex min-h-screen items-center justify-center text-mut">
        Загрузка…
      </div>
    );
  }

  const mins = Math.floor(leftMs / 60000);
  const secs = Math.floor((leftMs % 60000) / 1000);
  const settled =
    duel?.status === 'settled' || duel?.status === 'mutual_settled';

  return (
    <div className="flex min-h-screen flex-col px-4 pt-4 pb-8">
      <button
        type="button"
        onClick={() => nav(`/duel/${id}`)}
        className="text-sm font-bold text-blue"
      >
        ← Назад
      </button>

      <h1 className="mt-2 font-display text-xl font-bold">
        {verdict?.is_winner ? '🏆 Победа' : '⚖️ Вердикт ИИ'}
      </h1>
      <p className="mt-2 text-sm leading-relaxed text-white/90">
        {verdict?.reasoning}
      </p>
      <p className="mt-2 text-xs text-mut">
        Уверенность: {Math.round((verdict?.confidence ?? 0) * 100)}%
      </p>

      {duel?.status === 'appeal_window' && leftMs > 0 && (
        <p className="mt-3 rounded-xl bg-panel2 px-3 py-2 text-sm text-gold">
          Окно апелляции: {mins}:{secs.toString().padStart(2, '0')}
        </p>
      )}

      <div className="mt-6 flex flex-col gap-2">
        {verdict?.can_appeal && (
          <button
            type="button"
            disabled={loading}
            onClick={() => void appeal()}
            className="rounded-2xl border border-gold/40 py-3 text-sm font-extrabold text-gold"
          >
            Апелляция к человеку (5% банка)
          </button>
        )}
        {verdict?.is_winner && duel?.status === 'appeal_window' && (
          <button
            type="button"
            disabled={loading}
            onClick={() => void finalize()}
            className="rounded-2xl bg-green py-3 text-sm font-extrabold text-[#053022]"
          >
            Забрать выигрыш сейчас
          </button>
        )}
        {settled && (
          <button
            type="button"
            onClick={() => nav('/feed')}
            className="rounded-2xl bg-blue py-3 text-sm font-extrabold"
          >
            В ленту
          </button>
        )}
      </div>

      {err && <p className="mt-3 text-sm text-red">{err}</p>}
    </div>
  );
}
