GO ?= go
BINARY := redis-tui

.PHONY: build test integration vet fmt cross clean

build:
	$(GO) build -o bin/$(BINARY) .

test:
	$(GO) test ./...

integration:
	$(GO) test -tags integration ./test/integration

vet:
	$(GO) vet ./...

fmt:
	$(GO) fmt ./...

cross:
	mkdir -p dist
	GOOS=windows GOARCH=amd64 $(GO) build -o dist/$(BINARY)-windows-amd64.exe .
	GOOS=linux GOARCH=amd64 $(GO) build -o dist/$(BINARY)-linux-amd64 .
	GOOS=darwin GOARCH=amd64 $(GO) build -o dist/$(BINARY)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 $(GO) build -o dist/$(BINARY)-darwin-arm64 .

clean:
	rm -rf bin dist
