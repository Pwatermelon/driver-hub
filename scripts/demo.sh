#!/usr/bin/env bash
# Демо-сценарий через API (сервис должен быть на :8080)
set -euo pipefail
API=${API:-http://localhost:8080}

echo "== компания =="
CO=$(curl -s "$API/api/v1/auth/company/register" -H 'Content-Type: application/json' \
  -d '{"name":"ООО Логистика Плюс","inn":"7701234567","email":"hr@logplus.ru","password":"secret123"}')
echo "$CO" | head -c 400; echo
CO_TOKEN=$(echo "$CO" | sed -n 's/.*"token":"\([^"]*\)".*/\1/p')
CO_ID=$(echo "$CO" | sed -n 's/.*"id":"\([^"]*\)".*/\1/p' | head -1)

echo "== водитель =="
DR=$(curl -s "$API/api/v1/auth/driver/register" -H 'Content-Type: application/json' \
  -d '{"email":"driver@ex.ru","password":"secret123","first_name":"Иван","last_name":"Иванов","categories":["B","C"]}')
echo "$DR" | head -c 400; echo
DR_TOKEN=$(echo "$DR" | sed -n 's/.*"token":"\([^"]*\)".*/\1/p')
HUB=$(echo "$DR" | sed -n 's/.*"hub_id":"\([^"]*\)".*/\1/p')

echo "== ESIA mock verify =="
VER=$(curl -s "$API/api/v1/auth/esia/mock-verify" -H "Authorization: Bearer $DR_TOKEN" \
  -H 'Content-Type: application/json' -d '{"preset":"ivanov"}')
echo "$VER" | head -c 500; echo
DR_TOKEN=$(echo "$VER" | sed -n 's/.*"token":"\([^"]*\)".*/\1/p')
HUB=$(echo "$VER" | sed -n 's/.*"hub_id":"\([^"]*\)".*/\1/p')

echo "== grant =="
curl -s "$API/api/v1/me/grants" -H "Authorization: Bearer $DR_TOKEN" -H 'Content-Type: application/json' \
  -d "{\"company_id\":\"$CO_ID\",\"purpose\":\"employment\",\"days\":30}"; echo

echo "== complaint + blacklist + accident =="
curl -s "$API/api/v1/drivers/$HUB/complaints" -H "Authorization: Bearer $CO_TOKEN" -H 'Content-Type: application/json' \
  -d '{"category":"safety","text":"Нарушение режима труда","severity":"medium"}'; echo
curl -s "$API/api/v1/drivers/$HUB/blacklist" -H "Authorization: Bearer $CO_TOKEN" -H 'Content-Type: application/json' \
  -d '{"reason":"Систематические опоздания"}'; echo
curl -s "$API/api/v1/drivers/$HUB/accidents" -H "Authorization: Bearer $CO_TOKEN" -H 'Content-Type: application/json' \
  -d '{"description":"Легкое ДТП","fault":"mutual","damage_level":"minor","location":"МО"}'; echo

echo "== dossier =="
curl -s "$API/api/v1/drivers/$HUB/dossier" -H "Authorization: Bearer $CO_TOKEN"; echo
echo
echo "Hub ID: $HUB"
