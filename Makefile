REVISION := $(shell git describe --long --always --abbrev=12 --tags --dirty 2>/dev/null || echo UNRELEASED)

GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)

BIN:=factoid

all: $(BIN)

release: BIN=factoid-$(REVISION)_$(GOOS)_$(GOARCH)
release: GO_LDFLAGS += -ldflags "-s"
release: $(BIN)

$(BIN):
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(BIN) $(GO_LDFLAGS)

check:
	go mod tidy
	go mod verify
	go vet ./...
	staticcheck ./...
	go test -race -vet=off ./...

.PHONY: check release
