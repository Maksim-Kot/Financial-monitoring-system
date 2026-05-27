build:
	@go build -o ./bin/bot.exe ./cmd/bot

run: build
	@./bin/bot.exe

infra/up:
	@docker compose -f ci/docker-compose.yml -p fms-infra up -d postgres

infra/down:
	@docker compose -f ci/docker-compose.yml -p fms-infra down

infra/remove:
	@docker compose -f ci/docker-compose.yml -p fms-infra down -v

lint:
	@golangci-lint run ./...