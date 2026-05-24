import { useEffect, useState } from 'react';
import { authTelegram, fetchMe, setToken, type User } from './lib/api';
import { FeedShell } from './components/FeedShell';
import { WalletGate } from './components/WalletGate';

export default function App() {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const tg = window.Telegram?.WebApp;
    tg?.ready();
    tg?.expand();
    document.documentElement.style.setProperty(
      '--tg-theme-bg-color',
      '#0f0e16',
    );

    async function boot() {
      try {
        const initData = tg?.initData;
        if (!initData) {
          setError('Открой приложение через Telegram (@clutch_game_bot)');
          return;
        }
        const res = await authTelegram(initData);
        setToken(res.token.access_token);
        setUser(res.user);
      } catch (e) {
        setError(e instanceof Error ? e.message : 'Ошибка входа');
      } finally {
        setLoading(false);
      }
    }
    void boot();
  }, []);

  async function refreshUser() {
    const me = await fetchMe();
    setUser(me);
  }

  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center text-mut">
        Загрузка…
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex min-h-screen items-center justify-center px-6 text-center text-red">
        {error}
      </div>
    );
  }

  if (!user) return null;

  if (!user.wallet_linked) {
    return <WalletGate onLinked={refreshUser} />;
  }

  return <FeedShell user={user} />;
}
