import { Fragment, type ReactNode } from 'react'
import { Link, useLocation } from 'react-router-dom'
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
  const loc = useLocation()
  const showDock = !!onLogout && !centered

  return (
    <div className={centered ? 'shell shell-auth' : showDock ? 'shell shell-docked' : 'shell'}>
      <header className="topbar">
        <BrandLink />
        <div className="topbar-right">
          <span className="role-chip">{role}</span>
          {role === 'Водитель' ? (
            <Link className="nav-link nav-link-alt" to="/company" target="_blank" rel="noopener">
              Компания
            </Link>
          ) : (
            <Link className="nav-link nav-link-alt" to="/driver" target="_blank" rel="noopener">
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

      {showDock && role === 'Водитель' && (
        <nav className="mobile-dock" aria-label="Навигация">
          <Link className={loc.pathname === '/driver' ? 'dock-item active' : 'dock-item'} to="/driver">
            <span className="dock-ico" aria-hidden>
              ◇
            </span>
            Кабинет
          </Link>
          <Link className={loc.pathname.startsWith('/driver/verify') ? 'dock-item active' : 'dock-item'} to="/driver/verify">
            <span className="dock-ico" aria-hidden>
              ✓
            </span>
            Верификация
          </Link>
          <button type="button" className="dock-item" onClick={onLogout}>
            <span className="dock-ico" aria-hidden>
              →
            </span>
            Выйти
          </button>
        </nav>
      )}

      {showDock && role === 'Компания' && (
        <nav className="mobile-dock" aria-label="Навигация">
          <Link className="dock-item active" to="/company">
            <span className="dock-ico" aria-hidden>
              ◇
            </span>
            Поиск
          </Link>
          <button type="button" className="dock-item" onClick={onLogout}>
            <span className="dock-ico" aria-hidden>
              →
            </span>
            Выйти
          </button>
        </nav>
      )}
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

export function FilePick({
  label,
  file,
  preview,
  onChange,
}: {
  label: string
  file: File | null
  preview: string
  onChange: (f: File | null) => void
}) {
  return (
    <label className="file-pick">
      <span className="file-pick-label">{label}</span>
      <span className={`file-pick-box ${file ? 'has-file' : ''}`}>
        {preview ? <img src={preview} alt="" /> : <span className="file-pick-hint">Нажмите, чтобы сделать фото или выбрать файл</span>}
      </span>
      <input
        type="file"
        accept="image/*"
        capture="environment"
        required={!file}
        onChange={(e) => onChange(e.target.files?.[0] || null)}
      />
      {file && <span className="file-pick-name">{file.name}</span>}
    </label>
  )
}
