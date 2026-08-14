// The service worker, which now exists only to remove itself.
//
// ⛔ THIS FILE MUST KEEP BEING DEPLOYED. It is tempting to delete it now that
// nothing registers a worker -- do not. Deleting sw.js from the server does NOT
// remove a worker already installed on a device: the browser goes on running
// the copy it has, serving the shell it cached, indefinitely. The ONLY thing
// that reliably retires an installed worker is a new worker at the same URL
// that cleans up and unregisters itself. That is this file. It must stay until
// there is no device left that could still be carrying the old one, and there
// is no way to know that, so: it stays.
//
// WHY THE CACHING WENT (owner, 2026-08-14): "Can you make sure NOTHING is
// cached at all? Like the browser needs to forget (except the cookie for the
// session)." The previous worker cached the app shell so the logbook would open
// at an airfield with no signal. Twice that put a stale build in front of the
// owner, and the second time cost an afternoon of "is my phone caching it?"
// when the real answer was that a deploy had not been uploaded. Against that,
// what the cache actually bought was small: offline WRITES were never in scope
// (app/APP.md section 2), so an offline shell could only ever open an app that
// then failed every request it made.
//
// The trade is therefore stated plainly rather than hidden: THE APP NO LONGER
// OPENS WITHOUT A NETWORK. It is always the current build instead. The owner
// asked for exactly that.
//
// The session cookie is untouched by any of this. It is HttpOnly and lives in
// the cookie store, which is not a cache; clearing caches does not sign anyone
// out, and nothing here reads or writes it.

// Take over from the old worker at once. Waiting for it to die would leave it
// serving its cached shell until every tab was closed -- and a home-screen app
// is a tab that is never closed.
self.addEventListener('install', () => self.skipWaiting())

self.addEventListener('activate', (event) => {
  event.waitUntil(
    (async () => {
      // Every cache, by name, whatever it is called. Not a filter on the names
      // this project happened to use: the point is that the device is left
      // holding nothing, including anything an older version of this file
      // created under a name nobody remembers.
      for (const key of await caches.keys()) {
        await caches.delete(key)
      }

      // Then remove the registration itself, so the next load is plain network
      // with no worker in the way. Safe from a re-register loop because the app
      // no longer registers anything (src/main.tsx) -- the only code that could
      // put a worker back is an old cached bundle, which the line above has
      // just deleted.
      await self.registration.unregister()

      // Finally, move the open page onto the network. Without this the phone in
      // the owner's hand keeps running the JavaScript it already loaded, which
      // is the stale build they were complaining about; the cleanup would only
      // show up whenever the OS next killed the app. navigate() rather than a
      // postMessage-and-reload dance because there is no page code left to
      // co-operate with -- this worker may be the only new thing on the device.
      for (const client of await self.clients.matchAll({ type: 'window' })) {
        await client.navigate(client.url)
      }
    })(),
  )
})

// There is deliberately NO fetch handler. A worker that intercepts a request is
// a worker that can answer it from storage; this one cannot answer anything, so
// every request -- shell, asset and API alike -- goes to the network exactly as
// if no service worker had ever been installed.
