# Thanos [![build status](https://gitlab.meitu.com/platform/thanos/badges/dev/build.svg)](https://gitlab.meitu.com/platform/thanos/commits/dev) [![coverage report](https://gitlab.meitu.com/platform/thanos/badges/dev/coverage.svg)](https://gitlab.meitu.com/platform/thanos/commits/dev)
An distributed implementation of Redis compatible layer over TiKV

Visit [Thanos](http://cf.meitu.com/confluence/pages/viewpage.action?pageId=29745824) for more informations.

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
- [ ] incr
- [ ] incrby
- [ ] decr
- [ ] decrby
- [x] append
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

### List

- [x] lrange
- [x] linsert
- [x] lindex
- [x] llen
- [x] lrem
- [x] lset
- [x] ltrim
- [x] lpush
- [x] lpop
- [x] lpushhx
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
