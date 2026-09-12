#!/usr/bin/env bash
# Диагностика и подъём Driver Hub на сервере.
# Запуск: cd /opt/driver-hub && bash fix-prod.sh
set -euo pipefail

cd "$(dirname "$0")"

echo "== compose ps =="
docker compose ps -a || true

echo
echo "== api logs (last 80) =="
docker compose logs --tail=80 api || true

echo
echo "== .env keys (без секретов) =="
if [[ -f .env ]]; then
  grep -E '^(IMAGE_|DOMAIN|PUBLIC_BASE_URL|DATABASE_URL)=' .env | sed -E 's/(PASSWORD|SECRET|URL)=.*/\1=***/' || true
  # покажем, есть ли DATABASE_URL
  if grep -q '^DATABASE_URL=' .env; then
    echo "DATABASE_URL: set"
  else
    echo "DATABASE_URL: MISSING — это ломает api"
  fi
else
  echo "нет файла .env"
fi

echo
echo "== recreate stack =="
docker compose pull || true
docker compose up -d --remove-orphans

echo
echo "== wait api =="
for i in $(seq 1 30); do
  if docker compose exec -T api wget -q -O - http://127.0.0.1:8080/healthz >/dev/null 2>&1; then
    echo "api ok"
    break
  fi
  echo "waiting api ($i/30)..."
  sleep 2
done

echo
echo "== local healthz via caddy :80 =="
curl -sS -m 5 http://127.0.0.1/healthz || echo "FAIL http healthz"

echo
echo "== public =="
curl -sS -m 5 -o /dev/null -w "https healthz: %{http_code}\n" "https://${DOMAIN:-hackton-test.ru}/healthz" || true

docker compose ps
