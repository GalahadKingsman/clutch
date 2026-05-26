export function isTelegramWebApp(): boolean {
  return typeof window !== 'undefined' && Boolean(window.Telegram?.WebApp);
}

/** Ждём initData — в WebView он иногда появляется после ready(). */
export function waitForInitData(timeoutMs = 6000): Promise<string> {
  return new Promise((resolve, reject) => {
    const tg = window.Telegram?.WebApp;
    if (!tg) {
      reject(new Error('not_telegram'));
      return;
    }
    tg.ready();
    const deadline = Date.now() + timeoutMs;
    const tick = () => {
      const data = tg.initData?.trim();
      if (data) {
        resolve(data);
        return;
      }
      if (Date.now() >= deadline) {
        reject(new Error('init_data_timeout'));
        return;
      }
      window.setTimeout(tick, 80);
    };
    tick();
  });
}

export function telegramPlatform(): string {
  return window.Telegram?.WebApp?.platform ?? 'unknown';
}
