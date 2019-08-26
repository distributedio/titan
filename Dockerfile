# Builder image
FROM golang:1.12.7-alpine3.10 as builder

RUN apk add --no-cache \
    make \
    git

COPY . /go/src/github.com/distributedio/titan

WORKDIR /go/src/github.com/distributedio/titan

RUN env GOOS=linux CGO_ENABLED=0 make

# Executable image
FROM scratch

COPY --from=builder /go/src/github.com/distributedio/titan/titan /titan/bin/titan
COPY --from=builder /go/src/github.com/distributedio/titan/conf/titan.toml /titan/conf/titan.toml

WORKDIR /titan

EXPOSE 7369

ENTRYPOINT ["./bin/titan"]
