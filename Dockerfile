# Builder image
FROM golang:1.10.1-alpine as builder

RUN apk add --no-cache \
    make \
    git

COPY . /go/src/gitlab.meitu.com/platform/titan

WORKDIR /go/src/gitlab.meitu.com/platform/titan

RUN env GOOS=linux CGO_ENABLED=0 go build ./bin/titan/  

## Executable image
FROM scratch

COPY --from=builder /go/src/gitlab.meitu.com/platform/titan/titan /titan

WORKDIR /

EXPOSE 6380

ENTRYPOINT ["/titan"]
