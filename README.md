# Titan

[![Build Status](https://travis-ci.org/meitu/titan.svg?branch=master)](https://travis-ci.org/meitu/titan)
[![Go Report Card](https://goreportcard.com/badge/github.com/meitu/titan)](https://goreportcard.com/report/github.com/meitu/titan)
[![Coverage Status](https://coveralls.io/repos/github/meitu/titan/badge.svg?branch=master)](https://coveralls.io/github/meitu/titan?branch=master)

A distributed implementation of Redis compatible layer based on [TiKV](https://github.com/tikv/tikv/)

## Status

Active development, not ready for production. We welcome any form of contributions!

## Features

* Completely compatible with redis protocol
* Full distributed transaction with strong consistency
* Multi-tenancy support
* No painful scale out
* High availability 

Thanks [TiKV](https://github.com/tikv/tikv/) for supporting the core features 

## Installing

[Deploy Titan](docs/ops/deploy.md)

## Benchmarks

[Titan Benchmarks](docs/benchmark/benchmark.md)

## Commands

### Connections
- [x] auth 
- [x] echo 
- [x] ping
- [x] quit 
- [x] select 
- [ ] swapdb, not supported

### Transactions
- [x] multi 
- [x] exec
- [x] discard
- [x] watch
- [x] unwatch

### Server
- [x] client list
- [x] client kill
- [x] client pause
- [x] client reply 
- [x] client getname
- [x] client setname
- [x] monitor
- [x] debug object
- [x] flushdb
- [x] flushall
- [x] time
- [x] command
- [x] command count
- [x] command getkeys
- [x] command info
- [x] info
- [ ] slowlog

### Keys
- [x] del 
- [x] type
- [x] exists
- [x] expire
- [x] expireat
- [x] object 
- [x] pexpire
- [x] pexpireat
- [x] ttl
- [x] pttl
- [x] randomkey
- [ ] touch 
- [x] keys
- [x] scan
- [ ] unlink

### Strings

- [x] get 
- [x] set 
- [x] mget 
- [x] mset 
- [x] strlen
- [x] incr
- [x] incrby
- [x] decr
- [x] decrby
- [x] append
- [ ] bitcount
- [ ] bitfield
- [ ] bitop
- [ ] bitpos
- [ ] getbit
- [x] getrange
- [x] getset
- [x] incrbyfloat 
- [ ] msetnx
- [x] psetex
- [ ] setbit
- [x] setex
- [x] setnx
- [ ] setrange

### List

- [x] lrange
- [x] linsert
- [x] lindex
- [x] llen
- [x] lset
- [x] lpush
- [x] lpop
- [x] lpushx
- [ ] ltrim
- [ ] lrem
- [ ] rpop
- [ ] rpoplpush
- [ ] rpush
- [ ] rpushhx
- [ ] blpop
- [ ] brpop
- [ ] brpoplpush

### Hashes
- [ ] hset
- [ ] hget
- [ ] hgetall
- [ ] hdel
- [ ] hexists
- [ ] hincrby
- [ ] hincrbyfloat
- [ ] hkeys
- [ ] hlen
- [ ] hmget
- [ ] hmset
- [ ] hscan
- [ ] hsetnx
- [ ] hstrlen
- [ ] hvals

### Sets

- [ ] sadd
- [ ] scard
- [ ] sdiff
- [ ] sdiffstore
- [ ] sinter
- [ ] sinterstore
- [ ] sismember
- [ ] smembers
- [ ] smove
- [ ] spop
- [ ] srandmember
- [ ] srem
- [ ] sunion
- [ ] sunionstore
- [ ] sscan

### Sorted Sets

- [ ] bzpopmin
- [ ] bzpopmax
- [ ] zadd
- [ ] zcard
- [ ] zcount
- [ ] zincrby
- [ ] zinterstore
- [ ] zlexcount
- [ ] zpopmax
- [ ] zpopmin
- [ ] zrange
- [ ] zrangebylex
- [ ] zrevrangebylex
- [ ] zrangebyscore
- [ ] zrank
- [ ] zrem
- [ ] zremrangebylex
- [ ] zremrangebyrank
- [ ] zremrangebyscore
- [ ] zrevrange
- [ ] zrevrangebyscore
- [ ] zrevrank
- [ ] zscore
- [ ] zunionstore
- [ ] zscan

### Geo

- [ ] geoadd
- [ ] geohash
- [ ] geopos
- [ ] geodist
- [ ] georadius
- [ ] georadiusbymember

### hyperloglog

- [ ] pfadd
- [ ] pfcount
- [ ] pfmerge

### Pub/Sub

- [ ] psubscribe
- [ ] pubsub
- [ ] publish
- [ ] punsubscribe
- [ ] subscribe
- [ ] unsubscribe

### Scripting

- [ ] eval
- [ ] evalsha
- [ ] script debug
- [ ] script exists
- [ ] script flush
- [ ] script kill
- [ ] script load

### Streams

- [ ] xadd
- [ ] xrange
- [ ] xrevrange
- [ ] xlen
- [ ] xread
- [ ] xreadgroup
- [ ] xpending
