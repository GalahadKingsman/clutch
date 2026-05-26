const API_BASE = import.meta.env.VITE_API_URL || '/api/v1';

export type User = {
  id: string;
  telegram_id: number;
  telegram_username?: string;
  first_name: string;
  last_name?: string;
  photo_url?: string;
  wallet_address?: string;
  wallet_linked: boolean;
};

export type DuelParticipant = {
  id: string;
  first_name: string;
  telegram_username?: string;
  photo_url?: string;
};

export type DuelCard = {
  id: string;
  condition_text: string;
  side_creator: string;
  side_opponent: string;
  stake_usd_each: number;
  bank_usd: number;
  status: string;
  deadline_at: string;
  claimed_by?: string;
  appeal_window_ends_at?: string;
  creator_tx?: string;
  opponent_tx?: string;
  on_chain_duel_id?: string;
  creator: DuelParticipant;
  opponent?: DuelParticipant;
};

export type Proof = {
  id: string;
  duel_id: string;
  user_id: string;
  proof_type: string;
  url?: string;
  caption?: string;
  created_at: string;
};

export type AIVerdict = {
  id: string;
  duel_id: string;
  winner_id: string;
  reasoning: string;
  confidence: number;
  evidence_refs: string[];
  verdict_hash: string;
  appeal_window_ends_at?: string;
  can_appeal: boolean;
  is_winner?: boolean;
};

export type ClarifyResponse = {
  normalized_condition: string;
  win_criterion: string;
  tips: string;
};

export type FriendCard = {
  user: User;
  contact_alias?: string;
};

export type FeedResponse = {
  incoming_challenges: DuelCard[];
  active_duels: DuelCard[];
  activity: { id: string; text: string; created_at: string }[];
};

export type ChatMessage = {
  id: string;
  duel_id: string;
  user_id?: string;
  body: string;
  is_system: boolean;
  created_at: string;
};

function getToken(): string | null {
  return localStorage.getItem('clutch_token');
}

export function setToken(token: string) {
  localStorage.setItem('clutch_token', token);
}

const API_TIMEOUT_MS = 20_000;

export async function api<T>(path: string, options: RequestInit = {}): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options.headers as Record<string, string>),
  };
  const token = getToken();
  if (token) headers.Authorization = `Bearer ${token}`;

  const ctrl = new AbortController();
  const timer = window.setTimeout(() => ctrl.abort(), API_TIMEOUT_MS);
  let res: Response;
  try {
    res = await fetch(`${API_BASE}${path}`, {
      ...options,
      headers,
      signal: ctrl.signal,
    });
  } catch (e) {
    if (e instanceof Error && e.name === 'AbortError') {
      throw new Error('Сервер не отвечает. Проверьте API и HTTPS.');
    }
    throw e;
  } finally {
    window.clearTimeout(timer);
  }
  const body = await res.json().catch(() => ({}));
  if (!res.ok) {
    const err = body as { error?: string; message?: string };
    throw new Error(err.message || err.error || res.statusText);
  }
  return body as T;
}

export function authTelegram(initData: string) {
  return api<{ token: { access_token: string }; user: User }>('/auth/telegram', {
    method: 'POST',
    body: JSON.stringify({ init_data: initData }),
  });
}

export function fetchMe() {
  return api<User>('/auth/me');
}

export function walletNonce() {
  return api<{ nonce: string; message: string }>('/auth/wallet/nonce', { method: 'POST' });
}

export function walletLink(payload: {
  wallet_address: string;
  signature: string;
  nonce: string;
}) {
  return api<User>('/auth/wallet/link', {
    method: 'POST',
    body: JSON.stringify(payload),
  });
}

export function fetchFeed() {
  return api<FeedResponse>('/feed/friends');
}

export function fetchFriends() {
  return api<FriendCard[]>('/friends');
}

export function searchUsers(q: string) {
  return api<User[]>(`/users/search?q=${encodeURIComponent(q)}`);
}

export function createInvite() {
  return api<{ code: string; link: string }>('/friends/invite', { method: 'POST' });
}

export function acceptInvite(code: string) {
  return api<{ status: string }>('/friends/accept', {
    method: 'POST',
    body: JSON.stringify({ code }),
  });
}

