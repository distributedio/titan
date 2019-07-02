package autotest

import (
	"testing"

	"github.com/distributedio/titan/tools/autotest/cmd"
	"github.com/gomodule/redigo/redis"
)

//Abnormal check error message
type Abnormal struct {
	es   *cmd.ExampleString
	el   *cmd.ExampleList
	ek   *cmd.ExampleKey
	ez   *cmd.ExampleZSet
	ess  *cmd.ExampleSystem
	em   *cmd.ExampleMulti
	conn redis.Conn
}

//NewAbnormal create object
func NewAbnormal() *Abnormal {
	return &Abnormal{}
}

//Start  create abnormal client
func (an *Abnormal) Start(addr string) {
	conn, err := redis.Dial("tcp", addr)
	if err != nil {
		panic(err)
	}
	_, err = redis.String(conn.Do("auth", "test-1542098935-1-7ca41bda4efc2a1889c04e"))
	if err != nil {
		panic(err)
	}
	an.conn = conn
	an.es = cmd.NewExampleString(conn)
	an.ek = cmd.NewExampleKey(conn)
	an.el = cmd.NewExampleList(conn)
	an.ez = cmd.NewExampleZSet(conn)
	an.ess = cmd.NewExampleSystem(conn)
	an.em = cmd.NewExampleMulti(conn)
}

//Close close annormal client
func (an *Abnormal) Close() {
	an.conn.Close()
}

//StringCase check string case
func (an *Abnormal) StringCase(t *testing.T) {
	an.el.LpushEqual(t, "lpush", "key")
	//set
	an.es.SetEqualErr(t, "ERR wrong number of arguments for 'set' command", "fuck")
	an.es.SetEqualErr(t, "ERR value is not an integer or out of range", "key", "v", "ex", "second")
	an.es.SetEqualErr(t, "ERR syntax error", "key", "v", "nx", "second")

	an.es.GetEqualErr(t, "ERR wrong number of arguments for 'get' command", "hello", "fuck")
	an.es.GetEqualErr(t, "WRONGTYPE Operation against a key holding the wrong kind of value", "lpush")

	an.es.MSetEqualErr(t, "ERR wrong number of arguments for 'mset' command")
	an.es.MSetEqualErr(t, "ERR wrong number of arguments for 'mset' command", "key")

	an.es.MGetEqualErr(t, "ERR wrong number of arguments for 'mget' command")
	an.es.MGetEqual(t, "lpush")

	an.es.AppendEqualErr(t, "ERR wrong number of arguments for 'append' command", "he", "he", "he")
	an.es.AppendEqualErr(t, "ERR wrong number of arguments for 'append' command", "he")
	an.es.AppendEqualErr(t, "WRONGTYPE Operation against a key holding the wrong kind of value", "lpush", "hehe")

	an.es.IncrEqualErr(t, "ERR wrong number of arguments for 'incr' command", "1", "m")
	an.es.IncrEqualErr(t, "WRONGTYPE Operation against a key holding the wrong kind of value", "lpush")

	an.es.StrlenEqualErr(t, "ERR wrong number of arguments for 'strlen' command", "heng", "heng")
	an.es.StrlenEqualErr(t, "WRONGTYPE Operation against a key holding the wrong kind of value", "lpush")

	an.el.LpopEqual(t, "lpush")
}

//ListCase check list case
func (an *Abnormal) ListCase(t *testing.T) {

	an.es.SetEqual(t, "set", "v")
	an.el.LpushEqual(t, "lpush", "key")

	an.el.LlenEqualErr(t, "ERR wrong number of arguments for 'llen' command", "fuck", "z")
	an.el.LlenEqualErr(t, "WRONGTYPE Operation against a key holding the wrong kind of value", "set")

	an.el.LpopEqualErr(t, "ERR wrong number of arguments for 'lpop' command", "hello", "fuck")
	an.el.LpopEqualErr(t, "WRONGTYPE Operation against a key holding the wrong kind of value", "set")

	an.el.LpushEqualErr(t, "ERR wrong number of arguments for 'lpush' command", "z")
	an.el.LpushEqualErr(t, "WRONGTYPE Operation against a key holding the wrong kind of value", "set", "key")

	an.el.LindexEqualErr(t, "ERR wrong number of arguments for 'lindex' command", "z", "z", "z")
	an.el.LindexEqualErr(t, "WRONGTYPE Operation against a key holding the wrong kind of value", "set", 1)

	an.el.LrangeEqualErr(t, "ERR wrong number of arguments for 'lrange' command", "he", "he", "he", "he")
	an.el.LrangeEqualErr(t, "WRONGTYPE Operation against a key holding the wrong kind of value", "set", "1", "1")
	an.el.LrangeEqualErr(t, "ERR value is not an integer or out of range", "setx", "z", "1")
	an.el.LrangeEqualErr(t, "ERR value is not an integer or out of range", "setx", "1", "z")

	an.el.LsetEqualErr(t, "ERR wrong number of arguments for 'lset' command", "1", "h")
	an.el.LsetEqualErr(t, "WRONGTYPE Operation against a key holding the wrong kind of value", "set", "1", "z")
	an.el.LsetEqualErr(t, "ERR no such key", "setx", "1", "z")
	an.el.LsetEqualErr(t, "ERR value is not an integer or out of range", "lpush", "x", "z")
	an.el.LsetEqualErr(t, "ERR index out of range", "lpush", "-100", "z")

	// an.el.RpopEqualErr(t, "ERR wrong number of arguments for 'rpop' command", "heng", "heng")
	// an.el.RpopEqualErr(t, "WRONGTYPE Operation against a key holding the wrong kind of value", "set")

	an.el.RpushEqualErr(t, "ERR wrong number of arguments for 'rpush' command", "heng")
	an.el.RpushEqualErr(t, "WRONGTYPE Operation against a key holding the wrong kind of value", "set", "k")

	an.el.LpopEqual(t, "lpush")
	an.ek.DelEqual(t, 1, "set")
}

