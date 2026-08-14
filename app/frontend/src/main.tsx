import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { AuthProvider } from './auth'
import { App } from './App'
import './styles.css'

const root = document.getElementById('root')
if (!root) throw new Error('no #root element to mount into')

// NO SERVICE WORKER IS REGISTERED HERE, deliberately, and adding one back is a
// decision rather than a fix.
//
// Until 2026-08-14 this file registered `public/sw.js`, which cached the app
// shell so the logbook would open at an airfield with no signal. The owner
// ended it: "Can you make sure NOTHING is cached at all? Like the browser needs
// to forget (except the cookie for the session)." Twice a cached shell had put
// a stale build in front of them, and offline WRITES were never in scope
// (app/APP.md section 2) — so the cache could only ever open an app that then
// failed every request it made.
//
// `public/sw.js` is still deployed and is now a kill switch: it deletes every
// cache and unregisters itself, which is the only way to retire a worker that
// is already installed on a device. Registering it from here would install it
// afresh on every load — unregister, navigate, register, forever. There is a
// test that fails if this line comes back (src/noworker.test.ts).

createRoot(root).render(
  <StrictMode>
    <AuthProvider>
      <App />
    </AuthProvider>
  </StrictMode>,
)
