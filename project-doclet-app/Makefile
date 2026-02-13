.PHONY: up down logs run-document run-collab run-frontend

up:
	docker compose up -d

down:
	docker compose down -v

logs:
	docker compose logs -f

run-document:
	DOCLET_DATABASE_URL=$${DOCLET_DATABASE_URL:-"postgres://doclet:doclet@localhost:5432/doclet?sslmode=disable"} \
	DOCLET_NATS_URL=$${DOCLET_NATS_URL:-"nats://localhost:4222"} \
	go run -C service-go-document ./cmd

run-collab:
	DOCLET_NATS_URL=$${DOCLET_NATS_URL:-"nats://localhost:4222"} \
	go run -C service-go-collab ./cmd

run-frontend:
	cd webapp-react-frontend && npm install && npm run dev
