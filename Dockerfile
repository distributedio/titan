# Builder image
FROM golang:1.16.4-buster as builder

COPY . /go/src/github.com/distributedio/titan

WORKDIR /go/src/github.com/distributedio/titan

RUN \
    go mod download \
 && env GOOS=linux CGO_ENABLED=0 make

FROM alpine as fetch
RUN apk add redis

# Executable image
FROM alpine

COPY --from=builder /go/src/github.com/distributedio/titan/titan /titan/bin/titan
COPY --from=builder /go/src/github.com/distributedio/titan/conf/titan.toml /titan/conf/titan.toml
COPY --from=fetch /usr/bin/redis-cli /usr/bin/redis-cli

WORKDIR /titan

ENV PATH /usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/titan/bin

EXPOSE 7369

ENTRYPOINT ["./bin/titan"]
