import { Buffer } from 'buffer';
import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import App from './App';
import './index.css';
import { installTelegramWalletBridge } from './lib/telegram-wallet-bridge';

globalThis.Buffer = Buffer;
installTelegramWalletBridge();

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
);
