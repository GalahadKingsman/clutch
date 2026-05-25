import { useEffect, useState } from 'react';
import {
  acceptInvite,
  createInvite,
  fetchFriends,
  searchUsers,
  type FriendCard,
  type User,
} from '../lib/api';

export function FriendsPage() {
  const [friends, setFriends] = useState<FriendCard[]>([]);
  const [q, setQ] = useState('');
  const [results, setResults] = useState<User[]>([]);
  const [inviteLink, setInviteLink] = useState<string | null>(null);

  async function load() {
    setFriends(await fetchFriends());
  }

  useEffect(() => {
    void load();
  }, []);

  useEffect(() => {
    if (q.length < 2) {
      setResults([]);
      return;
    }
    const t = setTimeout(() => {
      void searchUsers(q).then(setResults).catch(() => setResults([]));
    }, 300);
    return () => clearTimeout(t);
  }, [q]);

  async function invite() {
    const res = await createInvite();
    setInviteLink(res.link);
    const tg = window.Telegram?.WebApp;
    if (tg?.openTelegramLink) {
      tg.openTelegramLink(`https://t.me/share/url?url=${encodeURIComponent(res.link)}&text=${encodeURIComponent('Дуэли в CLUTCH ⚔️')}`);
    }
  }

  return (
    <div className="px-4 pt-6">
      <h1 className="font-display text-xl font-bold">Друзья</h1>

      <button
        type="button"
        onClick={() => void invite()}
        className="mt-4 w-full rounded-2xl bg-gradient-to-b from-[#FFD45C] to-[#F2B01E] py-3 text-sm font-extrabold text-[#3a2600] shadow-[0_4px_0_#C9930F]"
      >
        Пригласить в CLUTCH
      </button>
      {inviteLink && (
        <p className="mt-2 break-all text-xs text-mut">{inviteLink}</p>
      )}

      <input
        className="mt-4 w-full rounded-xl border border-white/10 bg-panel2 px-4 py-3 text-sm font-semibold outline-none focus:border-blue"
        placeholder="Поиск по @username или имени"
        value={q}
        onChange={(e) => setQ(e.target.value)}
      />

      {results.length > 0 && (
        <div className="mt-2 space-y-2">
          {results.map((u) => (
            <div
              key={u.id}
              className="flex items-center justify-between rounded-xl border border-white/10 bg-panel px-3 py-2"
            >
              <span className="font-bold">
                {u.first_name}{' '}
                {u.telegram_username ? `@${u.telegram_username}` : ''}
              </span>
            </div>
          ))}
          <p className="text-xs text-mut">
            Добавление в друзья — только через инвайт-ссылку (Phase 1)
          </p>
        </div>
      )}

      <div className="mt-6 space-y-2">
        {friends.map((f) => (
          <div
            key={f.user.id}
            className="flex items-center gap-3 rounded-xl border border-white/10 bg-panel px-3 py-3"
          >
            <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-blue/20 text-lg font-bold">
              {f.user.first_name[0]}
            </div>
            <div>
              <p className="font-bold">
                {f.contact_alias || f.user.first_name}
              </p>
              {f.user.telegram_username && (
                <p className="text-xs text-mut">@{f.user.telegram_username}</p>
              )}
            </div>
          </div>
        ))}
      </div>

    </div>
  );
}

export async function tryAcceptInviteFromStartParam() {
  const param =
    window.Telegram?.WebApp?.initDataUnsafe?.start_param ||
    new URLSearchParams(window.location.search).get('startapp') ||
    '';
  if (!param || !param.includes('invite')) return;
  try {
    await acceptInvite(param);
  } catch {
    /* already friends or invalid */
  }
}
