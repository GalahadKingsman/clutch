import { openWalletHref, phantomWalletConnectUrl } from './telegram-wallet-bridge';

const WC_URI_RE = /wc:[a-f0-9]+@[a-f0-9]+(?:\?[^\s"'<>]*)?/i;

function collectRoots(): Array<Document | ShadowRoot> {
  const roots: Array<Document | ShadowRoot> = [document];
  document.querySelectorAll('*').forEach((el) => {
    if (el.shadowRoot) roots.push(el.shadowRoot);
  });
  return roots;
}

function extractWcUriFromRoot(root: Document | ShadowRoot): string | null {
  const anchors = root.querySelectorAll<HTMLAnchorElement>('a[href]');
  for (const a of anchors) {
    const href = decodeURIComponent(a.href);
    const m = href.match(WC_URI_RE);
    if (m) return m[0];
    const uriParam = href.match(/uri=(wc%3A[^&]+)/i);
    if (uriParam) {
      try {
        return decodeURIComponent(uriParam[1].replace(/\+/g, '%20'));
      } catch {
        /* skip */
      }
    }
  }

  const html =
    root instanceof Document
      ? root.documentElement.innerHTML
      : (root as ShadowRoot).innerHTML;
  const textMatch = html.match(WC_URI_RE);
  if (textMatch) return textMatch[0];

  return null;
}

/** Ищем WC URI в модалке Reown (включая shadow DOM) и открываем Phantom. */
export function watchAndOpenPhantom(timeoutMs = 25_000): () => void {
  let opened = false;

  const tick = () => {
    if (opened) return;
    for (const root of collectRoots()) {
      const uri = extractWcUriFromRoot(root);
      if (!uri) continue;
      opened = true;
      openWalletHref(phantomWalletConnectUrl(uri));
      return;
    }
  };

  tick();
  const interval = window.setInterval(tick, 350);
  const timeout = window.setTimeout(() => {
    window.clearInterval(interval);
  }, timeoutMs);

  return () => {
    window.clearInterval(interval);
    window.clearTimeout(timeout);
  };
}
