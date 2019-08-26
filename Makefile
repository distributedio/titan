PROJECT_NAME := titan
PKG := github.com/distributedio/$(PROJECT_NAME)
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)
GITHASH := $(shell git rev-parse --short HEAD)
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/ | grep -v _test.go)

LDFLAGS += -X "$(PKG)/context.ReleaseVersion=$(shell git tag  --contains)"
LDFLAGS += -X "$(PKG)/context.BuildTS=$(shell date -u '+%Y-%m-%d %I:%M:%S')"
LDFLAGS += -X "$(PKG)/context.GitHash=$(GITHASH)"
LDFLAGS += -X "$(PKG)/context.GolangVersion=$(shell go version)"
LDFLAGS += -X "$(PKG)/context.GitLog=$(shell git log --abbrev-commit --oneline -n 1 | sed 's/$(GITHASH)//g' | sed 's/"//g')"
LDFLAGS += -X "$(PKG)/context.GitBranch=$(shell git rev-parse --abbrev-ref HEAD)"

.PHONY: all build clean test coverage lint proto
all: build token

test:
	env GO111MODULE=on go test -short ${PKG_LIST}

coverage:
	env GO111MODULE=on go test -v -covermode=count -coverprofile=coverage.out ${PKG_LIST}

build:
	env GO111MODULE=on go build -ldflags '$(LDFLAGS)' -o titan ./bin/titan/

token: tools/token/main.go command/common.go
	env GO111MODULE=on go build -ldflags '$(LDFLAGS)' -o token ./tools/token/

clean:
	rm -f ./titan
	rm -rf ./token

lint:
	gometalinter --fast -t --errors --enable-gc ${GO_FILES}

proto:
	cd ./db/zlistproto && protoc --gofast_out=plugins=grpc:. ./zlist.proto

release: build
	./tools/release.sh
