import { useEffect, useState } from 'react';
import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom';
import { authTelegram, fetchMe, setToken, type User } from './lib/api';
import { WalletGate } from './components/WalletGate';
import { AppLayout } from './components/AppLayout';
import { FeedPage } from './pages/FeedPage';
import { FriendsPage, tryAcceptInviteFromStartParam } from './pages/FriendsPage';
import { CreateDuelPage } from './pages/CreateDuelPage';
import { DuelRoomPage } from './pages/DuelRoomPage';
import { ArbitrationPage } from './pages/ArbitrationPage';
import { VerdictPage } from './pages/VerdictPage';
import { WalletPage } from './pages/WalletPage';
import { ProfilePage } from './pages/ProfilePage';

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
        await tryAcceptInviteFromStartParam();
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

  return (
    <BrowserRouter>
      <Routes>
        <Route element={<AppLayout />}>
          <Route path="/feed" element={<FeedPage user={user} />} />
          <Route path="/friends" element={<FriendsPage />} />
          <Route path="/duel/create" element={<CreateDuelPage />} />
          <Route path="/wallet" element={<WalletPage user={user} />} />
          <Route path="/profile" element={<ProfilePage user={user} />} />
        </Route>
        <Route path="/duel/:id" element={<DuelRoomPage user={user} />} />
        <Route path="/duel/:id/arbitration" element={<ArbitrationPage />} />
        <Route path="/duel/:id/verdict" element={<VerdictPage />} />
        <Route path="/" element={<Navigate to="/feed" replace />} />
        <Route path="*" element={<Navigate to="/feed" replace />} />
      </Routes>
    </BrowserRouter>
  );
}
