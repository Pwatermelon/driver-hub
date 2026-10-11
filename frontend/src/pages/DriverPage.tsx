import { useEffect, useMemo, useState, type FormEvent } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { QRCodeSVG } from 'qrcode.react'
import { api, errMessage } from '../api/client'
import type { Driver } from '../api/types'
import { AuthPanel } from '../components/AuthPanel'
import { Badge, Facts, Field, Shell } from '../components/ui'

const TOKEN_KEY = 'dh_driver_token'

export function DriverPage() {
  const [params, setParams] = useSearchParams()
  const navigate = useNavigate()
  const [token, setToken] = useState(() => localStorage.getItem(TOKEN_KEY) || '')
  const [driver, setDriver] = useState<Driver | null>(null)
  const [booting, setBooting] = useState(() => !!localStorage.getItem(TOKEN_KEY))
  const [tab, setTab] = useState<'login' | 'reg'>('login')
  const [authError, setAuthError] = useState('')
  const [busy, setBusy] = useState(false)
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
    if (!token) {
      setBooting(false)
      return
    }
    let cancelled = false
    setBooting(true)
    api<Driver>('/api/v1/me/driver', { token })
      .then((d) => {
        if (!cancelled) setDriver(d)
      })
      .catch(() => {
        if (!cancelled) {
          localStorage.removeItem(TOKEN_KEY)
          setToken('')
          setDriver(null)
        }
      })
      .finally(() => {
        if (!cancelled) setBooting(false)
      })
    return () => {
      cancelled = true
    }
  }, [token])

  const qrValue = useMemo(() => {
    if (!driver) return ''
    return `${window.location.origin}/company?hub=${encodeURIComponent(driver.hub_id)}`
  }, [driver])

  async function login(e: FormEvent<HTMLFormElement>) {
    e.preventDefault()
    setAuthError('')
    setBusy(true)
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
      setAuthError(errMessage(err))
    } finally {
      setBusy(false)
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

  if (booting) {
    return (
      <Shell role="Водитель" centered>
        <p className="boot-msg">Загрузка профиля…</p>
      </Shell>
    )
  }

  if (!driver) {
    return (
      <Shell role="Водитель" centered>
        <AuthPanel
          title="Кабинет водителя"
          tab={tab}
          onTab={(t) => {
            setTab(t)
            setAuthError('')
          }}
          error={authError}
          loading={busy}
          onLogin={login}
          onRegister={register}
          loginDefaults={{ email: 'driver@demo.ru', password: 'demo1234' }}
          registerExtra={
            <Field label="Телефон">
              <input name="phone" inputMode="tel" disabled={busy} />
            </Field>
          }
        />
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
          <div className="qr-frame">
            <QRCodeSVG value={qrValue} size={168} level="M" includeMargin={false} bgColor="#ffffff" fgColor="#0b1f2a" />
          </div>
          <p>Покажите QR компании для идентификации</p>
        </div>
      </section>
    </Shell>
  )
}
