# Thanos borns from the Titan

This is an experiment project to verify my design ideas.

## Features

* distributed transactions
* expire on arbitrary keys
* delete complex data structure with gc

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
- [ ] info
- [ ] slowlog

### Keys
- [x] del 
- [x] type
- [x] exists
- [x] expire
- [x] expireat
- [ ] object 
- [ ] pexpire
- [ ] pexpireat
- [ ] ttl
- [ ] pttl
- [ ] randomkey
- [ ] touch 
- [ ] keys
- [ ] scan
- [ ] unlink

### Strings

- [x] get 
- [x] set 
- [x] mget 
- [x] mset 
- [x] strlen
- [ ] incr
- [ ] incrby
- [ ] decr
- [ ] decrby
- [ ] append
- [ ] bitcount
- [ ] bitfield
- [ ] bitop
- [ ] bitpos
- [ ] getbit
- [ ] getrange
- [ ] getset
- [ ] incrbyfloat 
- [ ] msetnx
- [ ] psetex
- [ ] setbit
- [ ] setex
- [ ] setnx
- [ ] setrange

### Lists

- [x] lpush
- [x] lpop
- [x] lrange
- [x] linsert
- [ ] blpop
- [ ] brpop
- [ ] brpoplpush
- [ ] lindex
- [ ] llen
- [ ] lpushhx
- [ ] lrem
- [ ] lset
- [ ] ltrim
- [ ] rpop
- [ ] rpoplpush
- [ ] rpush
- [ ] rpushhx

### Hashes
- [x] hset
- [x] hget
- [x] hgetall
- [x] hdel
- [x] hexists
- [x] hincrby
- [x] hincrbyfloat
- [x] hkeys
- [x] hlen
- [x] hmget
- [x] hmset
- [ ] hscan
- [x] hsetnx
- [x] hstrlen
- [x] hvals

### Sets

- [x] sadd
- [ ] scard
- [ ] sdiff
- [ ] sdiffstore
- [ ] sinter
- [ ] sinterstore
- [ ] sismember
- [x] smembers
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
