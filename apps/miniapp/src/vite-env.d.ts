/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_URL: string;
  readonly VITE_WALLETCONNECT_PROJECT_ID: string;
  readonly VITE_APP_URL: string;
  readonly VITE_SOLANA_RPC_URL: string;
  readonly VITE_SOLANA_NETWORK: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}

interface TelegramWebApp {
  ready(): void;
  expand(): void;
  initData: string;
  initDataUnsafe: { start_param?: string };
  openTelegramLink?(url: string): void;
}

interface Window {
  Telegram?: { WebApp: TelegramWebApp };
}
