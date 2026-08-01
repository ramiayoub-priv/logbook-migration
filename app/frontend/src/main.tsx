import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { AuthProvider } from './auth'
import { App } from './App'
import './styles.css'

const root = document.getElementById('root')
if (!root) throw new Error('no #root element to mount into')

// The service worker caches the app shell so the logbook opens at an airfield
// with no signal. It never caches an API response -- see public/sw.js and the
// policy tests in src/sw.test.ts.
//
// Registered only in a real build: in development the shell would be served
// from cache and every edit would appear not to take effect.
if ('serviceWorker' in navigator && import.meta.env.PROD) {
  window.addEventListener('load', () => {
    navigator.serviceWorker.register('/logbook/sw.js', { scope: '/logbook/' }).catch(() => {
      // An unregistered worker costs offline start-up and nothing else, so a
      // failure here must never stop the app from loading.
    })
  })
}

createRoot(root).render(
  <StrictMode>
    <AuthProvider>
      <App />
    </AuthProvider>
  </StrictMode>,
)
