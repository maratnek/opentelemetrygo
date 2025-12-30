# .PHONY: proto build run test clean

proto_gen:
	protoc \
	--go_out=. \
	--go-grpc_out=. \
	--go_opt=paths=source_relative \
	--go-grpc_opt=paths=source_relative \
	proto/trans.proto

# Папка для бинарников
BIN_DIR = bin
SERVER_BIN = $(BIN_DIR)/server

# Генерация proto (если изменился .proto)
proto:
	cd proto && protoc --go_out=. --go-grpc_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_opt=paths=source_relative \
		trans.proto

# Сборка сервера
build: proto
	@echo "Building server to $(SERVER_BIN)..."
	mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 go build -o $(SERVER_BIN) ./server

# Локальный запуск сервера
run:
	go run ./server

logs:
	docker logs -f server-otel

# Запуск всего стека (сервер + Jaeger)
up:
	docker compose up -d

# Остановка стека
down:
	docker compose down

# Запуск интеграционных тестов (сервер должен быть запущен через compose-up)
test:
	go test -v ./client_test

watch:
	watch 'docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"'

# Очистка
clean:
	docker compose down -v
	go clean