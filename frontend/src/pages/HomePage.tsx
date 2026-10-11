import { Link } from 'react-router-dom'
import { BrandLink, BrandMark } from '../components/Brand'

const drivers = [
  { hub: 'DH-DEMOTST2', email: 'driver@demo.ru', note: 'сильный профиль' },
  { hub: 'DH-PETROV01', email: 'petrov@demo.ru', note: 'чёрный список' },
  { hub: 'DH-SIDOROV2', email: 'sidorov@demo.ru', note: 'ДТП и штрафы' },
  { hub: 'DH-KOZLOVA3', email: 'kozlova@demo.ru', note: 'чистый профиль' },
  { hub: 'DH-NOVIKOV4', email: 'novikov@demo.ru', note: 'без верификации' },
  { hub: 'DH-ORLOV05', email: 'orlov@demo.ru', note: 'лишение ВУ' },
]

const companies = [
  { email: 'hr@logplus.ru', name: 'Логистика Плюс' },
  { email: 'hr@fasttruck.ru', name: 'ФастТрак' },
  { email: 'fleet@severtrans.ru', name: 'СеверТранс' },
]

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
            <Link className="btn btn-ghost btn-lg" to="/company">
              Войти как компания
            </Link>
          </div>
        </div>
      </section>

      <section className="home-features">
        <div className="home-feature">
          <strong>Hub ID + QR</strong>
          <span>Водитель идентифицируется за секунды — без бумажных копий ВУ</span>
        </div>
        <div className="home-feature">
          <strong>Досье для HR</strong>
          <span>Рекомендации, жалобы, ДТП, штрафы и чёрный список в одном месте</span>
        </div>
        <div className="home-feature">
          <strong>Верификация</strong>
          <span>Госуслуги и распознавание водительского удостоверения по фото</span>
        </div>
      </section>

      <section className="home-demo-wrap">
        <h2 className="home-demo-title">Демо — пароль везде demo1234</h2>
        <div className="home-demo-grid">
          <div className="home-demo-card">
            <h3>Компании</h3>
            <ul className="home-demo-list">
              {companies.map((c) => (
                <li key={c.email}>
                  <strong>{c.name}</strong>
                  <code>{c.email}</code>
                </li>
              ))}
            </ul>
          </div>
          <div className="home-demo-card home-demo-card-wide">
            <h3>Водители (Hub ID для поиска)</h3>
            <ul className="home-demo-list home-demo-list-drivers">
              {drivers.map((d) => (
                <li key={d.hub}>
                  <code className="hub-pill">{d.hub}</code>
                  <span>
                    <code>{d.email}</code>
                    <em>{d.note}</em>
                  </span>
                </li>
              ))}
            </ul>
          </div>
        </div>
      </section>
    </div>
  )
}
