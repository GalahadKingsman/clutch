/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_URL: string;
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