export function createDuel(body: {
  opponent_id: string;
  condition_text: string;
  side_creator: string;
  side_opponent: string;
  stake_usd_each: number;
  deadline_hours?: number;
}) {
  return api<DuelCard>('/duels', { method: 'POST', body: JSON.stringify(body) });
}

export function fetchDuel(id: string) {
  return api<DuelCard>(`/duels/${id}`);
}

export function acceptDuel(id: string) {
  return api<DuelCard>(`/duels/${id}/accept`, { method: 'POST', body: '{}' });
}

export function cancelDuel(id: string) {
  return api<DuelCard>(`/duels/${id}/cancel`, { method: 'POST', body: '{}' });
}

export function fetchMessages(duelId: string) {
  return api<ChatMessage[]>(`/duels/${duelId}/messages`);
}

export function claimDuel(id: string) {
  return api<{ status: string }>(`/duels/${id}/claim`, { method: 'POST', body: '{}' });
}

export function confirmDuel(id: string) {
  return api<DuelCard>(`/duels/${id}/confirm`, { method: 'POST', body: '{}' });
}

export function fetchPrices() {
  return api<Record<string, number>>('/prices');
}

export function postMessage(duelId: string, body: string) {
  return api<ChatMessage>(`/duels/${duelId}/messages`, {
    method: 'POST',
    body: JSON.stringify({ body }),
  });
}

export type WalletBalances = {
  wallet: string;
  sol: number;
  usdc: number;
  mint: string;
  network: string;
};

export function fetchWalletBalances() {
  return api<WalletBalances>('/wallet/balances');
}

export function fetchDuelTxCreate(duelId: string) {
  return api<{ transaction: string; on_chain_duel_id: string }>(
    `/duels/${duelId}/tx/create`,
  );
}

export function confirmDuelTxCreate(duelId: string, signature: string) {
  return api<DuelCard>(`/duels/${duelId}/tx/create`, {
    method: 'POST',
    body: JSON.stringify({ signature }),
  });
}

export function fetchDuelTxAccept(duelId: string) {
  return api<{ transaction: string }>(`/duels/${duelId}/tx/accept`);
}

export function confirmDuelTxAccept(duelId: string, signature: string) {
  return api<DuelCard>(`/duels/${duelId}/tx/accept`, {
    method: 'POST',
    body: JSON.stringify({ signature }),
  });
}

export function clarifyCondition(body: {
  condition_text: string;
  side_creator: string;
  side_opponent: string;
}) {
  return api<ClarifyResponse>('/ai/clarify-condition', {
    method: 'POST',
    body: JSON.stringify(body),
  });
}

export function openDispute(duelId: string) {
  return api<DuelCard>(`/duels/${duelId}/dispute`, { method: 'POST', body: '{}' });
}

export function listProofs(duelId: string) {
  return api<Proof[]>(`/duels/${duelId}/proofs`);
}

export async function uploadProof(duelId: string, file: File, caption?: string) {
  const form = new FormData();
  form.append('file', file);
  if (caption) form.append('caption', caption);
  const headers: Record<string, string> = {};
  const token = getToken();
  if (token) headers.Authorization = `Bearer ${token}`;
  const res = await fetch(`${API_BASE}/duels/${duelId}/proofs`, {
    method: 'POST',
    headers,
    body: form,
  });
  const body = await res.json().catch(() => ({}));
  if (!res.ok) {
    const err = body as { error?: string; message?: string };
    throw new Error(err.message || err.error || res.statusText);
  }
  return body as Proof;
}

export function runJudge(duelId: string) {
  return api<AIVerdict>(`/duels/${duelId}/judge`, { method: 'POST', body: '{}' });
}

export function fetchVerdict(duelId: string) {
  return api<AIVerdict>(`/duels/${duelId}/verdict`);
}

export function finalizeVerdict(duelId: string) {
  return api<DuelCard>(`/duels/${duelId}/verdict/finalize`, {
    method: 'POST',
    body: '{}',
  });
}

export function appealVerdict(duelId: string) {
  return api<{ duel: DuelCard; appeal: { id: string; status: string } }>(
    `/duels/${duelId}/appeal`,
    { method: 'POST', body: '{}' },
  );
}
