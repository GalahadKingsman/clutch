import { useEffect, useRef, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import {
  fetchDuel,
  listProofs,
  runJudge,
  uploadProof,
  type DuelCard,
  type Proof,
} from '../lib/api';

export function ArbitrationPage() {
  const { id } = useParams<{ id: string }>();
  const nav = useNavigate();
  const fileRef = useRef<HTMLInputElement>(null);
  const [duel, setDuel] = useState<DuelCard | null>(null);
  const [proofs, setProofs] = useState<Proof[]>([]);
  const [caption, setCaption] = useState('');
  const [loading, setLoading] = useState(false);
  const [err, setErr] = useState<string | null>(null);

  async function load() {
    if (!id) return;
    const [d, p] = await Promise.all([fetchDuel(id), listProofs(id)]);
    setDuel(d);
    setProofs(p);
  }

  useEffect(() => {
    void load().catch((e) =>
      setErr(e instanceof Error ? e.message : 'Ошибка'),
    );
  }, [id]);

  async function onFile(ev: React.ChangeEvent<HTMLInputElement>) {
    const file = ev.target.files?.[0];
    if (!id || !file) return;
    setLoading(true);
    setErr(null);
    try {
      await uploadProof(id, file, caption || undefined);
      setCaption('');
      if (fileRef.current) fileRef.current.value = '';
      await load();
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'Ошибка загрузки');
    } finally {
      setLoading(false);
    }
  }

  async function judge() {
    if (!id) return;
    setLoading(true);
    setErr(null);
    try {
      await runJudge(id);
      nav(`/duel/${id}/verdict`);
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'Ошибка судьи');
    } finally {
      setLoading(false);
    }
  }

  if (!duel && !err) {
    return (
      <div className="flex min-h-screen items-center justify-center text-mut">
        Загрузка…
      </div>
    );
  }

  return (
    <div className="flex min-h-screen flex-col px-4 pt-4 pb-8">
      <button
        type="button"
        onClick={() => nav(`/duel/${id}`)}
        className="text-sm font-bold text-blue"
      >
        ← Комната дуэли
      </button>

      <h1 className="mt-2 font-display text-lg font-bold">Арбитраж</h1>
      <p className="text-xs text-mut">{duel?.condition_text}</p>
      <p className="mt-1 text-xs text-gold">
        Загрузи скрин/фото как доказательство. Минимум 1 файл с каждой стороны
        желательно; судья запустится после загрузки.
      </p>

      <div className="mt-4 grid grid-cols-2 gap-2">
        {proofs.map((p) => (
          <a
            key={p.id}
            href={p.url}
            target="_blank"
            rel="noreferrer"
            className="block overflow-hidden rounded-xl border border-white/10 bg-panel2"
          >
            <img src={p.url} alt="" className="h-28 w-full object-cover" />
            {p.caption && (
              <p className="truncate px-2 py-1 text-xs text-mut">{p.caption}</p>
            )}
          </a>
        ))}
      </div>

      <div className="mt-4 space-y-2">
        <input
          value={caption}
          onChange={(e) => setCaption(e.target.value)}
          placeholder="Подпись к фото (опционально)"
          className="w-full rounded-xl border border-white/10 bg-panel2 px-3 py-2 text-sm"
        />
        <input
          ref={fileRef}
          type="file"
          accept="image/*"
          className="hidden"
          onChange={(e) => void onFile(e)}
        />
        <button
          type="button"
          disabled={loading}
          onClick={() => fileRef.current?.click()}
          className="w-full rounded-xl border border-white/10 py-3 text-sm font-bold"
        >
          {loading ? 'Загрузка…' : '+ Добавить пруф'}
        </button>
        <button
          type="button"
          disabled={loading || proofs.length === 0}
          onClick={() => void judge()}
          className="w-full rounded-2xl bg-gold py-4 font-extrabold text-[#3a2600] disabled:opacity-50"
        >
          ⚖️ Запустить ИИ-судью
        </button>
      </div>

      {err && <p className="mt-3 text-sm text-red">{err}</p>}
    </div>
  );
}
