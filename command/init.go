package command

var txnCommands map[string]TxnCommand
var commands map[string]Desc

func init() {
	// txnCommands will be searched in multi/exec
	txnCommands = map[string]TxnCommand{
		// lists
		"lindex":  LIndex,
		"linsert": LInsert,
		"llen":    LLen,
		"lpop":    LPop,
		"lpush":   LPush,
		"lpushx":  LPushx,
		"lrange":  LRange,
		"lset":    LSet,
		"rpush":   RPush,
		"rpushx":  RPushx,

		// strings
		"get":      Get,
		"set":      Set,
		"mget":     MGet,
		"mset":     MSet,
		"strlen":   Strlen,
		"append":   Append,
		"getset":   GetSet,
		"getrange": GetRange,
		// "msetnx":   MSetNx,
		"setnx":       SetNx,
		"setex":       SetEx,
		"psetex":      PSetEx,
		"setrange":    SetRange,
		"setbit":      SetBit,
		"getbit":      GetBit,
		"incr":        Incr,
		"incrby":      IncrBy,
		"decr":        Decr,
		"decrby":      DecrBy,
		"incrbyfloat": IncrByFloat,

		// keys
		"type":      Type,
		"exists":    Exists,
		"keys":      Keys,
		"del":       Delete,
		"expire":    Expire,
		"expireat":  ExpireAt,
		"pexpire":   PExpire,
		"pexpireat": PExpireAt,
		"persist":   Persist,
		"ttl":       TTL,
		"pttl":      PTTL,
		"object":    Object,
		"scan":      Scan,
		"randomkey": RandomKey,

		// server
		"debug":    Debug,
		"flushdb":  FlushDB,
		"flushall": FlushAll,

		// hashes
		"hdel":         HDel,
		"hset":         HSet,
		"hget":         HGet,
		"hgetall":      HGetAll,
		"hexists":      HExists,
		"hincrby":      HIncrBy,
		"hincrbyfloat": HIncrByFloat,
		"hkeys":        HKeys,
		"hvals":        HVals,
		"hlen":         HLen,
		"hstrlen":      HStrLen,
		"hsetnx":       HSetNX,
		"hmget":        HMGet,
		"hmset":        HMSet,
		"hmslot":       HMSlot,
		"hscan":        HScan,

		// sets
		"sadd":     SAdd,
		"smembers": SMembers,
	}

	// commands contains all commands that open to clients
	// exec should not be in this table to avoid 'initialization loop', and it indeed not necessary be here in fact.
	commands = map[string]Desc{
		// connections
		"auth":   Desc{Proc: Auth, Cons: Constraint{2, flags("sltF"), 0, 0, 0}},
		"echo":   Desc{Proc: Echo, Cons: Constraint{2, flags("F"), 0, 0, 0}},
		"ping":   Desc{Proc: Ping, Cons: Constraint{-1, flags("tF"), 0, 0, 0}},
		"quit":   Desc{Proc: Quit, Cons: Constraint{1, 0, 0, 0, 0}},
		"select": Desc{Proc: Select, Cons: Constraint{2, flags("lF"), 0, 0, 0}},
		"swapdb": Desc{Proc: SwapDB, Cons: Constraint{3, flags("wF"), 0, 0, 0}},

		// transactions, exec and discard should called explicitly, so they are registered here
		"multi":   Desc{Proc: Multi, Cons: Constraint{1, flags("sF"), 0, 0, 0}},
		"watch":   Desc{Proc: Watch, Cons: Constraint{-2, flags("sF"), 1, -1, 1}},
		"unwatch": Desc{Proc: Unwatch, Cons: Constraint{1, flags("sF"), 0, 0, 0}},

		// lists
		"lindex":  Desc{Proc: AutoCommit(LIndex), Cons: Constraint{3, flags("r"), 1, 1, 1}},
		"linsert": Desc{Proc: AutoCommit(LInsert), Cons: Constraint{5, flags("wm"), 1, 1, 1}},
		"llen":    Desc{Proc: AutoCommit(LLen), Cons: Constraint{2, flags("rF"), 1, 1, 1}},
		"lpop":    Desc{Proc: AutoCommit(LPop), Cons: Constraint{2, flags("wF"), 1, 1, 1}},
		"lpush":   Desc{Proc: AutoCommit(LPush), Cons: Constraint{-3, flags("wmF"), 1, 1, 1}},
		"lpushx":  Desc{Proc: AutoCommit(LPushx), Cons: Constraint{3, flags("wmF"), 1, 1, 1}},
		"lrange":  Desc{Proc: AutoCommit(LRange), Cons: Constraint{4, flags("r"), 1, 1, 1}},
		"lset":    Desc{Proc: AutoCommit(LSet), Cons: Constraint{4, flags("wm"), 1, 1, 1}},
		"rpush":   Desc{Proc: AutoCommit(RPush), Cons: Constraint{-3, flags("wmF"), 1, 1, 1}},
		"rpushx":  Desc{Proc: AutoCommit(RPushx), Cons: Constraint{-3, flags("wmF"), 1, 1, 1}},

		// strings
		"get":         Desc{Proc: AutoCommit(Get), Cons: Constraint{2, flags("rF"), 1, 1, 1}},
		"set":         Desc{Proc: AutoCommit(Set), Cons: Constraint{-3, flags("wm"), 1, 1, 1}},
		"setnx":       Desc{Proc: AutoCommit(SetNx), Cons: Constraint{3, flags("wmF"), 1, 1, 1}},
		"setex":       Desc{Proc: AutoCommit(SetEx), Cons: Constraint{4, flags("wm"), 1, 1, 1}},
		"psetex":      Desc{Proc: AutoCommit(PSetEx), Cons: Constraint{4, flags("wm"), 1, 1, 1}},
		"mget":        Desc{Proc: AutoCommit(MGet), Cons: Constraint{-2, flags("rF"), 1, -1, 1}},
		"mset":        Desc{Proc: AutoCommit(MSet), Cons: Constraint{-3, flags("wm"), 1, -1, 2}},
		"msetnx":      Desc{Proc: AutoCommit(MSetNx), Cons: Constraint{-3, flags("wm"), 1, -1, 2}},
		"strlen":      Desc{Proc: AutoCommit(Strlen), Cons: Constraint{2, flags("rF"), 1, 1, 1}},
		"append":      Desc{Proc: AutoCommit(Append), Cons: Constraint{3, flags("wm"), 1, 1, 1}},
		"setrange":    Desc{Proc: AutoCommit(SetRange), Cons: Constraint{4, flags("wm"), 1, 1, 1}},
		"getrange":    Desc{Proc: AutoCommit(GetRange), Cons: Constraint{4, flags("r"), 1, 1, 1}},
		"incr":        Desc{Proc: AutoCommit(Incr), Cons: Constraint{2, flags("wmF"), 1, 1, 1}},
		"decr":        Desc{Proc: AutoCommit(Decr), Cons: Constraint{2, flags("wmF"), 1, 1, 1}},
		"incrby":      Desc{Proc: AutoCommit(IncrBy), Cons: Constraint{3, flags("wmF"), 1, 1, 1}},
		"decrby":      Desc{Proc: AutoCommit(DecrBy), Cons: Constraint{3, flags("wmF"), 1, 1, 1}},
		"incrbyfloat": Desc{Proc: AutoCommit(IncrByFloat), Cons: Constraint{3, flags("wmF"), 1, 1, 1}},
		"setbit":      Desc{Proc: AutoCommit(SetBit), Cons: Constraint{4, flags("wm"), 1, 1, 1}},
		"getbit":      Desc{Proc: AutoCommit(GetBit), Cons: Constraint{3, flags("r"), 1, 1, 1}},

		// keys
		"type":      Desc{Proc: AutoCommit(Type), Cons: Constraint{2, flags("rF"), 1, 1, 1}},
		"exists":    Desc{Proc: AutoCommit(Exists), Cons: Constraint{-2, flags("rF"), 1, -1, 1}},
		"keys":      Desc{Proc: AutoCommit(Keys), Cons: Constraint{-2, flags("rS"), 0, 0, 0}},
		"del":       Desc{Proc: AutoCommit(Delete), Cons: Constraint{-2, flags("w"), 1, -1, 1}},
		"expire":    Desc{Proc: AutoCommit(Expire), Cons: Constraint{3, flags("wF"), 1, 1, 1}},
		"expireat":  Desc{Proc: AutoCommit(ExpireAt), Cons: Constraint{3, flags("wF"), 1, 1, 1}},
		"pexpire":   Desc{Proc: AutoCommit(PExpire), Cons: Constraint{3, flags("wF"), 1, 1, 1}},
		"pexpireat": Desc{Proc: AutoCommit(PExpireAt), Cons: Constraint{3, flags("wF"), 1, 1, 1}},
		"persist":   Desc{Proc: AutoCommit(Persist), Cons: Constraint{2, flags("wF"), 1, 1, 1}},
		"ttl":       Desc{Proc: AutoCommit(TTL), Cons: Constraint{2, flags("rF"), 1, 1, 1}},
		"pttl":      Desc{Proc: AutoCommit(PTTL), Cons: Constraint{2, flags("rF"), 1, 1, 1}},
		"object":    Desc{Proc: AutoCommit(Object), Cons: Constraint{-2, flags("rR"), 0, 0, 0}},
		"scan":      Desc{Proc: AutoCommit(Scan), Cons: Constraint{-2, flags("rR"), 0, 0, 0}},
		"randomkey": Desc{Proc: AutoCommit(RandomKey), Cons: Constraint{1, flags("rR"), 0, 0, 0}},

		// server
		"monitor":  Desc{Proc: Monitor, Cons: Constraint{1, flags("as"), 0, 0, 0}},
		"client":   Desc{Proc: Client, Cons: Constraint{-2, flags("as"), 0, 0, 0}},
		"debug":    Desc{Proc: AutoCommit(Debug), Cons: Constraint{-2, flags("as"), 0, 0, 0}},
		"command":  Desc{Proc: RedisCommand, Cons: Constraint{0, flags("lt"), 0, 0, 0}},
		"flushdb":  Desc{Proc: AutoCommit(FlushDB), Cons: Constraint{-1, flags("w"), 0, 0, 0}},
		"flushall": Desc{Proc: AutoCommit(FlushAll), Cons: Constraint{-1, flags("w"), 0, 0, 0}},
		"time":     Desc{Proc: Time, Cons: Constraint{1, flags("RF"), 0, 0, 0}},
		"info":     Desc{Proc: Info, Cons: Constraint{-1, flags("lt"), 0, 0, 0}},

		// hashes
		"hdel":         Desc{Proc: AutoCommit(HDel), Cons: Constraint{-3, flags("wF"), 1, 1, 1}},
		"hset":         Desc{Proc: AutoCommit(HSet), Cons: Constraint{-4, flags("wmF"), 1, 1, 1}},
		"hget":         Desc{Proc: AutoCommit(HGet), Cons: Constraint{3, flags("rF"), 1, 1, 1}},
		"hgetall":      Desc{Proc: AutoCommit(HGetAll), Cons: Constraint{2, flags("r"), 1, 1, 1}},
		"hexists":      Desc{Proc: AutoCommit(HExists), Cons: Constraint{3, flags("rF"), 1, 1, 1}},
		"hincrby":      Desc{Proc: AutoCommit(HIncrBy), Cons: Constraint{4, flags("wmF"), 1, 1, 1}},
		"hincrbyfloat": Desc{Proc: AutoCommit(HIncrByFloat), Cons: Constraint{4, flags("wmF"), 1, 1, 1}},
		"hkeys":        Desc{Proc: AutoCommit(HKeys), Cons: Constraint{2, flags("rS"), 1, 1, 1}},
		"hvals":        Desc{Proc: AutoCommit(HVals), Cons: Constraint{2, flags("rS"), 1, 1, 1}},
		"hlen":         Desc{Proc: AutoCommit(HLen), Cons: Constraint{2, flags("rF"), 1, 1, 1}},
		"hstrlen":      Desc{Proc: AutoCommit(HStrLen), Cons: Constraint{3, flags("rF"), 1, 1, 1}},
		"hsetnx":       Desc{Proc: AutoCommit(HSetNX), Cons: Constraint{4, flags("wmF"), 1, 1, 1}},
		"hmget":        Desc{Proc: AutoCommit(HMGet), Cons: Constraint{-3, flags("rF"), 1, 1, 1}},
		"hmset":        Desc{Proc: AutoCommit(HMSet), Cons: Constraint{-3, flags("wmF"), 1, 1, 1}},
		"hmslot":       Desc{Proc: AutoCommit(HMSlot), Cons: Constraint{3, flags("wF"), 1, 1, 1}},
		"hscan":        Desc{Proc: AutoCommit(HScan), Cons: Constraint{-3, flags("rR"), 0, 0, 0}},

		// sets
		"sadd":     Desc{Proc: AutoCommit(SAdd), Cons: Constraint{-3, flags("wmF"), 1, 1, 1}},
		"smembers": Desc{Proc: AutoCommit(SMembers), Cons: Constraint{2, flags("rS"), 1, 1, 1}},
	}
}
