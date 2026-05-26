import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  base: '/',
  define: {
    global: 'globalThis',
  },
  resolve: {
    alias: {
      buffer: 'buffer',
    },
  },
  optimizeDeps: {
    include: [
      'buffer',
      '@reown/appkit/react',
      '@reown/appkit-adapter-solana/react',
      '@reown/appkit-wallet-button/react',
      '@walletconnect/universal-provider',
    ],
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true,
    sourcemap: false,
    reportCompressedSize: false,
    commonjsOptions: {
      transformMixedEsModules: true,
    },
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (!id.includes('node_modules')) return;
          if (
            id.includes('@reown') ||
            id.includes('@walletconnect') ||
            id.includes('viem')
          ) {
            return 'wallet-vendor';
          }
          if (id.includes('@solana')) return 'solana-vendor';
          return 'vendor';
        },
      },
    },
  },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: process.env.VITE_DEV_API || 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
});
