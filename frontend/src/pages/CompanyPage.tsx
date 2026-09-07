import { useEffect, useState, type FormEvent, type ReactNode } from 'react'
import { useSearchParams } from 'react-router-dom'
import { api } from '../api/client'
import type { Company, DriverDossier } from '../api/types'
import { Alert, Badge, Card, Facts, Field, Shell } from '../components/ui'
import { BrandMark } from '../components/Brand'

const TOKEN_KEY = 'dh_company_token'

export function CompanyPage() {
  const [params] = useSearchParams()
  const [token, setToken] = useState(() => localStorage.getItem(TOKEN_KEY) || '')
  const [company, setCompany] = useState<Company | null>(null)
  const [tab, setTab] = useState<'login' | 'reg'>('login')
  const [authError, setAuthError] = useState('')
  const [hub, setHub] = useState(() => params.get('hub') || 'DH-DEMOTST2')
  const [dossier, setDossier] = useState<DriverDossier | null>(null)
  const [error, setError] = useState('')

  useEffect(() => {
    const fromQr = params.get('hub')
    if (fromQr) setHub(fromQr)
  }, [params])

  useEffect(() => {
    if (!token) return
    api<Company>('/api/v1/me/company', { token })
      .then(setCompany)
      .catch(() => {
        localStorage.removeItem(TOKEN_KEY)
        setToken('')
        setCompany(null)
      })
  }, [token])

  useEffect(() => {
    if (!company || !params.get('hub')) return
    void (async () => {
      setError('')
      try {
        setDossier(await api<DriverDossier>(`/api/v1/drivers/${encodeURIComponent(hub)}/dossier`, { token }))
      } catch (err) {
        setDossier(null)
        setError(err instanceof Error ? err.message : 'Ошибка')
      }
    })()
  }, [company, params, hub, token])

  async function login(e: FormEvent<HTMLFormElement>) {
    e.preventDefault()
    setAuthError('')
    const fd = new FormData(e.currentTarget)
    try {
      const data = await api<{ company: Company; token: string }>('/api/v1/auth/company/login', {
        method: 'POST',
        body: { email: fd.get('email'), password: fd.get('password') },
      })
      localStorage.setItem(TOKEN_KEY, data.token)
      setToken(data.token)
      setCompany(data.company)
    } catch (err) {
      setAuthError(err instanceof Error ? err.message : 'Ошибка')
    }
  }

  async function register(e: FormEvent<HTMLFormElement>) {
    e.preventDefault()
    setAuthError('')
    const fd = new FormData(e.currentTarget)
    try {
      const data = await api<{ company: Company; token: string }>('/api/v1/auth/company/register', {
        method: 'POST',
        body: {
          name: fd.get('name'),
          inn: fd.get('inn'),
          email: fd.get('email'),
          password: fd.get('password'),
        },
      })
      localStorage.setItem(TOKEN_KEY, data.token)
      setToken(data.token)
      setCompany(data.company)
    } catch (err) {
      setAuthError(err instanceof Error ? err.message : 'Ошибка')
    }
  }

  async function loadDossier(e?: FormEvent) {
    e?.preventDefault()
    setError('')
    try {
      setDossier(await api<DriverDossier>(`/api/v1/drivers/${encodeURIComponent(hub)}/dossier`, { token }))
    } catch (err) {
      setDossier(null)
      setError(err instanceof Error ? err.message : 'Ошибка')
    }
  }

  async function act(path: string, body: unknown) {
    setError('')
    try {
      await api(`/api/v1/drivers/${encodeURIComponent(hub)}${path}`, { method: 'POST', token, body })
      await loadDossier()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка')
    }
  }

  function logout() {
    localStorage.removeItem(TOKEN_KEY)
    setToken('')
    setCompany(null)
    setDossier(null)
  }

  if (!company) {
    return (
      <Shell role="Компания" centered>
        <div className="auth-panel">
          <div className="auth-brand">
            <BrandMark size={40} />
          </div>
          <Card title="Вход компании">
            <div className="tabs">
              <button type="button" className={tab === 'login' ? 'active' : ''} onClick={() => setTab('login')}>
                Вход
              </button>
              <button type="button" className={tab === 'reg' ? 'active' : ''} onClick={() => setTab('reg')}>
                Регистрация
              </button>
            </div>
            {tab === 'login' ? (
              <form className="form" onSubmit={login}>
                <Field label="Email">
                  <input name="email" type="email" defaultValue="hr@logplus.ru" autoComplete="username" required />
                </Field>
                <Field label="Пароль">
                  <input name="password" type="password" defaultValue="demo1234" autoComplete="current-password" required />
                </Field>
                <button className="btn btn-primary btn-block" type="submit">
                  Войти
                </button>
              </form>
            ) : (
              <form className="form" onSubmit={register}>
                <Field label="Название">
                  <input name="name" required />
                </Field>
                <Field label="ИНН">
                  <input name="inn" inputMode="numeric" required />
                </Field>
                <Field label="Email">
                  <input name="email" type="email" required />
                </Field>
                <Field label="Пароль">
                  <input name="password" type="password" required />
                </Field>
                <button className="btn btn-primary btn-block" type="submit">
                  Создать
                </button>
              </form>
            )}
            <Alert kind="error">{authError}</Alert>
          </Card>
        </div>
      </Shell>
    )
  }

  const d = dossier?.driver

  return (
    <Shell role="Компания" onLogout={logout}>
      <Card title={company.name}>
        <p className="sub">ИНН {company.inn}</p>
      </Card>

      <section className="card">
        <h2>Поиск водителя</h2>
        <p className="sub" style={{ marginBottom: 14 }}>
          По Hub ID или QR с телефона водителя
        </p>
        <form className="form form-row" onSubmit={loadDossier}>
          <Field label="Hub ID" className="grow">
            <input value={hub} onChange={(e) => setHub(e.target.value)} required />
          </Field>
          <button className="btn btn-primary" type="submit">
            Открыть досье
          </button>
        </form>
        <Alert kind="error">{error}</Alert>
      </section>

      {d && (
        <section className="card">
          <div className="card-head">
            <h2>Досье</h2>
            <Badge status={d.verification_status} />
          </div>
          <Facts
            rows={[
              ['Hub ID', d.hub_id],
              ['ФИО', [d.last_name, d.first_name, d.middle_name].filter(Boolean).join(' ')],
              ['Категории', (d.categories || []).join(', ') || '—'],
              ['Стаж', `${d.experience_years || 0} лет`],
              ['ВУ', d.license ? `${d.license.series} ${d.license.number}` : '—'],
              ['СНИЛС', d.snils || '—'],
            ]}
          />

          {dossier && (
            <>
              <div className="columns">
                <Records
                  title="Рекомендации"
                  items={dossier.recommendations}
                  render={(x) => (
                    <>
                      <b>{x.company_name || 'компания'}</b> · {x.rating}/5
                      <br />
                      {x.text}
                    </>
                  )}
                />
                <Records
                  title="Жалобы"
                  items={dossier.complaints}
                  render={(x) => (
                    <>
                      <b>{x.severity}</b> · {x.category}
                      <br />
                      {x.text}
                    </>
                  )}
                />
                <Records
                  title="ДТП"
                  items={dossier.accidents}
                  render={(x) => (
                    <>
                      {x.description}
                      <br />
                      {x.fault} / {x.damage_level}
                    </>
                  )}
                />
                <Records
                  title="Штрафы"
                  items={dossier.fines}
                  render={(x) => (
                    <>
                      {x.article || 'штраф'} · {x.amount} ₽ · {x.paid ? 'оплачен' : 'не оплачен'}
                    </>
                  )}
                />
                <div className="span-2">
                  <Records
                    title="Чёрный список"
                    items={dossier.blacklist_entries}
                    render={(x) => (
                      <>
                        <b>{x.active ? 'активен' : 'снят'}</b> · {x.company_name || ''}
                        <br />
                        {x.reason}
                      </>
                    )}
                  />
                </div>
              </div>

              <div className="actions">
                <button type="button" className="btn btn-primary" onClick={() => act('/recommendations', { text: 'Рекомендуем к найму', rating: 5 })}>
                  Рекомендация
                </button>
                <button type="button" className="btn" onClick={() => act('/complaints', { category: 'discipline', text: 'Нарушение дисциплины', severity: 'low' })}>
                  Жалоба
                </button>
                <button type="button" className="btn btn-danger" onClick={() => act('/blacklist', { reason: 'Нарушения требований безопасности' })}>
                  Чёрный список
                </button>
                <button
                  type="button"
                  className="btn"
                  onClick={() =>
                    act('/accidents', {
                      occurred_at: new Date().toISOString(),
                      description: 'ДТП',
                      fault: 'unknown',
                      damage_level: 'minor',
                      location: '',
                    })
                  }
                >
                  ДТП
                </button>
                <button
                  type="button"
                  className="btn"
                  onClick={() =>
                    act('/fines', {
                      article: '12.9 КоАП',
                      amount: 5000,
                      issued_at: new Date().toISOString().slice(0, 10),
                      paid: false,
                      description: '',
                    })
                  }
                >
                  Штраф
                </button>
              </div>
            </>
          )}
        </section>
      )}
    </Shell>
  )
}

function Records<T>({ title, items, render }: { title: string; items: T[]; render: (item: T) => ReactNode }) {
  return (
    <div>
      <h3>{title}</h3>
      <ul className="list">
        {!items?.length ? <li className="empty">нет записей</li> : items.map((item, i) => <li key={i}>{render(item)}</li>)}
      </ul>
    </div>
  )
}
