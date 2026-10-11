import type { FormEvent, ReactNode } from 'react'
import { Alert, Card, Field } from './ui'
import { BrandMark } from './Brand'

export function AuthPanel({
  title,
  tab,
  onTab,
  error,
  loading,
  onLogin,
  onRegister,
  loginDefaults,
  registerExtra,
}: {
  title: string
  tab: 'login' | 'reg'
  onTab: (t: 'login' | 'reg') => void
  error?: string
  loading?: boolean
  onLogin: (e: FormEvent<HTMLFormElement>) => void
  onRegister: (e: FormEvent<HTMLFormElement>) => void
  loginDefaults?: { email?: string; password?: string }
  registerExtra?: ReactNode
}) {
  return (
    <div className="auth-panel">
      <div className="auth-brand">
        <BrandMark size={40} />
      </div>
      <Card title={title}>
        <div className="tabs">
          <button type="button" className={tab === 'login' ? 'active' : ''} onClick={() => onTab('login')} disabled={loading}>
            Вход
          </button>
          <button type="button" className={tab === 'reg' ? 'active' : ''} onClick={() => onTab('reg')} disabled={loading}>
            Регистрация
          </button>
        </div>

        {tab === 'login' ? (
          <form className="form" onSubmit={onLogin}>
            <Field label="Email">
              <input
                name="email"
                type="email"
                defaultValue={loginDefaults?.email}
                autoComplete="username"
                required
                disabled={loading}
              />
            </Field>
            <Field label="Пароль">
              <input
                name="password"
                type="password"
                defaultValue={loginDefaults?.password}
                autoComplete="current-password"
                required
                disabled={loading}
              />
            </Field>
            <button className="btn btn-primary btn-block" type="submit" disabled={loading}>
              {loading ? 'Входим…' : 'Войти'}
            </button>
          </form>
        ) : (
          <form className="form" onSubmit={onRegister}>
            {registerExtra}
            <Field label="Email">
              <input name="email" type="email" required disabled={loading} autoComplete="email" />
            </Field>
            <Field label="Пароль">
              <input name="password" type="password" required minLength={6} disabled={loading} autoComplete="new-password" />
            </Field>
            <button className="btn btn-primary btn-block" type="submit" disabled={loading}>
              {loading ? 'Создаём…' : 'Создать аккаунт'}
            </button>
          </form>
        )}
        <Alert kind="error">{error}</Alert>
      </Card>
    </div>
  )
}
