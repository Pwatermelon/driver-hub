# Deploy Driver Hub via GHCR + GitHub Actions

Как у `mail-eco-product`: коммит с версией → сборка образа → push в GHCR → SSH на сервер → `docker compose up`.

## Триггер

В сообщении коммита (или PR title / workflow_dispatch):

```
var 1.0.0
```

Также принимаются: `ver 1.0.0`, `version 1.0.0`, `v1.0.0`.

Без версии workflow **не деплоит** (только skip).

## Secrets (Environment `production`)

| Secret | Назначение |
|--------|------------|
| `DOMAIN` | `hackton-test.ru` (как у mail-eco-product) |
| `ACME_EMAIL` | `admin@hackton-test.ru` |
| `DEPLOY_HOST` | IP/хост сервера |
| `DEPLOY_USER` | SSH user |
| `SSH_PRIVATE_KEY` | приватный SSH-ключ |
| `JWT_SECRET` | секрет JWT |
| `POSTGRES_PASSWORD` | опционально; иначе `driverhub` (не используй JWT как пароль БД) |
| `DEPLOY_PATH` | опционально, по умолчанию `/opt/driver-hub` |
| `SSH_PORT` | опционально, `22` |
| `GHCR_PULL_TOKEN` | PAT `read:packages`, если образ private |

## Если на сайте «Load failed» / API 502

На сервере:

```bash
cd /opt/driver-hub
docker compose ps -a
docker compose logs --tail=100 api
```

Частая причина: `api` в restart из‑за битого `DATABASE_URL` (раньше пароль БД = JWT со спецсимволами).

Починка вручную:

```bash
cd /opt/driver-hub
# в .env должно быть:
# POSTGRES_PASSWORD=driverhub
# DATABASE_URL=postgres://driverhub:driverhub@db:5432/driverhub?sslmode=disable
# IMAGE_API=... IMAGE_WEB=...

docker compose up -d --force-recreate api
curl -sS http://127.0.0.1/healthz
```

Если Postgres уже инициализирован другим паролем и api пишет `password authentication failed` — либо верни старый пароль в `.env`, либо (с потерей данных демо):

```bash
docker compose down
docker volume rm driver-hub_pgdata   # имя уточни: docker volume ls | grep pg
docker compose up -d
```

Деплой только с версией в коммите: `var 0.1.1` (иначе Actions **не выкатывает**).

## На сервере после деплоя

```
https://hackton-test.ru/
https://hackton-test.ru/driver
https://hackton-test.ru/company
```

> Если mail-eco уже занимает `:80/:443` на том же хосте — либо останови его compose в `/opt/mail-sdelki`, либо вынеси Driver Hub на поддомен и поменяй `DOMAIN`.

Демо:

- водитель: `driver@demo.ru` / `demo1234` (Hub `DH-DEMOTST2`)
- компания: `hr@logplus.ru` / `demo1234`

Образ: `ghcr.io/<owner>/driver-hub-api:<version>` + `ghcr.io/<owner>/driver-hub-web:<version>`

