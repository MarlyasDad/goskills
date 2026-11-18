APP ?= <appname>
GO ?= go
BIN := bin/$(APP)

.PHONY: build run test test-race lint fmt tidy clean

build:
	$(GO) build -o $(BIN) ./cmd/$(APP)

run:
	$(GO) run ./cmd/$(APP) $(ARGS)

test:
	$(GO) test ./...

test-race:
	$(GO) test -race -shuffle=on -count=1 ./...

lint:
	$(GO) vet ./...

fmt:
	$(GO) fmt ./...

tidy:
	$(GO) mod tidy

clean:
	rm -rf bin
