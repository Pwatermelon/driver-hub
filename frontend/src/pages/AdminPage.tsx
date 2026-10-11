import { useEffect, useState, type FormEvent } from 'react'
import { api, errMessage } from '../api/client'
import type { Company } from '../api/types'
import { Alert, Field } from '../components/ui'

const TOKEN_KEY = 'dh_admin_token'

function modLabel(s?: string) {
  return (
    {
      pending: 'на модерации',
      approved: 'одобрена',
      rejected: 'отклонена',
    }[s || ''] || s || '—'
  )
}

export function AdminPage() {
  const [token, setToken] = useState(() => localStorage.getItem(TOKEN_KEY) || '')
  const [password, setPassword] = useState('')
  const [authError, setAuthError] = useState('')
  const [busy, setBusy] = useState(false)
  const [filter, setFilter] = useState<'pending' | 'approved' | 'rejected' | ''>('pending')
  const [companies, setCompanies] = useState<Company[]>([])
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  async function load(t = token, status = filter) {
    setError('')
    setLoading(true)
    try {
      const q = status ? `?status=${encodeURIComponent(status)}` : ''
      setCompanies(await api<Company[]>(`/api/v1/admin/companies${q}`, { token: t }))
    } catch (err) {
      setCompanies([])
      const msg = errMessage(err)
      if (msg.includes('авторизац') || msg.includes('токен') || msg.includes('прав')) {
        localStorage.removeItem(TOKEN_KEY)
        setToken('')
      }
      setError(msg)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    if (token) void load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [token, filter])

  async function login(e: FormEvent) {
    e.preventDefault()
    setAuthError('')
    setBusy(true)
    try {
      const data = await api<{ token: string }>('/api/v1/auth/admin/login', {
        method: 'POST',
        body: { password },
      })
      localStorage.setItem(TOKEN_KEY, data.token)
      setToken(data.token)
      setPassword('')
    } catch (err) {
      setAuthError(errMessage(err))
    } finally {
      setBusy(false)
    }
  }

  async function moderate(id: string, action: 'approve' | 'reject') {
    setError('')
    try {
      await api(`/api/v1/admin/companies/${id}/${action}`, { method: 'POST', token })
      await load()
    } catch (err) {
      setError(errMessage(err))
    }
  }

  function logout() {
    localStorage.removeItem(TOKEN_KEY)
    setToken('')
    setCompanies([])
  }

  if (!token) {
    return (
      <div className="shell shell-auth">
        <main className="auth-stage">
          <section className="card auth-card">
            <h1>Админ-панель</h1>
            <p className="sub">Модерация компаний Driver Hub</p>
            <form className="form" onSubmit={login}>
              <Field label="Пароль">
                <input
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required
                  disabled={busy}
                  autoComplete="current-password"
                />
              </Field>
              <Alert kind="error">{authError}</Alert>
              <button className="btn btn-primary" type="submit" disabled={busy}>
                {busy ? 'Вход…' : 'Войти'}
              </button>
            </form>
          </section>
        </main>
      </div>
    )
  }

  return (
    <div className="shell">
      <header className="topbar">
        <strong>Driver Hub · Admin</strong>
        <div className="topbar-right">
          <button type="button" className="text-btn" onClick={logout}>
            Выйти
          </button>
        </div>
      </header>
      <main className="page">
        <section className="card">
          <div className="card-head">
            <h1>Компании</h1>
          </div>
          <div className="actions" style={{ marginBottom: 16 }}>
            {(
              [
                ['pending', 'На модерации'],
                ['approved', 'Одобренные'],
                ['rejected', 'Отклонённые'],
                ['', 'Все'],
              ] as const
            ).map(([value, label]) => (
              <button
                key={label}
                type="button"
                className={`btn ${filter === value ? 'btn-primary' : ''}`}
                onClick={() => setFilter(value)}
              >
                {label}
              </button>
            ))}
          </div>
          <Alert kind="error">{error}</Alert>
          {loading ? (
            <p className="sub">Загрузка…</p>
          ) : (
            <ul className="list">
              {!companies.length ? (
                <li className="empty">нет компаний</li>
              ) : (
                companies.map((c) => (
                  <li key={c.id}>
                    <div className="card-head" style={{ marginBottom: 8 }}>
                      <div>
                        <b>{c.name}</b>
                        <p className="sub" style={{ margin: '4px 0 0' }}>
                          ИНН {c.inn} · {c.email} · {modLabel(c.moderation_status)}
                        </p>
                      </div>
                      {c.moderation_status === 'pending' && (
                        <div className="actions">
                          <button type="button" className="btn btn-primary" onClick={() => moderate(c.id, 'approve')}>
                            Одобрить
                          </button>
                          <button type="button" className="btn btn-danger" onClick={() => moderate(c.id, 'reject')}>
                            Отклонить
                          </button>
                        </div>
                      )}
                      {c.moderation_status === 'rejected' && (
                        <button type="button" className="btn btn-primary" onClick={() => moderate(c.id, 'approve')}>
                          Одобрить
                        </button>
                      )}
                      {c.moderation_status === 'approved' && (
                        <button type="button" className="btn" onClick={() => moderate(c.id, 'reject')}>
                          Отозвать
                        </button>
                      )}
                    </div>
                  </li>
                ))
              )}
            </ul>
          )}
        </section>
      </main>
    </div>
  )
}
