package command

var txnCommands map[string]TxnCommand
var commands map[string]InfoCommand

func init() {
	// txnCommands will be searched in multi/exec
	txnCommands = map[string]TxnCommand{
		// lists
		"lindex":    LIndex,
		"linsert":   LInsert,
		"llen":      LLen,
		"lpop":      LPop,
		"lpush":     LPush,
		"lpushx":    LPushx,
		"lrange":    LRange,
		"lrem":      LRem,
		"lset":      LSet,
		"ltrim":     LTrim,
		"rpop":      RPop,
		"rpoplpush": RPopLPush,
		"rpush":     RPush,
		"rpushx":    RPushx,

		// strings
		"get":      Get,
		"set":      Set,
		"mget":     MGet,
		"mset":     MSet,
		"strlen":   Strlen,
		"append":   Append,
		"getset":   GetSet,
		"getrange": GetRange,
		"msetnx":   MSetNx,
		"setnx":    SetNx,
		"setex":    SetEx,
		"psetex":   PSetEx,
		// "setrange":    SetRange,
		// "setbit":      SetBit,
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

		// sets
		"sadd":     SAdd,
		"smembers": SMembers,
	}

	// commands contains all commands that open to clients
	// exec should not be in this table to avoid 'initialization loop', and it indeed not necessary be here in fact.
	commands = map[string]InfoCommand{
		// connections
		"auth":   InfoCommand{Proc: Auth, Cons: Constraint{2, flags("sltF"), 0, 0, 0}},
		"echo":   InfoCommand{Proc: Echo, Cons: Constraint{2, flags("F"), 0, 0, 0}},
		"ping":   InfoCommand{Proc: Ping, Cons: Constraint{-1, flags("tF"), 0, 0, 0}},
		"quit":   InfoCommand{Proc: Quit, Cons: Constraint{1, 0, 0, 0, 0}},
		"select": InfoCommand{Proc: Select, Cons: Constraint{2, flags("lF"), 0, 0, 0}},
		"swapdb": InfoCommand{Proc: SwapDB, Cons: Constraint{3, flags("wF"), 0, 0, 0}},

		// transactions, exec and discard should called explicitly, so they are registered here
		"multi":   InfoCommand{Proc: Multi, Cons: Constraint{1, flags("sF"), 0, 0, 0}},
		"watch":   InfoCommand{Proc: Watch, Cons: Constraint{-2, flags("sF"), 1, -1, 1}},
		"unwatch": InfoCommand{Proc: Unwatch, Cons: Constraint{1, flags("sF"), 0, 0, 0}},

		// lists
		"lindex":    InfoCommand{Proc: AutoCommit(LIndex), Cons: Constraint{3, flags("r"), 1, 1, 1}},
		"linsert":   InfoCommand{Proc: AutoCommit(LInsert), Cons: Constraint{5, flags("wm"), 1, 1, 1}},
		"llen":      InfoCommand{Proc: AutoCommit(LLen), Cons: Constraint{2, flags("rF"), 1, 1, 1}},
		"lpop":      InfoCommand{Proc: AutoCommit(LPop), Cons: Constraint{2, flags("wF"), 1, 1, 1}},
		"lpush":     InfoCommand{Proc: AutoCommit(LPush), Cons: Constraint{-3, flags("wmF"), 1, 1, 1}},
		"lpushx":    InfoCommand{Proc: AutoCommit(LPushx), Cons: Constraint{3, flags("wmF"), 1, 1, 1}},
		"lrange":    InfoCommand{Proc: AutoCommit(LRange), Cons: Constraint{4, flags("r"), 1, 1, 1}},
		"lrem":      InfoCommand{Proc: AutoCommit(LRem), Cons: Constraint{4, flags("w"), 1, 1, 1}},
		"lset":      InfoCommand{Proc: AutoCommit(LSet), Cons: Constraint{4, flags("wm"), 1, 1, 1}},
		"ltrim":     InfoCommand{Proc: AutoCommit(LTrim), Cons: Constraint{4, flags("w"), 1, 1, 1}},
		"rpop":      InfoCommand{Proc: AutoCommit(RPop), Cons: Constraint{2, flags("wF"), 1, 1, 1}},
		"rpoplpush": InfoCommand{Proc: AutoCommit(RPopLPush), Cons: Constraint{4, flags("wms"), 1, 2, 1}},
		"rpush":     InfoCommand{Proc: AutoCommit(RPush), Cons: Constraint{-3, flags("wmF"), 1, 1, 1}},
		"rpushx":    InfoCommand{Proc: AutoCommit(RPushx), Cons: Constraint{-3, flags("wmF"), 1, 1, 1}},

		// strings
		"get":    InfoCommand{Proc: AutoCommit(Get), Cons: Constraint{2, flags("rF"), 1, 1, 1}},
		"set":    InfoCommand{Proc: AutoCommit(Set), Cons: Constraint{-3, flags("wm"), 1, 1, 1}},
		"setnx":  InfoCommand{Proc: AutoCommit(SetNx), Cons: Constraint{3, flags("wmF"), 1, 1, 1}},
		"setex":  InfoCommand{Proc: AutoCommit(SetEx), Cons: Constraint{4, flags("wm"), 1, 1, 1}},
		"psetex": InfoCommand{Proc: AutoCommit(PSetEx), Cons: Constraint{4, flags("wm"), 1, 1, 1}},
		"mget":   InfoCommand{Proc: AutoCommit(MGet), Cons: Constraint{-2, flags("rF"), 1, -1, 1}},
		"mset":   InfoCommand{Proc: AutoCommit(MSet), Cons: Constraint{-3, flags("wm"), 1, -1, 2}},
		"msetnx": InfoCommand{Proc: AutoCommit(MSetNx), Cons: Constraint{-3, flags("wm"), 1, -1, 2}},
		"strlen": InfoCommand{Proc: AutoCommit(Strlen), Cons: Constraint{2, flags("rF"), 1, 1, 1}},
		"append": InfoCommand{Proc: AutoCommit(Append), Cons: Constraint{3, flags("wm"), 1, 1, 1}},
		// "setrange": InfoCommand{Proc: AutoCommit(SetRange), Cons: Constraint{4, flags("wm"), 1, 1, 1}},
		"getrange":    InfoCommand{Proc: AutoCommit(GetRange), Cons: Constraint{4, flags("r"), 1, 1, 1}},
		"incr":        InfoCommand{Proc: AutoCommit(Incr), Cons: Constraint{2, flags("wmF"), 1, 1, 1}},
		"decr":        InfoCommand{Proc: AutoCommit(Decr), Cons: Constraint{2, flags("wmF"), 1, 1, 1}},
		"incrby":      InfoCommand{Proc: AutoCommit(IncrBy), Cons: Constraint{3, flags("wmF"), 1, 1, 1}},
		"decrby":      InfoCommand{Proc: AutoCommit(DecrBy), Cons: Constraint{3, flags("wmF"), 1, 1, 1}},
		"incrbyfloat": InfoCommand{Proc: AutoCommit(IncrByFloat), Cons: Constraint{3, flags("wmF"), 1, 1, 1}},

		// keys
		"type":      InfoCommand{Proc: AutoCommit(Type), Cons: Constraint{2, flags("rF"), 1, 1, 1}},
		"exists":    InfoCommand{Proc: AutoCommit(Exists), Cons: Constraint{-2, flags("rF"), 1, -1, 1}},
		"keys":      InfoCommand{Proc: AutoCommit(Keys), Cons: Constraint{-2, flags("rS"), 0, 0, 0}},
		"del":       InfoCommand{Proc: AutoCommit(Delete), Cons: Constraint{-2, flags("w"), 1, -1, 1}},
		"expire":    InfoCommand{Proc: AutoCommit(Expire), Cons: Constraint{3, flags("wF"), 1, 1, 1}},
		"expireat":  InfoCommand{Proc: AutoCommit(ExpireAt), Cons: Constraint{3, flags("wF"), 1, 1, 1}},
		"pexpire":   InfoCommand{Proc: AutoCommit(PExpire), Cons: Constraint{3, flags("wF"), 1, 1, 1}},
		"pexpireat": InfoCommand{Proc: AutoCommit(PExpireAt), Cons: Constraint{3, flags("wF"), 1, 1, 1}},
		"persist":   InfoCommand{Proc: AutoCommit(Persist), Cons: Constraint{2, flags("wF"), 1, 1, 1}},
		"ttl":       InfoCommand{Proc: AutoCommit(TTL), Cons: Constraint{2, flags("rF"), 1, 1, 1}},
		"pttl":      InfoCommand{Proc: AutoCommit(PTTL), Cons: Constraint{2, flags("rF"), 1, 1, 1}},
		"object":    InfoCommand{Proc: AutoCommit(Object), Cons: Constraint{-2, flags("rR"), 0, 0, 0}},
		"scan":      InfoCommand{Proc: AutoCommit(Scan), Cons: Constraint{-2, flags("rR"), 0, 0, 0}},
		"randomkey": InfoCommand{Proc: AutoCommit(RandomKey), Cons: Constraint{1, flags("rR"), 0, 0, 0}},

		// server
		"monitor":  InfoCommand{Proc: Monitor, Cons: Constraint{1, flags("as"), 0, 0, 0}},
		"client":   InfoCommand{Proc: Client, Cons: Constraint{-2, flags("as"), 0, 0, 0}},
		"debug":    InfoCommand{Proc: AutoCommit(Debug), Cons: Constraint{-2, flags("as"), 0, 0, 0}},
		"command":  InfoCommand{Proc: RCommand, Cons: Constraint{0, flags("lt"), 0, 0, 0}},
		"flushdb":  InfoCommand{Proc: AutoCommit(FlushDB), Cons: Constraint{-1, flags("w"), 0, 0, 0}},
		"flushall": InfoCommand{Proc: AutoCommit(FlushAll), Cons: Constraint{-1, flags("w"), 0, 0, 0}},
		"time":     InfoCommand{Proc: Time, Cons: Constraint{1, flags("RF"), 0, 0, 0}},
		"info":     InfoCommand{Proc: Info, Cons: Constraint{-1, flags("lt"), 0, 0, 0}},

		// hashes
		"hdel":         InfoCommand{Proc: AutoCommit(HDel), Cons: Constraint{-3, flags("wF"), 1, 1, 1}},
		"hset":         InfoCommand{Proc: AutoCommit(HSet), Cons: Constraint{-4, flags("wmF"), 1, 1, 1}},
		"hget":         InfoCommand{Proc: AutoCommit(HGet), Cons: Constraint{3, flags("rF"), 1, 1, 1}},
		"hgetall":      InfoCommand{Proc: AutoCommit(HGetAll), Cons: Constraint{2, flags("r"), 1, 1, 1}},
		"hexists":      InfoCommand{Proc: AutoCommit(HExists), Cons: Constraint{3, flags("rF"), 1, 1, 1}},
		"hincrby":      InfoCommand{Proc: AutoCommit(HIncrBy), Cons: Constraint{4, flags("wmF"), 1, 1, 1}},
		"hincrbyfloat": InfoCommand{Proc: AutoCommit(HIncrByFloat), Cons: Constraint{4, flags("wmF"), 1, 1, 1}},
		"hkeys":        InfoCommand{Proc: AutoCommit(HKeys), Cons: Constraint{2, flags("rS"), 1, 1, 1}},
		"hvals":        InfoCommand{Proc: AutoCommit(HVals), Cons: Constraint{2, flags("rS"), 1, 1, 1}},
		"hlen":         InfoCommand{Proc: AutoCommit(HLen), Cons: Constraint{2, flags("rF"), 1, 1, 1}},
		"hstrlen":      InfoCommand{Proc: AutoCommit(HStrLen), Cons: Constraint{3, flags("rF"), 1, 1, 1}},
		"hsetnx":       InfoCommand{Proc: AutoCommit(HSetNX), Cons: Constraint{4, flags("wmF"), 1, 1, 1}},
		"hmget":        InfoCommand{Proc: AutoCommit(HMGet), Cons: Constraint{-3, flags("rF"), 1, 1, 1}},
		"hmset":        InfoCommand{Proc: AutoCommit(HMSet), Cons: Constraint{-3, flags("wmF"), 1, 1, 1}},

		// sets
		"sadd":     InfoCommand{Proc: AutoCommit(SAdd), Cons: Constraint{-3, flags("wmF"), 1, 1, 1}},
		"smembers": InfoCommand{Proc: AutoCommit(SMembers), Cons: Constraint{2, flags("rS"), 1, 1, 1}},
	}
}
