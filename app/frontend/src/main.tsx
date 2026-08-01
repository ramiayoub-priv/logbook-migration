import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { AuthProvider } from './auth'
import { App } from './App'
import { reloadWhenUpdated } from './swupdate'
import './styles.css'

const root = document.getElementById('root')
if (!root) throw new Error('no #root element to mount into')

// The service worker caches the app shell so the logbook opens at an airfield
// with no signal. It never caches an API response -- see public/sw.js and the
// policy tests in src/sw.test.ts.
//
// Registered only in a real build: in development the shell would be served
// from cache and every edit would appear not to take effect.
// A new worker taking control of this page means a deploy has landed, and the
// page reloads itself onto it -- see swupdate.ts. On a home-screen install
// there is no address bar to pull down, so without this the phone can sit on
// an old bundle indefinitely.
if ('serviceWorker' in navigator && import.meta.env.PROD) {
  reloadWhenUpdated(navigator.serviceWorker, () => window.location.reload())
  window.addEventListener('load', () => {
    navigator.serviceWorker
      .register('/logbook/sw.js', { scope: '/logbook/' })
      // Ask outright whether sw.js has changed. Registering an unchanged URL
      // is not by itself an update check on every browser, and this is the
      // request that starts the whole update -- new worker, then reload.
      .then((reg) => reg.update())
      .catch(() => {
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
