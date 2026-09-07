import { useEffect, useMemo, useState, type FormEvent } from 'react'
import { Link, Navigate } from 'react-router-dom'
import { api, apiForm } from '../api/client'
import type { Driver, LicenseDraft } from '../api/types'
import { Alert, Card, Field, Shell } from '../components/ui'

const TOKEN_KEY = 'dh_driver_token'

export function DriverVerifyPage() {
  const [token] = useState(() => localStorage.getItem(TOKEN_KEY) || '')
  const [driver, setDriver] = useState<Driver | null>(null)
  const [loading, setLoading] = useState(true)
  const [front, setFront] = useState<File | null>(null)
  const [back, setBack] = useState<File | null>(null)
  const [scanning, setScanning] = useState(false)
  const [draft, setDraft] = useState<LicenseDraft | null>(null)
  const [form, setForm] = useState<LicenseDraft | null>(null)
  const [msg, setMsg] = useState<{ kind: 'error' | 'ok'; text: string } | null>(null)

  const frontUrl = useMemo(() => (front ? URL.createObjectURL(front) : ''), [front])
  const backUrl = useMemo(() => (back ? URL.createObjectURL(back) : ''), [back])

  useEffect(() => {
    if (!token) {
      setLoading(false)
      return
    }
    api<Driver>('/api/v1/me/driver', { token })
      .then(setDriver)
      .catch(() => setDriver(null))
      .finally(() => setLoading(false))
  }, [token])

  if (!loading && (!token || !driver)) {
    return <Navigate to="/driver" replace />
  }

  async function esia() {
    setMsg(null)
    try {
      const data = await api<{ driver: Driver; token: string }>('/api/v1/auth/esia/mock-verify', {
        method: 'POST',
        token,
        body: { preset: 'ivanov' },
      })
      localStorage.setItem(TOKEN_KEY, data.token)
      setDriver(data.driver)
      setMsg({ kind: 'ok', text: 'Данные из Госуслуг сохранены' })
    } catch (err) {
      setMsg({ kind: 'error', text: err instanceof Error ? err.message : 'Ошибка' })
    }
  }

  async function scan(e: FormEvent) {
    e.preventDefault()
    setMsg(null)
    if (!front || !back) {
      setMsg({ kind: 'error', text: 'Загрузите обе стороны ВУ' })
      return
    }
    setScanning(true)
    try {
      const body = new FormData()
      body.append('front', front)
      body.append('back', back)
      const data = await apiForm<LicenseDraft>('/api/v1/me/license/scan', token, body)
      setDraft(data)
      setForm({ ...data })
      setMsg({ kind: 'ok', text: 'Проверьте поля и сохраните' })
    } catch (err) {
      setMsg({ kind: 'error', text: err instanceof Error ? err.message : 'Ошибка' })
    } finally {
      setScanning(false)
    }
  }

  async function confirmLicense(e: FormEvent) {
    e.preventDefault()
    if (!form) return
    setMsg(null)
    try {
      const d = await api<Driver>('/api/v1/me/license/confirm', {
        method: 'POST',
        token,
        body: {
          ...form,
          categories: form.categories,
          source: draft?.source || 'manual',
          confidence: draft?.confidence || 1,
        },
      })
      setDriver(d)
      setMsg({ kind: 'ok', text: 'ВУ сохранено' })
    } catch (err) {
      setMsg({ kind: 'error', text: err instanceof Error ? err.message : 'Ошибка' })
    }
  }

  function logout() {
    localStorage.removeItem(TOKEN_KEY)
    window.location.href = '/driver'
  }

  if (loading || !driver) {
    return (
      <Shell role="Водитель">
        <p className="sub">Загрузка…</p>
      </Shell>
    )
  }

  return (
    <Shell role="Водитель" onLogout={logout}>
      <div className="page-head">
        <Link className="nav-link" to="/driver">
          ← Кабинет
        </Link>
        <h1 className="page-title">Верификация</h1>
      </div>

      <Card title="Госуслуги">
        <p className="sub" style={{ marginBottom: 14 }}>
          Подтянуть ФИО и данные ВУ из подтверждённой учётной записи
        </p>
        <button type="button" className="btn btn-primary" onClick={esia}>
          Подтвердить через Госуслуги
        </button>
      </Card>

      <section className="card">
        <div className="card-head">
          <h2>Фото водительского удостоверения</h2>
        </div>
        <form className="form" onSubmit={scan}>
          <div className="grid-2">
            <Field label="Лицевая сторона">
              <input type="file" accept="image/*" capture="environment" required onChange={(e) => setFront(e.target.files?.[0] || null)} />
            </Field>
            <Field label="Оборотная сторона">
              <input type="file" accept="image/*" capture="environment" required onChange={(e) => setBack(e.target.files?.[0] || null)} />
            </Field>
          </div>
          <div className="preview-row">
            {frontUrl && <img className="preview" src={frontUrl} alt="" />}
            {backUrl && <img className="preview" src={backUrl} alt="" />}
          </div>
          <button className="btn btn-primary" type="submit" disabled={scanning}>
            {scanning ? 'Распознаём…' : 'Распознать ВУ'}
          </button>
        </form>

        {form && (
          <form className="form" style={{ marginTop: 16 }} onSubmit={confirmLicense}>
            <div className="grid-2">
              <Field label="Фамилия">
                <input value={form.last_name} onChange={(e) => setForm({ ...form, last_name: e.target.value })} />
              </Field>
              <Field label="Имя">
                <input value={form.first_name} onChange={(e) => setForm({ ...form, first_name: e.target.value })} />
              </Field>
            </div>
            <Field label="Отчество">
              <input value={form.middle_name} onChange={(e) => setForm({ ...form, middle_name: e.target.value })} />
            </Field>
            <div className="grid-2">
              <Field label="Дата рождения">
                <input type="date" value={form.birth_date} onChange={(e) => setForm({ ...form, birth_date: e.target.value })} />
              </Field>
              <Field label="Категории">
                <input
                  value={(form.categories || []).join(', ')}
                  onChange={(e) =>
                    setForm({
                      ...form,
                      categories: e.target.value
                        .split(',')
                        .map((s) => s.trim())
                        .filter(Boolean),
                    })
                  }
                />
              </Field>
            </div>
            <div className="grid-2">
              <Field label="Серия">
                <input required value={form.series} onChange={(e) => setForm({ ...form, series: e.target.value })} />
              </Field>
              <Field label="Номер">
                <input required value={form.number} onChange={(e) => setForm({ ...form, number: e.target.value })} />
              </Field>
            </div>
            <div className="grid-2">
              <Field label="Дата выдачи">
                <input type="date" value={form.issue_date} onChange={(e) => setForm({ ...form, issue_date: e.target.value })} />
              </Field>
              <Field label="Действует до">
                <input type="date" value={form.expiry_date} onChange={(e) => setForm({ ...form, expiry_date: e.target.value })} />
              </Field>
            </div>
            <Field label="Кем выдано">
              <input value={form.issuer} onChange={(e) => setForm({ ...form, issuer: e.target.value })} />
            </Field>
            <div className="actions">
              <button className="btn btn-primary" type="submit">
                Сохранить
              </button>
              <button
                className="btn"
                type="button"
                onClick={() => {
                  setDraft(null)
                  setForm(null)
                  setFront(null)
                  setBack(null)
                }}
              >
                Сбросить
              </button>
            </div>
          </form>
        )}
      </section>

      {msg && <Alert kind={msg.kind}>{msg.text}</Alert>}
    </Shell>
  )
}
