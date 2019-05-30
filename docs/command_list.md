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
- [x] rpush
- [ ] rpushhx
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
- [ ] zcount
- [ ] zincrby
- [ ] zinterstore
- [ ] zlexcount
- [ ] zpopmax
- [ ] zpopmin
- [x] zrange
- [ ] zrangebylex
- [ ] zrevrangebylex
- [ ] zrangebyscore
- [ ] zrank
- [x] zrem
- [ ] zremrangebylex
- [ ] zremrangebyrank
- [ ] zremrangebyscore
- [x] zrevrange
- [ ] zrevrangebyscore
- [ ] zrevrank
- [x] zscore
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

## Not supported yet
___NOTICE: commands beyond this table has already been fully supported___

|command|type|status|
|---|---|---|
|swapdb |Connections| |
|slowlog |Server | |
|touch  |Keys | |
|unlink |Keys | |
|bitcount|Strings|| 
|bitfield|Strings||
|bitop   |Strings||
|bitpos  |Strings||
|getbit  |Strings||
|msetnx  |Strings||
|setbit  |Strings||
|setrange|Strings||
|ltrim     |List||
|lrem      |List||
|rpop      |List||
|rpoplpush |List||
|rpushhx   |List||
|blpop     |List||
|brpop     |List||
|brpoplpush|List||
|sinterstore|Sets| |
|sdiffstore |Sets| |
|srandmember|Sets| |
|sunionstore|Sets| |
|sscan      |Sets| |
|bzpopmin        |Sorted set| |
|bzpopmax        |Sorted set| |
|zcount          |Sorted set| |
|zincrby         |Sorted set| |
|zinterstore     |Sorted set| |
|zlexcount       |Sorted set| |
|zpopmax         |Sorted set| |
|zpopmin         |Sorted set| |
|zrevrank        |Sorted set| |
|zunionstore     |Sorted set| |
|zscan           |Sorted set| |
|zrangebylex     |Sorted set| |
|zrevrangebylex  |Sorted set| |
|zrangebyscore   |Sorted set| |
|zrank           |Sorted set| |
|zremrangebylex  |Sorted set| |
|zremrangebyrank |Sorted set| |
|zremrangebyscore|Sorted set| |
|zrevrangebyscore|Sorted set| |
|geoadd           |Geo| |
|geohash          |Geo| |
|geopos           |Geo| |
|geodist          |Geo| |
|georadius        |Geo| |
|georadiusbymember|Geo| |
|pfadd  |Hyperhyperlog| |
|pfcount|Hyperhyperlog| |
|pfmerge|Hyperhyperlog| |
|psubscribe  |Pub/Sub| |
|pubsub      |Pub/Sub| |
|publish     |Pub/Sub| |
|punsubscribe|Pub/Sub| |
|subscribe   |Pub/Sub| |
|unsubscribe |Pub/Sub| |
|eval         |Script| |
|evalsha      |Script| |
|script debug |Script| |
|script flush |Script| |
|script kill  |Script| |
|script load  |Script| |
|script exists|Script| |
|xadd      |Stream| |
|xrange    |Stream| |
|xrevrange |Stream| |
|xlen      |Stream| |
|xread     |Stream| |
|xpending  |Stream| |
|xreadgroup|Stream| |
