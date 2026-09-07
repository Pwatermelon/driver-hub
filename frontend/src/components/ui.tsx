import { Fragment, type ReactNode } from 'react'
import { Link } from 'react-router-dom'
import { BrandLink } from './Brand'

export function Shell({
  role,
  children,
  onLogout,
  centered,
}: {
  role: 'Водитель' | 'Компания'
  children: ReactNode
  onLogout?: () => void
  centered?: boolean
}) {
  return (
    <div className={centered ? 'shell shell-auth' : 'shell'}>
      <header className="topbar">
        <BrandLink />
        <div className="topbar-right">
          <span className="role-chip">{role}</span>
          {role === 'Водитель' ? (
            <Link className="nav-link" to="/company" target="_blank" rel="noopener">
              Компания
            </Link>
          ) : (
            <Link className="nav-link" to="/driver" target="_blank" rel="noopener">
              Водитель
            </Link>
          )}
          {onLogout && (
            <button type="button" className="text-btn" onClick={onLogout}>
              Выйти
            </button>
          )}
        </div>
      </header>
      <main className={centered ? 'auth-stage' : 'page'}>{children}</main>
    </div>
  )
}

export function Card({
  title,
  badge,
  children,
}: {
  title?: string
  badge?: ReactNode
  children: ReactNode
}) {
  return (
    <section className="card">
      {(title || badge) && (
        <div className="card-head">
          {title ? <h1>{title}</h1> : <span />}
          {badge}
        </div>
      )}
      {children}
    </section>
  )
}

export function Badge({ status }: { status?: string }) {
  const ok = status === 'verified'
  const map: Record<string, string> = {
    none: 'не подтверждён',
    pending: 'на проверке',
    verified: 'подтверждён',
    rejected: 'отклонён',
    expired: 'истёк',
  }
  return <span className={`badge ${ok ? 'badge-ok' : 'badge-warn'}`}>{map[status || ''] || status || '—'}</span>
}

export function Alert({ kind, children }: { kind: 'error' | 'ok'; children?: string }) {
  if (!children) return null
  return <div className={`alert alert-${kind}`}>{children}</div>
}

export function Field({ label, children, className = '' }: { label: string; children: ReactNode; className?: string }) {
  return (
    <label className={className}>
      {label}
      {children}
    </label>
  )
}

export function Facts({ rows }: { rows: Array<[string, ReactNode]> }) {
  return (
    <dl className="kv">
      {rows.map(([k, v]) => (
        <Fragment key={k}>
          <dt>{k}</dt>
          <dd>{v}</dd>
        </Fragment>
      ))}
    </dl>
  )
}
