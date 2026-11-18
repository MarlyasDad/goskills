APPS := creator api

LOCAL_BIN := $(CURDIR)/bin

.DEFAULT_GOAL := help

.PHONY: help build-all test test-race lint fmt tidy tools clean

help:
	@echo "apps: $(APPS)"
	@echo "  make build-<app>   # bin/<app>"
	@echo "  make run-<app>     # -config configs/.env.local"
	@echo "  make build-all"
	@echo "  make test | test-race | lint | fmt | tidy | tools | clean"

build-all: $(addprefix build-,$(APPS))

build-%:
	go build -o $(LOCAL_BIN)/$* ./cmd/$*

run-%:
	go run ./cmd/$* -config configs/.env.local

test:
	go test ./...

test-race:
	go test -race -shuffle=on -count=1 ./...

lint:
	go vet ./...

fmt:
	go fmt ./...

tidy:
	go mod tidy

# protoc, mockery, swag, goose и другие Go-инструменты — в bin/, не в GOPATH.
# Версии зафиксировать вместо @latest.
tools:
	GOBIN=$(LOCAL_BIN) go install \
		github.com/vektra/mockery/v2@latest \
		github.com/swaggo/swag/cmd/swag@latest \
		github.com/pressly/goose/v3/cmd/goose@latest

clean:
	rm -rf $(LOCAL_BIN)
