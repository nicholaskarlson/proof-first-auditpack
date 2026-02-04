.PHONY: test demo build clean verify-demo self-check

VERSION ?= dev
LDFLAGS := -X main.version=$(VERSION)

test:
	go test ./...

demo:
	go run ./cmd/auditpack demo --out ./out

build:
	mkdir -p bin
	go build -ldflags "$(LDFLAGS)" -o bin/auditpack ./cmd/auditpack

clean:
	rm -rf ./out ./bin

verify-demo:
	go run ./cmd/auditpack demo --out ./out
	go run ./cmd/auditpack verify --out ./out --in ./out/demo_input --strict

self-check:
	go run ./cmd/auditpack self-check
