import { useState } from 'react'
import { ApiError } from '../api'
import { useAuth } from '../auth'

export function LoginPage() {
  const { signIn } = useAuth()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [busy, setBusy] = useState(false)

  async function submit(e: React.FormEvent) {
    e.preventDefault()
    setBusy(true)
    setError(null)
    try {
      await signIn(username, password)
    } catch (err) {
      // The server answers a wrong username and a wrong password identically,
      // and takes the same time over both, so the login form cannot be used to
      // discover whether an account exists. This message must stay just as
      // uninformative -- saying "no such user" here would undo the control
      // (docs/security.md).
      if (err instanceof ApiError && err.status === 429) {
        setError(
          err.retryAfter
            ? `Too many attempts. Try again in ${err.retryAfter} seconds.`
            : 'Too many attempts. Try again shortly.',
        )
      } else if (err instanceof ApiError && err.status === 0) {
        setError(err.message)
      } else {
        setError('Wrong username or password.')
      }
      setBusy(false)
    }
  }

  return (
    <div className="login">
      <form className="card" onSubmit={submit}>
        <h2>Logbook</h2>
        <p className="muted small">Sign in to see your flights.</p>

        {error && (
          <p className="error" role="alert">
            {error}
          </p>
        )}

        <div className="field">
          <label htmlFor="username">Username</label>
          <input
            id="username"
            name="username"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            autoComplete="username"
            autoCapitalize="none"
            autoCorrect="off"
            spellCheck={false}
            required
          />
        </div>

        <div className="field">
          <label htmlFor="password">Password</label>
          <input
            id="password"
            name="password"
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            autoComplete="current-password"
            required
          />
        </div>

        <button className="primary" type="submit" disabled={busy}>
          {busy ? 'Signing in…' : 'Sign in'}
        </button>

        <p className="muted small" style={{ marginBottom: 0 }}>
          Accounts are created on the server. There is no sign-up.
        </p>
      </form>
    </div>
  )
}
