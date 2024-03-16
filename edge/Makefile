DOCKER_PLATFORM := --platform linux/amd64
GO_ENV := GOARCH=amd64 GOOS=linux CGO_ENABLED=0

VERSION := $(shell date +%Y%m%d-%H%M%S)

.PHONY: clean
clean:
	rm -rf bin

.PHONY: fmt
fmt:
	goimports -w .

.PHONY: vet
vet:
	go vet ./...

.PHONY: staticcheck
staticcheck:
	staticcheck ./...

.PHONY: build
build: clean fmt vet staticcheck
	$(GO_ENV) go build -ldflags '-w -s -X main.version=$(VERSION)' -trimpath -o bin/edge ./cmd/edge
