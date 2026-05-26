import { useEffect, useState } from 'react';
import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom';
import { authTelegram, fetchMe, setToken, type User } from './lib/api';
import {
  isTelegramWebApp,
  telegramPlatform,
  waitForInitData,
} from './lib/telegram';
import { AppKitInitProvider } from './components/AppKitInitProvider';
import { ErrorBoundary } from './components/ErrorBoundary';
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
import { TgWalletPage } from './pages/TgWalletPage';

function isSecureContext(): boolean {
  if (typeof window === 'undefined') return true;
  return (
    window.isSecureContext ||
    window.location.protocol === 'https:' ||
    window.location.hostname === 'localhost'
  );
}

export default function App() {
  if (typeof window !== 'undefined' && window.location.pathname.startsWith('/tg-wallet')) {
    return <TgWalletPage />;
  }

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
        const inTg = isTelegramWebApp();

        if (inTg && !isSecureContext()) {
          setError(
            'Mini App требует HTTPS. Включите SSL (Cloudflare или Certbot) и укажите https:// в BotFather. Подробнее: docs/HTTPS_TELEGRAM.md',
          );
          return;
        }

        let initData = tg?.initData?.trim() ?? '';
        if (!initData && inTg) {
          try {
            initData = await waitForInitData(6000);
          } catch {
            /* fall through */
          }
        }

        if (!initData) {
          if (inTg) {
            setError(
              'Telegram не передал данные входа. Проверьте: 1) домен с https:// 2) тот же URL в BotFather и MINIAPP_PUBLIC_URL 3) откройте из @clutch_game_bot',
            );
          } else {
            setError('Открой приложение через Telegram (@clutch_game_bot)');
          }
          return;
        }

        const res = await authTelegram(initData);
        setToken(res.token.access_token);
        setUser(res.user);
        await tryAcceptInviteFromStartParam();
      } catch (e) {
        const msg = e instanceof Error ? e.message : 'Ошибка входа';
        if (isTelegramWebApp()) {
          setError(`${msg} (${telegramPlatform()})`);
        } else {
          setError(msg);
        }
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
      <div className="flex min-h-screen flex-col items-center justify-center gap-2 px-6 text-mut">
        <p className="text-sm font-semibold">Загрузка CLUTCH…</p>
        {isTelegramWebApp() && (
          <p className="text-center text-xs opacity-80">
            {telegramPlatform()}
            {!isSecureContext() ? ' · нужен HTTPS' : ''}
          </p>
        )}
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex min-h-screen items-center justify-center px-6 text-center text-sm leading-relaxed text-red">
        {error}
      </div>
    );
  }

  if (!user) {
    return (
      <div className="flex min-h-screen items-center justify-center px-6 text-center text-sm text-red">
        Не удалось загрузить профиль. Перезапусти Mini App.
      </div>
    );
  }

  if (!user.wallet_linked) {
    return (
      <AppKitInitProvider>
        <WalletGate onLinked={refreshUser} />
      </AppKitInitProvider>
    );
  }

  return (
    <AppKitInitProvider>
    <ErrorBoundary>
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
    </ErrorBoundary>
    </AppKitInitProvider>
  );
}
