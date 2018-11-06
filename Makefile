PROJECT_NAME := thanos
PKG := gitlab.meitu.com/platform/$(PROJECT_NAME)
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)
GITHASH := $(shell git rev-parse --short HEAD)
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/ | grep -v _test.go)

LDFLAGS += -X "$(PKG)/server.ReleaseVersion=$(shell git tag  --contains)"
LDFLAGS += -X "$(PKG)/server.BuildTS=$(shell date -u '+%Y-%m-%d %I:%M:%S')"
LDFLAGS += -X "$(PKG)/server.GitHash=$(GITHASH)"
LDFLAGS += -X "$(PKG)/server.GolangVersion=$(shell go version)"
LDFLAGS += -X "$(PKG)/server.GitLog=$(shell git log --abbrev-commit --oneline -n 1 | sed 's/$(GITHASH)//g')"
LDFLAGS += -X "$(PKG)/server.GitBranch=$(shell git rev-parse --abbrev-ref HEAD)"

.PHONY: all build clean test coverage lint proto
all: build

test:
	go test -short ${PKG_LIST}

coverage:
	go test -covermode=count -v -coverprofile cover.cov ${PKG_LIST}

build:
	go build -ldflags '$(LDFLAGS)' -o thanos ./bin/thanos/

clean:
	rm -f ./thanos

cleanvendor:
	find vendor \( -type f -or -type l \)  -not -name "*.go" -not -name "LICENSE" -not -name "*.s" -not -name "PATENTS" | xargs -I {} rm {}
	find vendor -type f -name "*_generated.go" | xargs -I {} rm {}
	find vendor -type f -name "*_test.go" | xargs -I {} rm {}
	find vendor -type d -name "_vendor" | xargs -I {} rm -rf {}
	find vendor -type d -empty | xargs -I {} rm -rf {}

lint:
	gometalinter --fast -t --errors --enable-gc ${GO_FILES}

proto:
	cd ./db/zlistproto && protoc --gofast_out=plugins=grpc:. ./zlist.proto

release: build
	./tools/release.sh
