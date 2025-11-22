PROJECT_NAME := titan
PKG := github.com/distributedio/$(PROJECT_NAME)
GITHASH := $(shell git rev-parse --short HEAD)
GO_LOGS := $(shell git log --abbrev-commit --oneline -n 1 | sed 's/$(GITHASH)//g' | sed 's/"//g' | sed "s/'//g")

LDFLAGS += -X "$(PKG)/context.ReleaseVersion=$(shell git tag  --contains)"
LDFLAGS += -X "$(PKG)/context.BuildTS=$(shell date -u '+%Y-%m-%d %I:%M:%S')"
LDFLAGS += -X "$(PKG)/context.GitHash=$(GITHASH)"
LDFLAGS += -X "$(PKG)/context.GolangVersion=$(shell go version)"
LDFLAGS += -X "$(PKG)/context.GitLog=$(GO_LOGS)"
LDFLAGS += -X "$(PKG)/context.GitBranch=$(shell git rev-parse --abbrev-ref HEAD)"

GOMINVERSION := 1.21
GOVERSION := $(shell go version | awk '{print $$3}' | tr -d 'go')

.PHONY: all build clean test coverage lint proto check-go-version
all: build token

check-go-version:
	@if [ "$(shell printf '%s\n' $(GOMINVERSION) $(GOVERSION) | sort -V | head -n1)" != "$(GOMINVERSION)" ]; then \
		echo "Error: Go version must be >= $(GOMINVERSION), found $(GOVERSION)"; \
		exit 1; \
	fi

test: check-go-version
	env GO111MODULE=on go test -short ./...

coverage: check-go-version
	env GO111MODULE=on go test -covermode=count -v -coverprofile cover.cov ./...

build: check-go-version
	env GO111MODULE=on go build -ldflags '$(LDFLAGS)' -o titan ./bin/titan/

token: tools/token/main.go command/common.go
	env GO111MODULE=on go build -ldflags '$(LDFLAGS)' -o token ./tools/token/

clean:
	rm -f ./titan
	rm -rf ./token

lint:
	golangci-lint run

proto:
	cd ./db/zlistproto && protoc --gofast_out=plugins=grpc:. ./zlist.proto

release: build
	./tools/release.sh
