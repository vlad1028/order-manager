APP_NAME = order-manager

SRC_DIR = .

OUT_PATH:=$(SRC_DIR)/pkg

LOCAL_BIN:=$(CURDIR)/bin

BIN = $(APP_NAME).exe

MIGRATIONS_DIR = $(SRC_DIR)/migrations

TEST_DB_DSN = "postgres://postgres:postgres@localhost:5431/postgres?sslmode=disable"

all: deps lint proto-generate build run

deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod download

build:
	go build -o $(BIN) -v $(SRC_DIR)/cmd/order-service/

run:
	./$(BIN)

clean:
	rm -f $(BIN)

install-linters:
	@echo "==> Installing gocyclo, gocognit, squawk and golangci-lint..."
	go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
	go install github.com/uudashr/gocognit/cmd/gocognit@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	pip install squawk-cli

lint-migrations:
	@echo "Linting SQL migrations..."
	squawk $(MIGRATIONS_DIR)/*.sql

lint: lint-migrations
	golangci-lint run $(SRC_DIR)/...

test: lint goose-up
	go test ./... -v

cover: lint goose-up
	go test ./... -coverprofile=cover.out
	go tool cover -html=cover.out -o cover.html

# ---------------------------
# Запуск базы данных в Docker
# ---------------------------

compose-up:
	docker-compose -f docker-compose.yml up -d

compose-down:
	docker-compose -f docker-compose.yml down

compose-ps:
	docker-compose -f docker-compose.yml ps

# ---------------------------
# Запуск миграций через Goose
# ---------------------------

goose-install:
	go install github.com/pressly/goose/v3/cmd/goose@latest

goose-add:
	goose -dir $(MIGRATIONS_DIR) postgres $(TEST_DB_DSN) create rename_me sql

goose-up: lint-migrations
	goose -dir $(MIGRATIONS_DIR) postgres $(TEST_DB_DSN) up

goose-status:
	goose -dir $(MIGRATIONS_DIR) postgres $(TEST_DB_DSN) status

.PHONY: all deps update build run clean install-linters lint test cover depgraph-install depgraph-build depgraph help lint-migrations

# ---------------------------------
# Запуск кодогенерации через protoc
# ---------------------------------

proto-bin-deps: .vendor-proto
	GOBIN=$(LOCAL_BIN) go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	GOBIN=$(LOCAL_BIN) go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	GOBIN=$(LOCAL_BIN) go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	GOBIN=$(LOCAL_BIN) go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
	GOBIN=$(LOCAL_BIN) go install github.com/envoyproxy/protoc-gen-validate@latest
	GOBIN=$(LOCAL_BIN) go install github.com/rakyll/statik@latest

proto-generate:
	mkdir -p ${OUT_PATH}
	protoc --proto_path api --proto_path vendor.protogen \
		--plugin=protoc-gen-go=$(LOCAL_BIN)/protoc-gen-go.exe --go_out=${OUT_PATH} --go_opt=paths=source_relative \
		--plugin=protoc-gen-go-grpc=$(LOCAL_BIN)/protoc-gen-go-grpc.exe --go-grpc_out=${OUT_PATH} --go-grpc_opt=paths=source_relative \
		--plugin=protoc-gen-grpc-gateway=$(LOCAL_BIN)/protoc-gen-grpc-gateway.exe --grpc-gateway_out ${OUT_PATH} --grpc-gateway_opt paths=source_relative \
		--plugin=protoc-gen-openapiv2=$(LOCAL_BIN)/protoc-gen-openapiv2.exe --openapiv2_out=${OUT_PATH} \
		--plugin=protoc-gen-validate=$(LOCAL_BIN)/protoc-gen-validate.exe --validate_out="lang=go,paths=source_relative:${OUT_PATH}" \
		./api/order-service/v1/order_service.proto

.vendor-proto: .vendor-proto/google/protobuf .vendor-proto/google/api .vendor-proto/protoc-gen-openapiv2/options .vendor-proto/validate

.vendor-proto/protoc-gen-openapiv2/options:
	git clone -b main --single-branch -n --depth=1 --filter=tree:0 \
 		https://github.com/grpc-ecosystem/grpc-gateway vendor.protogen/grpc-ecosystem && \
 		cd vendor.protogen/grpc-ecosystem && \
		git sparse-checkout set --no-cone protoc-gen-openapiv2/options && \
		git checkout
		mkdir -p vendor.protogen/protoc-gen-openapiv2
		mv vendor.protogen/grpc-ecosystem/protoc-gen-openapiv2/options vendor.protogen/protoc-gen-openapiv2
		rm -rf vendor.protogen/grpc-ecosystem

.vendor-proto/google/protobuf:
	git clone -b main --single-branch -n --depth=1 --filter=tree:0 \
		https://github.com/protocolbuffers/protobuf vendor.protogen/protobuf &&\
		cd vendor.protogen/protobuf &&\
		git sparse-checkout set --no-cone src/google/protobuf &&\
		git checkout
		mkdir -p vendor.protogen/google
		mv vendor.protogen/protobuf/src/google/protobuf vendor.protogen/google
		rm -rf vendor.protogen/protobuf

.vendor-proto/google/api:
	git clone -b master --single-branch -n --depth=1 --filter=tree:0 \
 		https://github.com/googleapis/googleapis vendor.protogen/googleapis && \
 		cd vendor.protogen/googleapis && \
		git sparse-checkout set --no-cone google/api && \
		git checkout
		mkdir -p  vendor.protogen/google
		mv vendor.protogen/googleapis/google/api vendor.protogen/google
		rm -rf vendor.protogen/googleapis

.vendor-proto/validate:
	git clone -b main --single-branch --depth=2 --filter=tree:0 \
		https://github.com/bufbuild/protoc-gen-validate vendor.protogen/tmp && \
		cd vendor.protogen/tmp && \
		git sparse-checkout set --no-cone validate &&\
		git checkout
		mkdir -p vendor.protogen/validate
		mv vendor.protogen/tmp/validate vendor.protogen/
		rm -rf vendor.protogen/tmp
