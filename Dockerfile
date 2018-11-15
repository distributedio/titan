# Builder image
FROM golang:1.10.1-alpine as builder

RUN apk add --no-cache \
    make \
    git

COPY . /go/src/github.com/meitu/titan

WORKDIR /go/src/github.com/meitu/titan

RUN env GOOS=linux CGO_ENABLED=0 go build ./bin/titan/  

## Executable image
FROM scratch

COPY --from=builder /go/src/github.com/meitu/titan/titan /titan

WORKDIR /

EXPOSE 6380

ENTRYPOINT ["/titan"]
