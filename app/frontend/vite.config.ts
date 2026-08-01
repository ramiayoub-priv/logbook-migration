import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// The app is served from https://ayoub.fi/logbook, not from the root of the
// host: the box also serves the owner's other sites. Every asset URL therefore
// has to be built with that prefix, and the API lives under /logbook/api
// because plain /api is already taken by a stale transit proxy (docs/deploy.md).
export default defineConfig({
  base: '/logbook/',
  plugins: [react()],
  build: {
    outDir: 'dist',
    // Source maps would publish the whole frontend source next to the bundle
    // on a public host. Nothing here is secret, but there is no reason to
    // serve it either.
    sourcemap: false,
  },
  server: {
    port: 5173,
    // `npm run dev` talks to a locally running API so the cookie stays
    // same-origin, exactly as it is in production. Proxying rather than
    // enabling CORS keeps the server's Origin check meaningful in development.
    proxy: {
      '/logbook/api': {
        target: 'http://127.0.0.1:8099',
        changeOrigin: false,
      },
    },
  },
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./src/test-setup.ts'],
    css: false,
  },
})
