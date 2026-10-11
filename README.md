# Driver Hub

Реестр водителей для логистики.

- Backend: Go + PostgreSQL (`api`)
- Frontend: React + Vite + TypeScript (`web` / nginx)
- Раздельные контейнеры; nginx проксирует `/api` на API

## Запуск

```bash
docker compose up --build
```

Открывай: **http://localhost/**

- Водитель: http://localhost/driver
- Компания: http://localhost/company

Пароль для всех демо-аккаунтов: **`demo1234`**

### Компании

| Компания | Email |
|----------|-------|
| ООО Логистика Плюс | `hr@logplus.ru` |
| ООО ФастТрак Логистик | `hr@fasttruck.ru` |
| АО СеверТранс | `fleet@severtrans.ru` |

### Водители

| Hub ID | Email | Что смотреть в досье |
|--------|-------|----------------------|
| `DH-DEMOTST2` | `driver@demo.ru` | Сильный профиль, рекомендации |
| `DH-PETROV01` | `petrov@demo.ru` | Чёрный список, жалобы |
| `DH-SIDOROV2` | `sidorov@demo.ru` | ДТП и штрафы |
| `DH-KOZLOVA3` | `kozlova@demo.ru` | Чистый профиль |
| `DH-NOVIKOV4` | `novikov@demo.ru` | Без верификации |
| `DH-ORLOV05` | `orlov@demo.ru` | Лишение ВУ, блэклист |

## Dev фронта

```bash
docker compose up -d db api
cd frontend && npm run dev
```

## Деплой

`var 1.0.0` → GHCR (`*-api` + `*-web`) → сервер  
См. [deploy/README.md](deploy/README.md)
