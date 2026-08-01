// Making a deploy visible on a device that has no reload button.
//
// The app is installed to a phone home screen, so there is no address bar to
// pull down and no way for the owner to force a refresh: a stale bundle stays
// on the phone until the operating system happens to kill the tab. That is not
// a cosmetic problem here -- the frontend and the backend ship together, and a
// phone running last week's form against this week's API is exactly the kind of
// silent mismatch this project spends its rules avoiding.
//
// The service worker calls skipWaiting() and claim(), so a new worker takes
// control of the open page as soon as it activates. This turns that moment
// into a reload, which is the only thing that swaps the page's own JavaScript.

/**
 * reloadWhenUpdated reloads the page once, when a NEW worker takes over an
 * already-controlled page.
 *
 * The `controller` check is what distinguishes an update from a first-ever
 * registration: the first worker on a device also fires controllerchange, and
 * reloading there would reload the first visit for no reason.
 *
 * The latch is its own flag rather than `{once: true}`: a reload loop on a
 * phone at an airfield would make the app unusable exactly where it is needed,
 * and that guarantee should not depend on an event-listener option one browser
 * might treat differently.
 */
export function reloadWhenUpdated(
  container: ServiceWorkerContainer,
  reload: () => void,
): void {
  if (!container.controller) return
  let reloaded = false
  container.addEventListener('controllerchange', () => {
    if (reloaded) return
    reloaded = true
    reload()
  })
}
