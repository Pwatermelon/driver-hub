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
- API напрямую: http://localhost:8080/healthz

### Демо

| Роль | Логин | Пароль |
|------|-------|--------|
| Водитель | `driver@demo.ru` | `demo1234` |
| Компания | `hr@logplus.ru` | `demo1234` |

Hub ID: `DH-DEMOTST2`

## Dev фронта

```bash
docker compose up -d db api
cd frontend && npm run dev
```

Vite на `:5173`, API через proxy на `:8080`.

## Деплой

`var 1.0.0` → GHCR (`*-api` + `*-web`) → `hackton-test.ru`  
См. [deploy/README.md](deploy/README.md)
