.PHONY: test demo verify build clean verify-demo self-check

VERSION ?= dev
LDFLAGS := -X main.version=$(VERSION)

test:
	go test -count=1 ./...

demo:
	go run ./cmd/auditpack demo --out ./out

verify: test demo

build:
	mkdir -p bin
	go build -ldflags "$(LDFLAGS)" -o bin/auditpack ./cmd/auditpack

clean:
	rm -rf ./out ./bin

verify-demo:
	go run ./cmd/auditpack demo --out ./out
	go run ./cmd/auditpack verify --pack ./out/case01 --in ./fixtures/input/case01 --strict

self-check:
	go run ./cmd/auditpack self-check
