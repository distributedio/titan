## Commands list

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
- [x] touch
- [x] keys
- [x] scan
- [x] unlink

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
- [x] ltrim
- [x] lrem
- [x] rpop
- [x] rpoplpush
- [x] rpush
- [x] rpushx
- [ ] blpop
- [ ] brpop
- [ ] brpoplpush

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
- [x] hscan
- [x] hsetnx
- [x] hstrlen
- [x] hvals

### Sets

- [x] sadd
- [x] scard
- [x] sdiff
- [ ] sdiffstore
- [x] sinter
- [ ] sinterstore
- [x] sismember
- [x] smembers
- [x] smove
- [x] spop
- [ ] srandmember
- [x] srem
- [x] sunion
- [ ] sunionstore
- [ ] sscan

### Sorted Sets

- [ ] bzpopmin
- [ ] bzpopmax
- [x] zadd
- [x] zcard
- [x] zcount
- [ ] zincrby
- [ ] zinterstore
- [x] zlexcount
- [ ] zpopmax
- [ ] zpopmin
- [x] zrange
- [x] zrangebylex
- [x] zrevrangebylex
- [x] zrangebyscore
- [ ] zrank
- [x] zrem
- [x] zremrangebylex
- [ ] zremrangebyrank
- [ ] zremrangebyscore
- [x] zrevrange
- [x] zrevrangebyscore
- [ ] zrevrank
- [x] zscore
- [ ] zunionstore
- [x] zscan

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
