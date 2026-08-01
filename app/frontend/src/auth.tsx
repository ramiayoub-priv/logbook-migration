import { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react'
import { api, ApiError, type User } from './api'

// Who is signed in.
//
// There is no token here and no persisted state. The session lives entirely in
// an HttpOnly cookie the page cannot read, so "am I signed in?" is answered by
// asking the server (GET /me) rather than by inspecting anything local. That
// is slower by one request at startup and it is the honest answer: a cached
// "logged in" flag would survive a revoked session and show an empty logbook
// with no explanation.

type State =
  | { status: 'checking' }
  | { status: 'anonymous' }
  | { status: 'signedIn'; user: User }

interface AuthValue {
  state: State
  signIn: (username: string, password: string) => Promise<void>
  signOut: () => Promise<void>
  /** Called when any request comes back 401, so the app drops to the login page. */
  sessionExpired: () => void
}

const AuthContext = createContext<AuthValue | null>(null)

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [state, setState] = useState<State>({ status: 'checking' })

  useEffect(() => {
    let live = true
    api
      .me()
      .then((user) => live && setState({ status: 'signedIn', user }))
      .catch(() => live && setState({ status: 'anonymous' }))
    return () => {
      live = false
    }
  }, [])

  const signIn = useCallback(async (username: string, password: string) => {
    const user = await api.login(username, password)
    setState({ status: 'signedIn', user })
  }, [])

  const signOut = useCallback(async () => {
    try {
      await api.logout()
    } catch {
      // A logout that fails because the session was already gone has still
      // achieved what the user asked for. Anything else is not worth keeping
      // them on a page they wanted to leave.
    }
    setState({ status: 'anonymous' })
  }, [])

  const sessionExpired = useCallback(() => setState({ status: 'anonymous' }), [])

  const value = useMemo(
    () => ({ state, signIn, signOut, sessionExpired }),
    [state, signIn, signOut, sessionExpired],
  )
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth(): AuthValue {
  const v = useContext(AuthContext)
  if (!v) throw new Error('useAuth used outside AuthProvider')
  return v
}

/**
 * useApi runs a request and reports loading, data and error as one value.
 *
 * A 401 anywhere means the session ended -- expired, revoked from another
 * device, or the account disabled -- and the whole app drops to the login
 * page rather than the page showing an error it cannot recover from.
 */
export function useApi<T>(fn: () => Promise<T>, deps: unknown[]): {
  data: T | null
  error: string | null
  loading: boolean
  reload: () => void
} {
  const { sessionExpired } = useAuth()
  const [data, setData] = useState<T | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [nonce, setNonce] = useState(0)

  useEffect(() => {
    let live = true
    setLoading(true)
    setError(null)
    fn()
      .then((d) => {
        if (!live) return
        setData(d)
        setLoading(false)
      })
      .catch((e: unknown) => {
        if (!live) return
        if (e instanceof ApiError && e.isUnauthenticated) {
          sessionExpired()
          return
        }
        // The error is shown; the previous data is left alone rather than
        // cleared, so a failed refresh does not blank a page of real figures.
        setError(e instanceof Error ? e.message : 'Something went wrong.')
        setLoading(false)
      })
    return () => {
      live = false
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [...deps, nonce])

  const reload = useCallback(() => setNonce((n) => n + 1), [])
  return { data, error, loading, reload }
}
