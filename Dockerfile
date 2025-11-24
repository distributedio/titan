# Builder image
FROM golang:1.25.4 as builder

COPY . /go/src/github.com/distributedio/titan

WORKDIR /go/src/github.com/distributedio/titan

RUN \
    go mod download \
 && env GOOS=linux CGO_ENABLED=0 make

# Executable image
FROM alpine:3.22.2
RUN apk add redis

COPY --from=builder /go/src/github.com/distributedio/titan/titan /titan/bin/titan
COPY --from=builder /go/src/github.com/distributedio/titan/conf/titan.toml /titan/conf/titan.toml

WORKDIR /titan

ENV PATH /usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/titan/bin

EXPOSE 7369

ENTRYPOINT ["./bin/titan"]
