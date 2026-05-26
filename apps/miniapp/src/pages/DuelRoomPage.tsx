import { useEffect, useRef, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import {
  cancelDuel,
  claimDuel,
  confirmDuel,
  confirmDuelTxAccept,
  fetchDuel,
  fetchDuelTxAccept,
  fetchMessages,
  openDispute,
  type ChatMessage,
  type DuelCard,
  type User,
} from '../lib/api';
import { useClutchWallet } from '../lib/use-clutch-wallet';
import { WalletConnectBanner } from '../components/WalletConnectBanner';

const API_BASE = import.meta.env.VITE_API_URL || '/api/v1';

function getToken() {
  return localStorage.getItem('clutch_token');
}

type Props = { user: User };

export function DuelRoomPage({ user }: Props) {
  const { id } = useParams<{ id: string }>();
  const nav = useNavigate();
  const { sendBase64Transaction } = useClutchWallet();
  const [duel, setDuel] = useState<DuelCard | null>(null);
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [text, setText] = useState('');
  const [err, setErr] = useState<string | null>(null);
  const bottomRef = useRef<HTMLDivElement>(null);
  const wsRef = useRef<WebSocket | null>(null);

  async function load() {
    if (!id) return;
    const [d, msgs] = await Promise.all([fetchDuel(id), fetchMessages(id)]);
    setDuel(d);
    setMessages(msgs);
  }

  useEffect(() => {
    void load().catch((e) =>
      setErr(e instanceof Error ? e.message : 'Ошибка'),
    );
  }, [id]);

  useEffect(() => {
    if (!id) return;
    const token = getToken();
    if (!token) return;
    const proto = location.protocol === 'https:' ? 'wss:' : 'ws:';
    const ws = new WebSocket(
      `${proto}//${location.host}${API_BASE}/duels/${id}/ws?token=${encodeURIComponent(token)}`,
    );
    wsRef.current = ws;
    ws.onmessage = (ev) => {
      try {
        const msg = JSON.parse(ev.data) as ChatMessage;
        setMessages((prev) => [...prev, msg]);
      } catch {
        /* ignore */
      }
    };
    return () => ws.close();
  }, [id]);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  useEffect(() => {
    if (!duel || !id) return;
    if (
      ['disputed', 'arbitration_upload', 'ai_judging'].includes(duel.status)
    ) {
      nav(`/duel/${id}/arbitration`, { replace: true });
    } else if (
      ['appeal_window', 'human_arbitration'].includes(duel.status)
    ) {
      nav(`/duel/${id}/verdict`, { replace: true });
    }
  }, [duel?.status, id, nav]);

  async function send() {
    if (!id || !text.trim() || !wsRef.current) return;
    wsRef.current.send(JSON.stringify({ body: text.trim() }));
    setText('');
  }

  async function acceptOnChain() {
    if (!id) return;
    setErr(null);
    try {
      const { transaction } = await fetchDuelTxAccept(id);
      const sig = await sendBase64Transaction(transaction);
      const updated = await confirmDuelTxAccept(id, sig);
      setDuel(updated);
      const msgs = await fetchMessages(id);
      setMessages(msgs);
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'Ошибка accept');
    }
  }

  if (!duel && !err) {
    return (
      <div className="flex min-h-screen items-center justify-center text-mut">
        Загрузка…
      </div>
    );
  }

  if (err || !duel) {
    return (
      <div className="px-4 pt-10 text-center text-red">{err || 'Не найдено'}</div>
    );
  }

  const isPending = duel.status === 'pending_opponent';
  const isActive = duel.status === 'active';
  const isAwaitingClaim = duel.status === 'awaiting_claim';
  const isClaimer = duel.claimed_by === user.id;
  const canChat =
    isActive || isAwaitingClaim || duel.status === 'disputed';

  return (
    <div className="flex min-h-screen flex-col px-4 pt-4 pb-6">
      <button
        type="button"
        onClick={() => nav(-1)}
        className="text-sm font-bold text-blue"
      >
        ← Назад
      </button>

      <h1 className="mt-2 font-display text-lg font-bold">{duel.condition_text}</h1>
      <p className="text-xs text-mut">
        Банк ${duel.bank_usd} · {duel.status}
      </p>

      <WalletConnectBanner className="mt-3" />

      <div className="mt-3 flex flex-wrap gap-2">
        {isPending && (
          <>
            <button
              type="button"
              onClick={() => void acceptOnChain()}
              className="rounded-xl bg-blue px-4 py-2 text-sm font-extrabold"
            >
              Принять (devnet)
            </button>
            <button
              type="button"
              onClick={() => void cancelDuel(duel.id).then(() => nav('/feed'))}
              className="rounded-xl border border-white/10 px-4 py-2 text-sm font-bold"
            >
              Отклонить
            </button>
          </>
        )}
        {isActive && (
          <button
            type="button"
            onClick={() => void claimDuel(duel.id).then(load)}
            className="rounded-xl bg-gold px-4 py-2 text-sm font-extrabold text-[#3a2600]"
          >
            Я победил
          </button>
        )}
        {isAwaitingClaim && !isClaimer && (
          <>
            <button
              type="button"
              onClick={() => void confirmDuel(duel.id).then(() => nav('/feed'))}
              className="rounded-xl bg-green px-4 py-2 text-sm font-extrabold text-[#053022]"
            >
              Подтвердить победу соперника
            </button>
            <button
              type="button"
              onClick={() =>
                void openDispute(duel.id).then(() =>
                  nav(`/duel/${duel.id}/arbitration`),
                )
              }
              className="rounded-xl border border-red/40 px-4 py-2 text-sm font-extrabold text-red"
            >
              Оспорить
            </button>
          </>
        )}
        {isAwaitingClaim && isClaimer && (
          <p className="text-sm text-mut">
            Ждём подтверждения или спора от соперника…
          </p>
        )}
        {(duel.status === 'disputed' ||
          duel.status === 'arbitration_upload' ||
          duel.status === 'ai_judging') && (
          <button
            type="button"
            onClick={() => nav(`/duel/${duel.id}/arbitration`)}
            className="rounded-xl bg-gold px-4 py-2 text-sm font-extrabold text-[#3a2600]"
          >
            Арбитраж →
          </button>
        )}
        {duel.status === 'appeal_window' && (
          <button
            type="button"
            onClick={() => nav(`/duel/${duel.id}/verdict`)}
            className="rounded-xl bg-blue px-4 py-2 text-sm font-extrabold"
          >
            Вердикт →
          </button>
        )}
      </div>

      <div className="mt-4 flex-1 space-y-2 overflow-y-auto rounded-2xl border border-white/10 bg-panel/50 p-3 min-h-[280px] max-h-[50vh]">
        {messages.map((m) => (
          <div
            key={m.id}
            className={`text-sm ${m.is_system ? 'font-semibold text-gold' : ''}`}
          >
            {m.body}
          </div>
        ))}
        <div ref={bottomRef} />
      </div>

      {canChat && (
        <div className="mt-3 flex gap-2">
          <input
            value={text}
            onChange={(e) => setText(e.target.value)}
            className="flex-1 rounded-xl border border-white/10 bg-panel2 px-3 py-2 text-sm"
            placeholder="Сообщение…"
            onKeyDown={(e) => e.key === 'Enter' && void send()}
          />
          <button
            type="button"
            onClick={() => void send()}
            className="rounded-xl bg-blue px-4 py-2 text-sm font-extrabold"
          >
            →
          </button>
        </div>
      )}
    </div>
  );
}
