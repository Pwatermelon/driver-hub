import { useEffect, useState, type FormEvent, type ReactNode } from 'react'
import { useSearchParams } from 'react-router-dom'
import { api, errMessage } from '../api/client'
import type { Company, DriverDossier, Employment } from '../api/types'
import { AuthPanel } from '../components/AuthPanel'
import { Alert, Badge, Card, Facts, Field, Shell } from '../components/ui'

const TOKEN_KEY = 'dh_company_token'

function modBadge(status?: string) {
  if (status === 'approved') return <span className="badge badge-ok">модерация пройдена</span>
  if (status === 'rejected') return <span className="badge badge-warn">отклонена</span>
  return <span className="badge badge-warn">на модерации</span>
}

function empLabel(status?: string) {
  return (
    {
      pending: 'ожидает подтверждения водителя',
      confirmed: 'водитель подтвердил работу',
      rejected: 'водитель отклонил',
      ended: 'завершено',
    }[status || ''] || status || 'нет связи'
  )
}

export function CompanyPage() {
  const [params] = useSearchParams()
  const [token, setToken] = useState(() => localStorage.getItem(TOKEN_KEY) || '')
  const [company, setCompany] = useState<Company | null>(null)
  const [booting, setBooting] = useState(() => !!localStorage.getItem(TOKEN_KEY))
  const [tab, setTab] = useState<'login' | 'reg'>('login')
  const [authError, setAuthError] = useState('')
  const [busy, setBusy] = useState(false)
  const [searching, setSearching] = useState(false)
  const [hub, setHub] = useState(() => params.get('hub') || 'DH-DEMOTST2')
  const [dossier, setDossier] = useState<DriverDossier | null>(null)
  const [employments, setEmployments] = useState<Employment[]>([])
  const [error, setError] = useState('')
  const [okMsg, setOkMsg] = useState('')

  useEffect(() => {
    const fromQr = params.get('hub')
    if (fromQr) setHub(fromQr)
  }, [params])

  useEffect(() => {
    if (!token) {
      setBooting(false)
      return
    }
    let cancelled = false
    setBooting(true)
    Promise.all([
      api<Company>('/api/v1/me/company', { token }),
      api<Employment[]>('/api/v1/me/employments', { token }).catch(() => [] as Employment[]),
    ])
      .then(([c, emps]) => {
        if (!cancelled) {
          setCompany(c)
          setEmployments(emps)
        }
      })
      .catch(() => {
        if (!cancelled) {
          localStorage.removeItem(TOKEN_KEY)
          setToken('')
          setCompany(null)
        }
      })
      .finally(() => {
        if (!cancelled) setBooting(false)
      })
    return () => {
      cancelled = true
    }
  }, [token])

  useEffect(() => {
    if (!company || !params.get('hub')) return
    void loadDossier()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [company])

  async function login(e: FormEvent<HTMLFormElement>) {
    e.preventDefault()
    setAuthError('')
    setBusy(true)
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
      setAuthError(errMessage(err))
    } finally {
      setBusy(false)
    }
  }

  async function register(e: FormEvent<HTMLFormElement>) {
    e.preventDefault()
    setAuthError('')
    setBusy(true)
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
      setAuthError(errMessage(err))
    } finally {
      setBusy(false)
    }
  }

  async function refreshEmployments() {
    try {
      setEmployments(await api<Employment[]>('/api/v1/me/employments', { token }))
    } catch {
      /* ignore */
    }
  }

  async function loadDossier(e?: FormEvent) {
    e?.preventDefault()
    setError('')
    setOkMsg('')
    setSearching(true)
    try {
      setDossier(await api<DriverDossier>(`/api/v1/drivers/${encodeURIComponent(hub.trim())}/dossier`, { token }))
      await refreshEmployments()
    } catch (err) {
      setDossier(null)
      setError(errMessage(err))
    } finally {
      setSearching(false)
    }
  }

  async function claimEmployment() {
    setError('')
    setOkMsg('')
    try {
      await api(`/api/v1/drivers/${encodeURIComponent(hub.trim())}/employments`, { method: 'POST', token })
      setOkMsg('Запрос отправлен — водитель должен подтвердить работу в компании')
      await loadDossier()
    } catch (err) {
      setError(errMessage(err))
    }
  }

  async function act(path: string, body: unknown) {
    setError('')
    setOkMsg('')
    try {
      await api(`/api/v1/drivers/${encodeURIComponent(hub.trim())}${path}`, { method: 'POST', token, body })
      await loadDossier()
    } catch (err) {
      setError(errMessage(err))
    }
  }

  function logout() {
    localStorage.removeItem(TOKEN_KEY)
    setToken('')
    setCompany(null)
    setDossier(null)
    setEmployments([])
  }

  if (booting) {
    return (
      <Shell role="Компания" centered>
        <p className="boot-msg">Загрузка…</p>
      </Shell>
    )
  }

  if (!company) {
    return (
      <Shell role="Компания" centered>
        <AuthPanel
          title="Кабинет компании"
          tab={tab}
          onTab={(t) => {
            setTab(t)
            setAuthError('')
          }}
          error={authError}
          loading={busy}
          onLogin={login}
          onRegister={register}
          loginDefaults={{ email: 'hr@logplus.ru', password: 'demo1234' }}
          registerExtra={
            <>
              <Field label="Название">
                <input name="name" required disabled={busy} />
              </Field>
              <Field label="ИНН">
                <input name="inn" inputMode="numeric" required disabled={busy} />
              </Field>
            </>
          }
        />
      </Shell>
    )
  }

  const approved = company.moderation_status === 'approved' || company.verified
  const d = dossier?.driver
  const emp = dossier?.viewer_employment
  const canWrite = approved && emp?.status === 'confirmed'

  return (
    <Shell role="Компания" onLogout={logout}>
      <Card title={company.name} badge={modBadge(company.moderation_status)}>
        <p className="sub">ИНН {company.inn}</p>
        {!approved && (
          <Alert kind="error">
            {company.moderation_status === 'rejected'
              ? 'Регистрация отклонена модератором. Жалобы и рекомендации недоступны.'
              : 'Компания на модерации. После одобрения можно будет указывать водителей и писать отзывы.'}
          </Alert>
        )}
      </Card>

      {employments.filter((e) => e.status === 'pending' || e.status === 'confirmed').length > 0 && (
        <section className="card">
          <h2>Ваши водители</h2>
          <ul className="list">
            {employments
              .filter((e) => e.status === 'pending' || e.status === 'confirmed')
              .map((e) => (
                <li key={e.id}>
                  <b>{e.driver_name || 'Водитель'}</b> · {e.driver_hub_id}
                  <br />
                  <span className="sub">{empLabel(e.status)}</span>
                </li>
              ))}
          </ul>
        </section>
      )}

      <section className="card">
        <h2>Поиск водителя</h2>
        <p className="sub" style={{ marginBottom: 14 }}>
          Введите Hub ID или отсканируйте QR с телефона водителя
        </p>
        <form className="form form-row" onSubmit={loadDossier}>
          <Field label="Hub ID" className="grow">
            <input value={hub} onChange={(e) => setHub(e.target.value)} required disabled={searching} />
          </Field>
          <button className="btn btn-primary" type="submit" disabled={searching}>
            {searching ? 'Ищем…' : 'Открыть досье'}
          </button>
        </form>
        <Alert kind="error">{error}</Alert>
        <Alert kind="ok">{okMsg}</Alert>
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
              ['Работа у вас', empLabel(emp?.status)],
            ]}
          />

          {approved && emp?.status !== 'confirmed' && emp?.status !== 'pending' && (
            <div className="verify-callout" style={{ marginTop: 16 }}>
              <div>
                <strong>Указать, что водитель работает у вас</strong>
                <p>Водитель получит запрос и должен подтвердить. После этого можно писать жалобы и рекомендации.</p>
              </div>
              <button type="button" className="btn btn-primary" onClick={claimEmployment}>
                Отправить запрос
              </button>
            </div>
          )}
          {emp?.status === 'pending' && (
            <Alert kind="ok">Запрос отправлен. Ждём подтверждения водителя.</Alert>
          )}

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

              {canWrite ? (
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
              ) : (
                <p className="sub" style={{ marginTop: 16 }}>
                  Жалобы и рекомендации доступны только после подтверждения работы водителя в вашей компании.
                </p>
              )}
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
