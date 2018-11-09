package command

var txnCommands map[string]TxnCommand
var commands map[string]CommandInfo

func init() {
	// txnCommands will be searched in multi/exec
	txnCommands = map[string]TxnCommand{
		// lists
		"lpush":   LPush,
		"lpop":    LPop,
		"lrange":  LRange,
		"linsert": LInsert,

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
	commands = map[string]CommandInfo{
		// connections
		"auth":   CommandInfo{Proc: Auth, Cons: Constraint{2, flags("sltF"), 0, 0, 0}},
		"echo":   CommandInfo{Proc: Echo, Cons: Constraint{2, flags("F"), 0, 0, 0}},
		"ping":   CommandInfo{Proc: Ping, Cons: Constraint{-1, flags("tF"), 0, 0, 0}},
		"quit":   CommandInfo{Proc: Quit, Cons: Constraint{1, 0, 0, 0, 0}},
		"select": CommandInfo{Proc: Select, Cons: Constraint{2, flags("lF"), 0, 0, 0}},
		"swapdb": CommandInfo{Proc: SwapDB, Cons: Constraint{3, flags("wF"), 0, 0, 0}},

		// transactions, exec and discard should called explicitly, so they are registered here
		"multi":   CommandInfo{Proc: Multi, Cons: Constraint{1, flags("sF"), 0, 0, 0}},
		"watch":   CommandInfo{Proc: Watch, Cons: Constraint{-2, flags("sF"), 1, -1, 1}},
		"unwatch": CommandInfo{Proc: Unwatch, Cons: Constraint{1, flags("sF"), 0, 0, 0}},

		// lists
		"lpush":   CommandInfo{Proc: AutoCommit(LPush), Cons: Constraint{-3, flags("wmF"), 1, 1, 1}},
		"lpop":    CommandInfo{Proc: AutoCommit(LPop), Cons: Constraint{2, flags("wF"), 1, 1, 1}},
		"lrange":  CommandInfo{Proc: AutoCommit(LRange), Cons: Constraint{4, flags("r"), 1, 1, 1}},
		"linsert": CommandInfo{Proc: AutoCommit(LInsert), Cons: Constraint{5, flags("wm"), 1, 1, 1}},

		// strings
		"get":    CommandInfo{Proc: AutoCommit(Get), Cons: Constraint{2, flags("rF"), 1, 1, 1}},
		"set":    CommandInfo{Proc: AutoCommit(Set), Cons: Constraint{-3, flags("wm"), 1, 1, 1}},
		"setnx":  CommandInfo{Proc: AutoCommit(SetNx), Cons: Constraint{3, flags("wmF"), 1, 1, 1}},
		"setex":  CommandInfo{Proc: AutoCommit(SetEx), Cons: Constraint{4, flags("wm"), 1, 1, 1}},
		"psetex": CommandInfo{Proc: AutoCommit(PSetEx), Cons: Constraint{4, flags("wm"), 1, 1, 1}},
		"mget":   CommandInfo{Proc: AutoCommit(MGet), Cons: Constraint{-2, flags("rF"), 1, -1, 1}},
		"mset":   CommandInfo{Proc: AutoCommit(MSet), Cons: Constraint{-3, flags("wm"), 1, -1, 2}},
		"msetnx": CommandInfo{Proc: AutoCommit(MSetNx), Cons: Constraint{-3, flags("wm"), 1, -1, 2}},
		"strlen": CommandInfo{Proc: AutoCommit(Strlen), Cons: Constraint{2, flags("rF"), 1, 1, 1}},
		"append": CommandInfo{Proc: AutoCommit(Append), Cons: Constraint{3, flags("wm"), 1, 1, 1}},
		// "setrange": CommandInfo{Proc: AutoCommit(SetRange), Cons: Constraint{4, flags("wm"), 1, 1, 1}},
		"getrange":    CommandInfo{Proc: AutoCommit(GetRange), Cons: Constraint{4, flags("r"), 1, 1, 1}},
		"incr":        CommandInfo{Proc: AutoCommit(Incr), Cons: Constraint{2, flags("wmF"), 1, 1, 1}},
		"decr":        CommandInfo{Proc: AutoCommit(Decr), Cons: Constraint{2, flags("wmF"), 1, 1, 1}},
		"incrby":      CommandInfo{Proc: AutoCommit(IncrBy), Cons: Constraint{3, flags("wmF"), 1, 1, 1}},
		"decrby":      CommandInfo{Proc: AutoCommit(DecrBy), Cons: Constraint{3, flags("wmF"), 1, 1, 1}},
		"incrbyfloat": CommandInfo{Proc: AutoCommit(IncrByFloat), Cons: Constraint{3, flags("wmF"), 1, 1, 1}},

		// keys
		"type":      CommandInfo{Proc: AutoCommit(Type), Cons: Constraint{2, flags("rF"), 1, 1, 1}},
		"exists":    CommandInfo{Proc: AutoCommit(Exists), Cons: Constraint{-2, flags("rF"), 1, -1, 1}},
		"keys":      CommandInfo{Proc: AutoCommit(Keys), Cons: Constraint{-2, flags("rS"), 0, 0, 0}},
		"del":       CommandInfo{Proc: AutoCommit(Delete), Cons: Constraint{-2, flags("w"), 1, -1, 1}},
		"expire":    CommandInfo{Proc: AutoCommit(Expire), Cons: Constraint{3, flags("wF"), 1, 1, 1}},
		"expireat":  CommandInfo{Proc: AutoCommit(ExpireAt), Cons: Constraint{3, flags("wF"), 1, 1, 1}},
		"pexpire":   CommandInfo{Proc: AutoCommit(PExpire), Cons: Constraint{3, flags("wF"), 1, 1, 1}},
		"pexpireat": CommandInfo{Proc: AutoCommit(PExpireAt), Cons: Constraint{3, flags("wF"), 1, 1, 1}},
		"persist":   CommandInfo{Proc: AutoCommit(Persist), Cons: Constraint{2, flags("wF"), 1, 1, 1}},
		"ttl":       CommandInfo{Proc: AutoCommit(TTL), Cons: Constraint{2, flags("rF"), 1, 1, 1}},
		"pttl":      CommandInfo{Proc: AutoCommit(PTTL), Cons: Constraint{2, flags("rF"), 1, 1, 1}},
		"object":    CommandInfo{Proc: AutoCommit(Object), Cons: Constraint{-2, flags("rR"), 0, 0, 0}},
		"scan":      CommandInfo{Proc: AutoCommit(Scan), Cons: Constraint{-2, flags("rR"), 0, 0, 0}},
		"randomkey": CommandInfo{Proc: AutoCommit(RandomKey), Cons: Constraint{1, flags("rR"), 0, 0, 0}},

		// server
		"monitor":  CommandInfo{Proc: Monitor, Cons: Constraint{1, flags("as"), 0, 0, 0}},
		"client":   CommandInfo{Proc: Client, Cons: Constraint{-2, flags("as"), 0, 0, 0}},
		"debug":    CommandInfo{Proc: AutoCommit(Debug), Cons: Constraint{-2, flags("as"), 0, 0, 0}},
		"command":  CommandInfo{Proc: CommandCommand, Cons: Constraint{0, flags("lt"), 0, 0, 0}},
		"flushdb":  CommandInfo{Proc: AutoCommit(FlushDB), Cons: Constraint{-1, flags("w"), 0, 0, 0}},
		"flushall": CommandInfo{Proc: AutoCommit(FlushAll), Cons: Constraint{-1, flags("w"), 0, 0, 0}},
		"time":     CommandInfo{Proc: Time, Cons: Constraint{1, flags("RF"), 0, 0, 0}},

		// hashes
		"hdel":         CommandInfo{Proc: AutoCommit(HDel), Cons: Constraint{-3, flags("wF"), 1, 1, 1}},
		"hset":         CommandInfo{Proc: AutoCommit(HSet), Cons: Constraint{-4, flags("wmF"), 1, 1, 1}},
		"hget":         CommandInfo{Proc: AutoCommit(HGet), Cons: Constraint{3, flags("rF"), 1, 1, 1}},
		"hgetall":      CommandInfo{Proc: AutoCommit(HGetAll), Cons: Constraint{2, flags("r"), 1, 1, 1}},
		"hexists":      CommandInfo{Proc: AutoCommit(HExists), Cons: Constraint{3, flags("rF"), 1, 1, 1}},
		"hincrby":      CommandInfo{Proc: AutoCommit(HIncrBy), Cons: Constraint{4, flags("wmF"), 1, 1, 1}},
		"hincrbyfloat": CommandInfo{Proc: AutoCommit(HIncrByFloat), Cons: Constraint{4, flags("wmF"), 1, 1, 1}},
		"hkeys":        CommandInfo{Proc: AutoCommit(HKeys), Cons: Constraint{2, flags("rS"), 1, 1, 1}},
		"hvals":        CommandInfo{Proc: AutoCommit(HVals), Cons: Constraint{2, flags("rS"), 1, 1, 1}},
		"hlen":         CommandInfo{Proc: AutoCommit(HLen), Cons: Constraint{2, flags("rF"), 1, 1, 1}},
		"hstrlen":      CommandInfo{Proc: AutoCommit(HStrLen), Cons: Constraint{3, flags("rF"), 1, 1, 1}},
		"hsetnx":       CommandInfo{Proc: AutoCommit(HSetNX), Cons: Constraint{4, flags("wmF"), 1, 1, 1}},
		"hmget":        CommandInfo{Proc: AutoCommit(HMGet), Cons: Constraint{-3, flags("rF"), 1, 1, 1}},
		"hmset":        CommandInfo{Proc: AutoCommit(HMSet), Cons: Constraint{-3, flags("wmF"), 1, 1, 1}},

		// sets
		"sadd":     CommandInfo{Proc: AutoCommit(SAdd), Cons: Constraint{-3, flags("wmF"), 1, 1, 1}},
		"smembers": CommandInfo{Proc: AutoCommit(SMembers), Cons: Constraint{2, flags("rS"), 1, 1, 1}},
	}
}
