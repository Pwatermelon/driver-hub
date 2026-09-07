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
| `JWT_SECRET` | секрет JWT (+ пароль Postgres в проде) |
| `DEPLOY_PATH` | опционально, по умолчанию `/opt/driver-hub` |
| `SSH_PORT` | опционально, `22` |
| `GHCR_PULL_TOKEN` | PAT `read:packages`, если образ private |

После первого пуша: GitHub → Packages → пакет → Visibility → Public  
(или оставь private и задай `GHCR_PULL_TOKEN`).

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

