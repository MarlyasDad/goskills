APPS := creator api

LOCAL_BIN := $(CURDIR)/bin
MOCKERY_VERSION := v3.9.1
GOOSE_VERSION := v3.30.1
SQLC_VERSION := v1.31.1
SWAG_VERSION := v1.16.6

# Приложение для make swag; по умолчанию первое из APPS.
SWAG_APP ?= $(firstword $(APPS))

.DEFAULT_GOAL := help

.PHONY: help build-all test test-race lint fmt tidy mocks swag .bin-deps clean

help:
	@echo "apps: $(APPS)"
	@echo "  make build-<app>   # bin/<app>"
	@echo "  make run-<app>     # -config configs/.env.local"
	@echo "  make build-all"
	@echo "  make test | test-race | lint | fmt | tidy | clean"
	@echo "  make mocks         # генерация моков по .mockery.yaml"
	@echo "  make swag          # Swagger для SWAG_APP"
	@echo "  make .bin-deps     # goose, sqlc, mockery, swag"

build-all: $(addprefix build-,$(APPS))

build-%:
	go build -o $(LOCAL_BIN)/$* ./cmd/$*

run-%:
	go run ./cmd/$* -config configs/.env.local

test:
	go test ./...

test-race:
	go test -race -shuffle=on -count=1 ./...

# Интеграционные и e2e — раскомментировать, когда появятся тесты.
# При включении добавить цели в .PHONY.
# test-integration:
# 	go test -tags=integration -count=1 ./tests/integration/...
#
# test-e2e:
# 	go test -tags=e2e -count=1 ./tests/e2e/...

lint:
	go vet ./...

fmt:
	go fmt ./...

tidy:
	go mod tidy

mocks:
	$(LOCAL_BIN)/mockery --config=.mockery.yaml

swag:
	$(LOCAL_BIN)/swag init -g ./cmd/$(SWAG_APP)/main.go -o ./internal/transport/$(SWAG_APP)/http/docs --parseDependency --parseInternal

# Go-инструменты — в bin/, не в GOPATH.
# Версии — в переменных выше.
.bin-deps:
	$(info Installing binary dependencies...)

	GOBIN=$(LOCAL_BIN) go install github.com/pressly/goose/v3/cmd/goose@$(GOOSE_VERSION)
	GOBIN=$(LOCAL_BIN) go install github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION)
	GOBIN=$(LOCAL_BIN) go install github.com/vektra/mockery/v3@$(MOCKERY_VERSION)
	GOBIN=$(LOCAL_BIN) go install github.com/swaggo/swag/cmd/swag@$(SWAG_VERSION)
	# gRPC: protoc-gen-go, protoc-gen-go-grpc
	# minimock вместо mockery: github.com/gojuno/minimock/v3/cmd/minimock

clean:
	rm -rf $(LOCAL_BIN)
