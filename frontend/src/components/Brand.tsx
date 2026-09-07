import { Link } from 'react-router-dom'

/** Знак: hub + полоса дороги — монохромный с янтарным акцентом */
export function BrandMark({ size = 28, className = '' }: { size?: number; className?: string }) {
  return (
    <svg
      className={`brand-mark ${className}`}
      width={size}
      height={size}
      viewBox="0 0 32 32"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      aria-hidden
    >
      <rect width="32" height="32" rx="8" fill="currentColor" className="brand-mark-bg" />
      <path
        d="M8 11.5h16M8 16h16M8 20.5h16"
        stroke="#F5B942"
        strokeWidth="1.6"
        strokeLinecap="round"
        opacity="0.95"
      />
      <circle cx="16" cy="16" r="5.2" fill="#0B1F2A" stroke="#F5B942" strokeWidth="1.5" />
      <circle cx="16" cy="16" r="2" fill="#F5B942" />
    </svg>
  )
}

export function BrandLink({ large = false }: { large?: boolean }) {
  return (
    <Link className={large ? 'brand brand-lg' : 'brand'} to="/" aria-label="Driver Hub">
      <BrandMark size={large ? 40 : 28} />
      <span className="brand-word">
        Driver&nbsp;<em>Hub</em>
      </span>
    </Link>
  )
}
