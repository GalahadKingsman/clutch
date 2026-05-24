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

export type AuthResponse = {
  token: { access_token: string; expires_in: number; token_type: string };
  user: User;
};

function getToken(): string | null {
  return localStorage.getItem('clutch_token');
}

export function setToken(token: string) {
  localStorage.setItem('clutch_token', token);
}

export async function api<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options.headers as Record<string, string>),
  };
  const token = getToken();
  if (token) headers.Authorization = `Bearer ${token}`;

  const res = await fetch(`${API_BASE}${path}`, { ...options, headers });
  const body = await res.json().catch(() => ({}));
  if (!res.ok) {
    const err = body as { error?: string; message?: string };
    throw new Error(err.message || err.error || res.statusText);
  }
  return body as T;
}

export function authTelegram(initData: string) {
  return api<AuthResponse>('/auth/telegram', {
    method: 'POST',
    body: JSON.stringify({ init_data: initData }),
  });
}

export function fetchMe() {
  return api<User>('/auth/me');
}

export function walletNonce() {
  return api<{ nonce: string; message: string }>('/auth/wallet/nonce', {
    method: 'POST',
  });
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
