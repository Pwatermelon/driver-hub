import { Link } from 'react-router-dom'
import { BrandLink, BrandMark } from '../components/Brand'

export function HomePage() {
  return (
    <div className="home">
      <header className="topbar topbar-home">
        <BrandLink />
      </header>

      <section className="home-hero">
        <div className="home-hero-bg" aria-hidden />
        <div className="home-hero-inner">
          <div className="home-brand-block animate-in">
            <BrandMark size={56} className="home-mark" />
            <p className="home-brand-name">
              Driver <em>Hub</em>
            </p>
          </div>
          <h1 className="animate-in delay-1">Единый реестр водителей</h1>
          <p className="home-lead animate-in delay-2">
            Проверка досье и обмен данными между логистическими компаниями
          </p>
          <div className="home-cta animate-in delay-3">
            <Link className="btn btn-primary btn-lg" to="/driver">
              Войти как водитель
            </Link>
            <Link className="btn btn-ghost btn-lg" to="/company" target="_blank" rel="noopener">
              Войти как компания
            </Link>
          </div>
        </div>
      </section>

      <section className="home-demo-wrap">
        <h2 className="home-demo-title">Демо-доступ</h2>
        <div className="home-demo">
          <div>
            <span>Водитель</span>
            <code>driver@demo.ru</code>
            <code>demo1234</code>
          </div>
          <div>
            <span>Компания</span>
            <code>hr@logplus.ru</code>
            <code>demo1234</code>
          </div>
          <div>
            <span>Hub ID</span>
            <code>DH-DEMOTST2</code>
          </div>
        </div>
      </section>
    </div>
  )
}
