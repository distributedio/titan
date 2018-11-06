# Builder image
FROM golang:1.10.1-alpine as builder

RUN apk add --no-cache \
    make \
    git

COPY . /go/src/gitlab.meitu.com/platform/thanos

WORKDIR /go/src/gitlab.meitu.com/platform/thanos

RUN env GOOS=linux CGO_ENABLED=0 go build ./bin/thanos/  

## Executable image
FROM scratch

COPY --from=builder /go/src/gitlab.meitu.com/platform/thanos/thanos /thanos

WORKDIR /

EXPOSE 6380

ENTRYPOINT ["/thanos"]
