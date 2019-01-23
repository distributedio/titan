# Builder image
FROM golang:1.11.4-alpine3.8 as builder

RUN apk add --no-cache \
    make \
    git

COPY . /go/src/github.com/meitu/titan

WORKDIR /go/src/github.com/meitu/titan

RUN env GOOS=linux CGO_ENABLED=0 make

# Executable image
FROM scratch

COPY --from=builder /go/src/github.com/meitu/titan/titan /titan/bin/titan
COPY --from=builder /go/src/github.com/meitu/titan/conf/titan.toml /titan/conf/titan.toml

WORKDIR /titan

EXPOSE 7369

ENTRYPOINT ["./bin/titan"]
