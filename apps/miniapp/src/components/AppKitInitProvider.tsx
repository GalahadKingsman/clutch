import {
  createContext,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from 'react';
import { initAppKit, walletConnectConfigured } from '../lib/appkit-init';

type AppKitState = {
  ready: boolean;
  configured: boolean;
  error: string | null;
};

const AppKitCtx = createContext<AppKitState>({
  ready: false,
  configured: false,
  error: null,
});

export function useAppKitInit() {
  return useContext(AppKitCtx);
}

export function AppKitInitProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<AppKitState>({
    ready: false,
    configured: walletConnectConfigured,
    error: null,
  });

  useEffect(() => {
    if (!walletConnectConfigured) {
      setState({
        ready: true,
        configured: false,
        error: 'VITE_WALLETCONNECT_PROJECT_ID не в сборке',
      });
      return;
    }

    const result = initAppKit();
    setState({
      ready: true,
      configured: true,
      error: result.ok ? null : result.error ?? 'Ошибка WalletConnect',
    });
  }, []);

  if (!state.ready) {
    return (
      <div className="flex min-h-screen items-center justify-center px-6 text-sm text-gold">
        Загрузка WalletConnect…
      </div>
    );
  }

  return <AppKitCtx.Provider value={state}>{children}</AppKitCtx.Provider>;
}
