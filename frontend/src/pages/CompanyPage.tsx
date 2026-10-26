import { useCallback, useEffect, useMemo, useState, type FormEvent, type ReactNode } from 'react'
import { useSearchParams } from 'react-router-dom'
import { api, errMessage } from '../api/client'
import type { Company, DriverDossier, Employment } from '../api/types'
import { AuthPanel } from '../components/AuthPanel'
import { QrScanner } from '../components/QrScanner'
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
      pending: 'ожидает подтверждения',
      confirmed: 'работа подтверждена',
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
  const [hub, setHub] = useState(() => params.get('hub') || '')
  const [dossier, setDossier] = useState<DriverDossier | null>(null)
  const [employments, setEmployments] = useState<Employment[]>([])
  const [driverFilter, setDriverFilter] = useState('')
  const [scanOpen, setScanOpen] = useState(false)
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

  const openDossier = useCallback(
    async (hubID: string) => {
      const id = hubID.trim()
      if (!id) return
      setHub(id)
      setError('')
      setOkMsg('')
      setSearching(true)
      try {
        setDossier(await api<DriverDossier>(`/api/v1/drivers/${encodeURIComponent(id)}/dossier`, { token }))
        await refreshEmployments()
      } catch (err) {
        setDossier(null)
        setError(errMessage(err))
      } finally {
        setSearching(false)
      }
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [token],
  )

  async function loadDossier(e?: FormEvent) {
    e?.preventDefault()
    await openDossier(hub)
  }

  const handleScan = useCallback(
    (hubId: string) => {
      void openDossier(hubId)
    },
    [openDossier],
  )
  const closeScanner = useCallback(() => setScanOpen(false), [])

  async function claimEmployment() {
    setError('')
    setOkMsg('')
    try {
      await api(`/api/v1/drivers/${encodeURIComponent(hub.trim())}/employments`, { method: 'POST', token })
      setOkMsg('Запрос отправлен — водитель должен подтвердить работу в компании')
      await openDossier(hub)
    } catch (err) {
      setError(errMessage(err))
    }
  }

  async function act(path: string, body: unknown) {
    setError('')
    setOkMsg('')
    try {
      await api(`/api/v1/drivers/${encodeURIComponent(hub.trim())}${path}`, { method: 'POST', token, body })
      await openDossier(hub)
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

  const myDrivers = useMemo(
    () => employments.filter((e) => e.status === 'pending' || e.status === 'confirmed'),
    [employments],
  )

  const filteredDrivers = useMemo(() => {
    const q = driverFilter.trim().toLowerCase()
    if (!q) return myDrivers
    return myDrivers.filter((e) => {
      const hay = `${e.driver_name || ''} ${e.driver_hub_id || ''} ${empLabel(e.status)}`.toLowerCase()
      return hay.includes(q)
    })
  }, [myDrivers, driverFilter])

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
  const activeHub = d?.hub_id || hub.trim()

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

      <div className="company-workspace">
        <aside className="company-sidebar card">
          <div className="card-head">
            <h2>Ваши водители</h2>
            <span className="count-chip">{myDrivers.length}</span>
          </div>

          <label className="driver-filter">
            <span className="sr-only">Поиск по списку</span>
            <input
              value={driverFilter}
              onChange={(e) => setDriverFilter(e.target.value)}
              placeholder="Поиск по имени или Hub ID"
            />
          </label>

          <div className="driver-list-scroll">
            {myDrivers.length === 0 ? (
              <p className="empty-state">Пока нет связанных водителей. Отсканируйте QR или найдите по Hub ID.</p>
            ) : filteredDrivers.length === 0 ? (
              <p className="empty-state">Никого не найдено по запросу «{driverFilter}».</p>
            ) : (
              <ul className="driver-list">
                {filteredDrivers.map((e) => {
                  const selected = !!e.driver_hub_id && e.driver_hub_id === activeHub
                  return (
                    <li key={e.id}>
                      <button
                        type="button"
                        className={`driver-item ${selected ? 'active' : ''}`}
                        disabled={!e.driver_hub_id || searching}
                        onClick={() => void openDossier(e.driver_hub_id || '')}
                      >
                        <span className="driver-item-name">{e.driver_name || 'Водитель'}</span>
                        <span className="driver-item-hub">{e.driver_hub_id}</span>
                        <span className={`driver-item-status ${e.status === 'confirmed' ? 'ok' : 'wait'}`}>
                          {empLabel(e.status)}
                        </span>
                      </button>
                    </li>
                  )
                })}
              </ul>
            )}
          </div>
        </aside>

        <div className="company-main">
          <section className="card company-search">
            <h2>Найти водителя</h2>
            <p className="sub" style={{ marginBottom: 14 }}>
              Сканируйте QR с телефона водителя или введите Hub ID вручную
            </p>
            <form className="form form-row company-search-row" onSubmit={loadDossier}>
              <Field label="Hub ID" className="grow">
                <input
                  value={hub}
                  onChange={(e) => setHub(e.target.value)}
                  placeholder="DH-…"
                  required
                  disabled={searching}
                />
              </Field>
              <button type="button" className="btn btn-scan" onClick={() => setScanOpen(true)} disabled={searching}>
                Сканировать QR
              </button>
              <button className="btn btn-primary" type="submit" disabled={searching}>
                {searching ? 'Ищем…' : 'Открыть досье'}
              </button>
            </form>
            <Alert kind="error">{error}</Alert>
            <Alert kind="ok">{okMsg}</Alert>
          </section>

          <section className="card company-dossier">
            {!d && !searching && (
              <div className="dossier-placeholder">
                <h2>Досье водителя</h2>
                <p className="sub">
                  Выберите водителя слева или нажмите «Сканировать QR», чтобы открыть карточку с телефона водителя.
                </p>
              </div>
            )}

            {searching && (
              <div className="dossier-placeholder">
                <h2>Досье водителя</h2>
                <p className="sub">Загружаем карточку…</p>
              </div>
            )}

            {d && !searching && (
              <>
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
                {emp?.status === 'pending' && <Alert kind="ok">Запрос отправлен. Ждём подтверждения водителя.</Alert>}

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
                        <button
                          type="button"
                          className="btn btn-primary"
                          onClick={() => act('/recommendations', { text: 'Рекомендуем к найму', rating: 5 })}
                        >
                          Рекомендация
                        </button>
                        <button
                          type="button"
                          className="btn"
                          onClick={() => act('/complaints', { category: 'discipline', text: 'Нарушение дисциплины', severity: 'low' })}
                        >
                          Жалоба
                        </button>
                        <button
                          type="button"
                          className="btn btn-danger"
                          onClick={() => act('/blacklist', { reason: 'Нарушения требований безопасности' })}
                        >
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
              </>
            )}
          </section>
        </div>
      </div>

      <button type="button" className="scan-fab" onClick={() => setScanOpen(true)} aria-label="Сканировать QR водителя">
        QR
      </button>

      {scanOpen && <QrScanner onClose={closeScanner} onScan={handleScan} />}
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
