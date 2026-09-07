.PHONY: up down logs build frontend

up:
	docker compose up --build -d

down:
	docker compose down

logs:
	docker compose logs -f api web

build:
	docker compose build

frontend:
	cd frontend && npm install && npm run dev
