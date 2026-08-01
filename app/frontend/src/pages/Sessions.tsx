import { useCallback, useState } from 'react'
import { api, ApiError } from '../api'
import { useApi, useAuth } from '../auth'

/**
 * The signed-in devices, and the way to sign out of any of them.
 *
 * Sessions are rows in the database rather than signed tokens precisely so
 * they can be withdrawn: the owner wants 90-day sessions, which makes a stolen
 * cookie a 90-day liability unless there is a page like this one
 * (docs/security.md).
 */
export function SessionsPage() {
  const { signOut } = useAuth()
  const load = useCallback(() => api.sessions(), [])
  const { data, error, loading, reload } = useApi(load, [])
  const [busy, setBusy] = useState<number | null>(null)
  const [problem, setProblem] = useState<string | null>(null)

  async function revoke(id: number, current: boolean) {
    if (current && !confirm('This is the device you are using. Signing it out will return you to the login page.')) {
      return
    }
    setBusy(id)
    setProblem(null)
    try {
      await api.revokeSession(id)
      if (current) {
        // The cookie is gone server-side; drop to the login page rather than
        // leaving a page that will 401 on its next request.
        await signOut()
        return
      }
      reload()
    } catch (e) {
      setProblem(e instanceof ApiError ? e.message : 'Could not sign that device out.')
    } finally {
      setBusy(null)
    }
  }

  return (
    <>
      <div className="card">
        <h2>Signed-in devices</h2>
        <p className="muted small">
          A session lasts 90 days from its last use and can be withdrawn here at any time.
          Changing your password on the server signs every device out.
        </p>
        <button onClick={() => void signOut()}>Sign out of this device</button>
      </div>

      {problem && <p className="error" role="alert">{problem}</p>}
      {error && <p className="error" role="alert">{error}</p>}
      {loading && !data && <p className="center">Loading your devices…</p>}

      {data?.sessions.map((s) => (
        <div className="card" key={s.id}>
          <strong>
            {s.current ? 'This device' : 'Another device'}
          </strong>
          <p className="muted small" style={{ margin: '6px 0' }}>
            {describeAgent(s.user_agent)}
            <br />
            Last used {formatWhen(s.last_used_at)} · from {s.ip || 'an unknown address'}
            <br />
            Signed in {formatWhen(s.created_at)} · expires {formatWhen(s.expires_at)}
          </p>
          <button className="link" disabled={busy === s.id} onClick={() => void revoke(s.id, s.current)}>
            {busy === s.id ? 'Signing out…' : s.current ? 'Sign out of this device' : 'Sign this device out'}
          </button>
        </div>
      ))}
    </>
  )
}

function formatWhen(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return 'an unknown time'
  // UTC, like every other instant this application shows (rule 0.4).
  return `${d.toISOString().slice(0, 10)} ${d.toISOString().slice(11, 16)} UTC`
}

/**
 * describeAgent shortens a user-agent to something a person can recognise.
 *
 * It is only ever a hint for "is this me?", so an unrecognised string is shown
 * as-is rather than guessed at -- and it is rendered as text by React, so a
 * hostile user-agent string cannot do anything but look odd.
 */
function describeAgent(ua: string): string {
  if (!ua) return 'An unidentified browser'
  const platform =
    /iPhone/.test(ua) ? 'iPhone'
    : /iPad/.test(ua) ? 'iPad'
    : /Android/.test(ua) ? 'Android'
    : /Macintosh/.test(ua) ? 'Mac'
    : /Windows/.test(ua) ? 'Windows'
    : /Linux/.test(ua) ? 'Linux'
    : null
  const browser =
    /Edg\//.test(ua) ? 'Edge'
    : /Chrome\//.test(ua) ? 'Chrome'
    : /Firefox\//.test(ua) ? 'Firefox'
    : /Safari\//.test(ua) ? 'Safari'
    : null
  if (platform && browser) return `${browser} on ${platform}`
  return ua.length > 80 ? `${ua.slice(0, 80)}…` : ua
}
