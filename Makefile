.PHONY: dev build start test migrate rollback status docker-build docker-run docker-stop

# ── Dev ──────────────────────────────────────────────────────────────────────
dev:
	go run cmd/api/main.go

build:
	CGO_ENABLED=0 go build -o bin/server ./cmd/api

start: build
	./bin/server

test:
	go test ./... -v -race

# ── Database ─────────────────────────────────────────────────────────────────
migrate:
	@export $$(cat .env | xargs) && \
	goose -dir migrations postgres "$$DATABASE_URL" up

rollback:
	@export $$(cat .env | xargs) && \
	goose -dir migrations postgres "$$DATABASE_URL" down

status:
	@export $$(cat .env | xargs) && \
	goose -dir migrations postgres "$$DATABASE_URL" status

# ── Docker ───────────────────────────────────────────────────────────────────
docker-build:
	docker build -t vlab-api .

docker-run:
	@export $$(cat .env | xargs) && \
	docker run -p 8089:8089 \
		-e DATABASE_URL="$$DATABASE_URL" \
		-e JWT_SECRET="$$JWT_SECRET" \
		-e PORT=8089 \
		vlab-api

docker-stop:
	docker stop $$(docker ps -q --filter ancestor=vlab-api)
