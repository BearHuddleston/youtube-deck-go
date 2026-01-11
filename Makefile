.PHONY: build run generate clean dev

GO := /usr/local/go/bin/go

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
