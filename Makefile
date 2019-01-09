PROJECT_NAME := titan
PKG := github.com/meitu/$(PROJECT_NAME)
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)
GITHASH := $(shell git rev-parse --short HEAD)
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/ | grep -v _test.go)

LDFLAGS += -X "$(PKG)/context.ReleaseVersion=$(shell git tag  --contains)"
LDFLAGS += -X "$(PKG)/context.BuildTS=$(shell date -u '+%Y-%m-%d %I:%M:%S')"
LDFLAGS += -X "$(PKG)/context.GitHash=$(GITHASH)"
LDFLAGS += -X "$(PKG)/context.GolangVersion=$(shell go version)"
LDFLAGS += -X "$(PKG)/context.GitLog=$(shell git log --abbrev-commit --oneline -n 1 | sed 's/$(GITHASH)//g')"
LDFLAGS += -X "$(PKG)/context.GitBranch=$(shell git rev-parse --abbrev-ref HEAD)"

.PHONY: all build clean test coverage lint proto
all: build

test:
	env GO111MODULE=on go test -short ${PKG_LIST}

coverage:
	env GO111MODULE=on go test -covermode=count -v -coverprofile cover.cov ${PKG_LIST}

build:
	env GO111MODULE=on go build -ldflags '$(LDFLAGS)' -o titan ./bin/titan/

clean:
	rm -f ./titan

lint:
	gometalinter --fast -t --errors --enable-gc ${GO_FILES}

proto:
	cd ./db/zlistproto && protoc --gofast_out=plugins=grpc:. ./zlist.proto

release: build
	./tools/release.sh
