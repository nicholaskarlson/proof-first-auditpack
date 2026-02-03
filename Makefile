.PHONY: test demo build clean

test:
	go test ./...

demo:
	go run ./cmd/auditpack demo --out ./out

build:
	mkdir -p bin
	go build -o bin/auditpack ./cmd/auditpack

clean:
	rm -rf ./out ./bin
