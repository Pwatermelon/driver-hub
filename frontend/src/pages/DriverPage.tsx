import { useEffect, useMemo, useState, type FormEvent } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { QRCodeSVG } from 'qrcode.react'
import { api } from '../api/client'
import type { Driver } from '../api/types'
import { Alert, Badge, Card, Facts, Field, Shell } from '../components/ui'
import { BrandMark } from '../components/Brand'

const TOKEN_KEY = 'dh_driver_token'

export function DriverPage() {
  const [params, setParams] = useSearchParams()
  const navigate = useNavigate()
  const [token, setToken] = useState(() => localStorage.getItem(TOKEN_KEY) || '')
  const [driver, setDriver] = useState<Driver | null>(null)
  const [tab, setTab] = useState<'login' | 'reg'>('login')
  const [authError, setAuthError] = useState('')
  const [copied, setCopied] = useState(false)

  useEffect(() => {
    const t = params.get('token')
    if (params.get('esia_ok') && t) {
      localStorage.setItem(TOKEN_KEY, t)
      setToken(t)
      setParams({}, { replace: true })
    }
  }, [params, setParams])

  useEffect(() => {
    if (!token) return
    api<Driver>('/api/v1/me/driver', { token })
      .then(setDriver)
      .catch(() => {
        localStorage.removeItem(TOKEN_KEY)
        setToken('')
        setDriver(null)
      })
  }, [token])

  const qrValue = useMemo(() => {
    if (!driver) return ''
    const base = window.location.origin
    return `${base}/company?hub=${encodeURIComponent(driver.hub_id)}`
  }, [driver])

  async function login(e: FormEvent<HTMLFormElement>) {
    e.preventDefault()
    setAuthError('')
    const fd = new FormData(e.currentTarget)
    try {
      const data = await api<{ driver: Driver; token: string }>('/api/v1/auth/driver/login', {
        method: 'POST',
        body: { email: fd.get('email'), password: fd.get('password') },
      })
      localStorage.setItem(TOKEN_KEY, data.token)
      setToken(data.token)
      setDriver(data.driver)
    } catch (err) {
      setAuthError(err instanceof Error ? err.message : 'Ошибка')
    }
  }

  async function register(e: FormEvent<HTMLFormElement>) {
    e.preventDefault()
    setAuthError('')
    const fd = new FormData(e.currentTarget)
    try {
      const data = await api<{ driver: Driver; token: string }>('/api/v1/auth/driver/register', {
        method: 'POST',
        body: {
          email: fd.get('email'),
          password: fd.get('password'),
          phone: fd.get('phone') || '',
        },
      })
      localStorage.setItem(TOKEN_KEY, data.token)
      setToken(data.token)
      setDriver(data.driver)
    } catch (err) {
      setAuthError(err instanceof Error ? err.message : 'Ошибка')
    }
  }

  function logout() {
    localStorage.removeItem(TOKEN_KEY)
    setToken('')
    setDriver(null)
  }

  async function copyHub() {
    if (!driver) return
    try {
      await navigator.clipboard.writeText(driver.hub_id)
      setCopied(true)
      setTimeout(() => setCopied(false), 1500)
    } catch {
      /* ignore */
    }
  }

  if (!driver) {
    return (
      <Shell role="Водитель" centered>
        <div className="auth-panel">
          <div className="auth-brand">
            <BrandMark size={40} />
          </div>
          <Card title="Вход водителя">
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
                  <input name="email" type="email" defaultValue="driver@demo.ru" autoComplete="username" required />
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
                <Field label="Email">
                  <input name="email" type="email" required />
                </Field>
                <Field label="Пароль">
                  <input name="password" type="password" required minLength={6} />
                </Field>
                <Field label="Телефон">
                  <input name="phone" inputMode="tel" />
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

  const fio = [driver.last_name, driver.first_name, driver.middle_name].filter(Boolean).join(' ') || 'Профиль не заполнен'
  const needsVerify = driver.verification_status !== 'verified'

  return (
    <Shell role="Водитель" onLogout={logout}>
      <section className="id-card">
        <div className="id-card-main">
          <div className="id-card-top">
            <div>
              <p className="id-kicker">Идентификатор водителя</p>
              <h1 className="id-name">{fio}</h1>
            </div>
            <Badge status={driver.verification_status} />
          </div>

          <div className="hub-row">
            <code className="hub-code">{driver.hub_id}</code>
            <button type="button" className="btn" onClick={copyHub}>
              {copied ? 'Скопировано' : 'Копировать'}
            </button>
          </div>

          <Facts
            rows={[
              ['Email', driver.email || '—'],
              ['Телефон', driver.phone || '—'],
              ['Категории', (driver.categories || []).join(', ') || '—'],
              ['ВУ', driver.license ? `${driver.license.series} ${driver.license.number}` : '—'],
            ]}
          />

          {needsVerify ? (
            <div className="verify-callout">
              <div>
                <strong>Профиль не подтверждён</strong>
                <p>Загрузите фото ВУ или подтвердите через Госуслуги</p>
              </div>
              <button type="button" className="btn btn-primary" onClick={() => navigate('/driver/verify')}>
                Верификация
              </button>
            </div>
          ) : (
            <div className="verify-done">
              <span>Личность подтверждена</span>
              <Link className="nav-link" to="/driver/verify">
                Обновить данные ВУ
              </Link>
            </div>
          )}
        </div>

        <div className="id-card-qr">
          <QRCodeSVG value={qrValue} size={168} level="M" includeMargin={false} bgColor="#ffffff" fgColor="#0b1f2a" />
          <p>Покажите QR компании для идентификации</p>
        </div>
      </section>
    </Shell>
  )
}
