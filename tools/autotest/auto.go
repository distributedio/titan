package autotest

import (
	"strconv"
	"testing"
	"time"

	"github.com/meitu/titan/tools/autotest/cmd"

	"github.com/gomodule/redigo/redis"
)

//AutoClient check redis comman
type AutoClient struct {
	es *cmd.ExampleString
	el *cmd.ExampleList
	ek *cmd.ExampleKey
	*cmd.ExampleSystem
	em *cmd.ExampleMulti
	// addr string
	pool *redis.Pool
	conn redis.Conn
}

//NewAutoClient creat auto client
func NewAutoClient() *AutoClient {
	return &AutoClient{}
}

//Start run client
func (ac *AutoClient) Start(addr string) {
	// ac.pool = newPool(addr)
	conn, err := redis.Dial("tcp", addr)
	if err != nil {
		panic(err)
	}
	_, err = redis.String(conn.Do("auth", "test-1542098935-1-7ca41bda4efc2a1889c04e"))
	if err != nil {
		panic(err)
	}
	ac.conn = conn
	ac.es = cmd.NewExampleString(conn)
	ac.ek = cmd.NewExampleKey(conn)
	ac.el = cmd.NewExampleList(conn)
	ac.ExampleSystem = cmd.NewExampleSystem(conn)
	ac.em = cmd.NewExampleMulti(conn)
}

//Close shut client
func (ac *AutoClient) Close() {
	// ac.pool.Close()
	ac.conn.Close()
}

//StringCase check string case
func (ac *AutoClient) StringCase(t *testing.T) {
	ac.es.SetNxEqual(t, "key-setx", "v1")
	ac.es.SetExEqual(t, "key-set", "v2", 1)
	ac.es.PSetexEqual(t, "key-set", "v3", 3000)

	ac.es.SetEqual(t, "key-set", "value")
	ac.es.AppendEqual(t, "key-set", "value")
	ac.es.AppendEqual(t, "append", "value")
	ac.es.StrlenEqual(t, "key-set")
	// ac.es.MSetNxEqual(t, 1, "key-setm", "value", "key-set", "value")

	ac.es.MSetEqual(t, "key-set", "value")
	ac.es.MGetEqual(t, "key-not-exist")
	ac.es.IncrEqual(t, "incr")
	ac.es.IncrEqual(t, "incr")
	ac.es.IncrByEqual(t, "incr", 19)
	ac.es.DecrEqual(t, "incr")
	ac.es.DecrByEqual(t, "incr", 19)
	ac.es.IncrByFloatEqual(t, "incr", 1.111111)
	ac.es.IncrByFloatEqual(t, "incr", 2.1e2)
	ac.es.StrlenEqual(t, "heng")

}

//ListCase check list case
//TODO
func (ac *AutoClient) ListCase(t *testing.T) {
	ac.el.LpushEqual(t, "key-list", "v1", "v2", "v3", "v4")
	ac.el.LlenEqual(t, "key-list")
	ac.el.LsetEqual(t, "key-list", 3, "v0")
	ac.el.LindexEqual(t, "key-list", 3)
	ac.el.LrangeEqual(t, "key-list", 0, 10)
	ac.el.LrangeEqual(t, "key-list", 99, 100)
	ac.el.LpopEqual(t, "key-list")
	ac.el.LpopEqual(t, "key-list-l")

	var keys []string
	for i := 0; i < 4000; i++ {
		num := strconv.Itoa(i)
		keys = append(keys, "v", num)
	}
	ac.el.LpushEqual(t, "zkey-list", keys...)
	ac.el.LlenEqual(t, "zkey-list")
	ac.el.LsetEqual(t, "zkey-list", 3, "v0")
	ac.el.LindexEqual(t, "zkey-list", 3)
	ac.el.LrangeEqual(t, "zkey-list", 0, 10)
	ac.el.LrangeEqual(t, "zkey-list", 99, 100)
	ac.el.LpopEqual(t, "zkey-list")
}

//KeyCase check key case
//TODO
func (ac *AutoClient) KeyCase(t *testing.T) {
	ac.ek.TTLEqual(t, "key-set", -1)
	ac.ek.RandomKeyEqual(t)
	ac.ek.ScanEqual(t, "", 4)
	ac.ek.ExistsEqual(t, 3, "key-set", "incr", "foo", "append")
	ac.ek.DelEqual(t, 4, "key-set", "incr", "foo", "append", "key-setx")
	ac.ek.TTLEqual(t, "key-set", -2)
	ac.ek.RandomKeyEqual(t)
	ac.ek.ScanEqual(t, "", 0)

	ac.es.SetEqual(t, "key-set", "value")
	ac.ek.TypeEqual(t, "key-set", "string")
	ac.ek.ObjectEqual(t, "key-set", "raw")
	ac.ek.ExpireEqual(t, "key-set", 2, 1)
	ac.ek.TTLEqual(t, "key-set", 1)
	time.Sleep(time.Second * 2)
	ac.ek.ExpireEqual(t, "key-set", 1, 0)
	ac.ek.ExpireEqual(t, "key-set", 0, 0)

	//test PExpire
	ac.el.LpushEqual(t, "key-set", "value")
	ac.ek.TypeEqual(t, "key-set", "list")
	ac.ek.PExpireEqual(t, "key-set", 2000, 1)
	ac.ek.TTLEqual(t, "key-set", 1)
	time.Sleep(time.Second * 2)
	ac.ek.PExpireEqual(t, "key-set", 1, 0)
	ac.ek.PExpireEqual(t, "key-set", 0, 0)

	at := time.Now().Unix() + 1
	var key []string
	for i := 0; i < 4000; i++ {
		num := strconv.Itoa(i)
		key = append(key, "v", num)
	}
	ac.el.LpushEqual(t, "zkey-listx", key...)
	ac.ek.TypeEqual(t, "zkey-listx", "list")
	ac.ek.ObjectEqual(t, "zkey-listx", "ziplist")
	ac.ek.ExpireAtEqual(t, "zkey-listx", int(at), 1)
	time.Sleep(time.Second * 1)
	ac.ek.ExpireAtEqual(t, "zkey-listx", int(at), 0)

	//test PExpire
	at = (time.Now().Unix() + 1) * 1000
	ac.el.LpushEqual(t, "key-setz", "value")
	ac.ek.TypeEqual(t, "key-setz", "list")
	ac.ek.ObjectEqual(t, "key-setz", "linkedlist")
	ac.ek.PExpireAtEqual(t, "key-setz", int(at), 1)
	ac.ek.TTLEqual(t, "key-setz", 0)
	time.Sleep(time.Second * 2)
	ac.ek.PExpireAtEqual(t, "key-setz", int(at), 0)

	at = time.Now().Unix() + 1
	ac.es.SetEqual(t, "zkey-listx", "value")
	ac.ek.ExpireAtEqual(t, "zkey-listx", int(at), 1)
	ac.ek.PersistEqual(t, "zkey-listx", 1)
	ac.ek.PersistEqual(t, "zkey-listx", 0)
}

//SystemCase check system case
func (ac *AutoClient) SystemCase(t *testing.T) {
	//auth
	ac.AuthEqual(t, "test-1542098935-1-7ca41bda4efc2a1889c04e")
	//ping
	ac.PingEqual(t)
}

//MultiCase bug mutil exec repley msg is error
func (ac *AutoClient) MultiCase(t *testing.T) {
	//multi
	ac.em.MultiEqual(t)
	ac.em.Cmd(t)
	// exec
	ac.em.ExecEqual(t)
}
