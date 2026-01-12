.PHONY: build run generate clean dev lint test reset

GO ?= go
GOLANGCI_LINT ?= golangci-lint

build: generate
	$(GO) build -o bin/server ./cmd/server

run: build
	./bin/server

dev: generate
	$(GO) run ./cmd/server

generate: generate-templ generate-sqlc

generate-templ:
	$(GO) tool templ generate

generate-sqlc:
	$(GO) tool sqlc generate

clean:
	rm -rf bin/
	find . -name "*_templ.go" -delete

tidy:
	$(GO) mod tidy

check:
	$(GO) vet ./...
	$(GO) build ./...

lint:
	$(GOLANGCI_LINT) run

test:
	$(GO) test ./...

reset:
	rm -f data.db token.json
	rm -rf cache/