//ZSetCase check zset case
func (an *Abnormal) ZSetCase(t *testing.T) {
	an.es.SetEqual(t, "set", "v")
	an.ez.ZAddEqual(t, "key-zset-abnormal", "1.2", "member1")

	an.ez.ZCardEqualErr(t, "ERR wrong number of arguments for 'zcard' command", "set", "v")
	an.ez.ZCardEqualErr(t, "WRONGTYPE Operation against a key holding the wrong kind of value", "set")

	an.ez.ZAddEqualErr(t, "ERR wrong number of arguments for 'zadd' command", "set")
	an.ez.ZAddEqualErr(t, "ERR wrong number of arguments for 'zadd' command", "set", "v")
	an.ez.ZAddEqualErr(t, "ERR wrong number of arguments for 'zadd' command", "set", "v", "m1", "v2")
	an.ez.ZAddEqualErr(t, "WRONGTYPE Operation against a key holding the wrong kind of value", "set", "1", "m1")
	an.ez.ZAddEqualErr(t, "ERR value is not a valid float", "key-zset-abnormal", "v", "m1")
	an.ez.ZAddEqualErr(t, "ERR value is not a valid float", "key-zset-abnormal", "nan", "m1")

	an.ez.ZRangeEqualErr(t, "ERR wrong number of arguments for 'zrange' command", "set")
	an.ez.ZRangeEqualErr(t, "ERR wrong number of arguments for 'zrange' command", "set", "0")
	an.ez.ZRangeEqualErr(t, "WRONGTYPE Operation against a key holding the wrong kind of value", "set", "0", "1")
	an.ez.ZRangeEqualErr(t, "ERR value is not an integer or out of range", "key-zset-abnormal", "0", "a")
	an.ez.ZRangeEqualErr(t, "ERR value is not an integer or out of range", "key-zset-abnormal", "a", "0")
	an.ez.ZRangeEqualErr(t, "ERR value is not an integer or out of range", "key-zset-abnormal", "0", "9223372036854775808")
	an.ez.ZRangeEqualErr(t, "ERR value is not an integer or out of range", "key-zset-abnormal", "9223372036854775808", "0")

	an.ez.ZRevRangeEqualErr(t, "ERR wrong number of arguments for 'zrevrange' command", "set")
	an.ez.ZRevRangeEqualErr(t, "ERR wrong number of arguments for 'zrevrange' command", "set", "0")
	an.ez.ZRevRangeEqualErr(t, "WRONGTYPE Operation against a key holding the wrong kind of value", "set", "0", "1")
	an.ez.ZRevRangeEqualErr(t, "ERR value is not an integer or out of range", "key-zset-abnormal", "0", "a")
	an.ez.ZRevRangeEqualErr(t, "ERR value is not an integer or out of range", "key-zset-abnormal", "a", "0")
	an.ez.ZRevRangeEqualErr(t, "ERR value is not an integer or out of range", "key-zset-abnormal", "0", "9223372036854775808")
	an.ez.ZRevRangeEqualErr(t, "ERR value is not an integer or out of range", "key-zset-abnormal", "9223372036854775808", "0")

	an.ez.ZScoreEqualErr(t, "ERR wrong number of arguments for 'zscore' command", "set")
	an.ez.ZScoreEqualErr(t, "ERR wrong number of arguments for 'zscore' command", "set", "m1", "m2")
	an.ez.ZScoreEqualErr(t, "WRONGTYPE Operation against a key holding the wrong kind of value", "set", "m1")

	an.ez.ZRemEqualErr(t, "ERR wrong number of arguments for 'zrem' command", "set")
	an.ez.ZRemEqualErr(t, "WRONGTYPE Operation against a key holding the wrong kind of value", "set", "m1")

	an.ek.DelEqual(t, 1, "key-zset-abnormal")
	an.ek.DelEqual(t, 1, "set")
}

//KeyCase check key case
func (an *Abnormal) KeyCase(t *testing.T) {
	an.ek.DelEqualErr(t, "ERR wrong number of arguments for 'del' command")

	an.ek.ExistsEqualErr(t, "ERR wrong number of arguments for 'exists' command")

	an.ek.ExpireEqualErr(t, "ERR wrong number of arguments for 'expire' command")
	an.ek.ExpireEqualErr(t, "ERR value is not an integer or out of range", "key", "z")

	// an.InfoEqualErr(t, "ERR wrong number of arguments for 'INFO' command")

	an.ek.RandomKeyEqualErr(t, "ERR wrong number of arguments for 'randomkey' command", "", "", "")

	an.ek.ScanEqualErr(t, "ERR wrong number of arguments for 'scan' command")

	an.ek.TTLEqualErr(t, "ERR wrong number of arguments for 'ttl' command", "heng", "heng")
}

//SystemCase check system case
func (an *Abnormal) SystemCase(t *testing.T) {
	an.ess.PingEqualErr(t, "ERR wrong number of arguments for 'ping' command", "ping", "hello", "fuck")
}

//MultiCase check multi case
func (an *Abnormal) MultiCase(t *testing.T) {

	an.em.MultiEqualErr(t, "ERR wrong number of arguments for 'multi' command", "he", "he")
	an.em.ExecEqualErr(t, "ERR wrong number of arguments for 'exec' command", "he", "he")
	an.em.ExecEqualErr(t, "ERR EXEC without MULTI")
	an.em.MultiEqual(t)
	an.em.MultiEqualErr(t, "ERR MULTI calls can not be nested")
	an.em.ExecEqualErr(t, "")
}
