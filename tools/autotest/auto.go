package autotest

import (
	"strconv"
	"testing"
	"time"

	"gitlab.meitu.com/platform/thanos/tools/autotest/cmd"

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
	_, err = redis.String(conn.Do("auth", "test-1541501672-1-98d9882bb7a8ba2c16974e"))
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
//TODO
func (ac *AutoClient) StringCase(t *testing.T) {
	ac.es.SetEqual(t, "key-set", "value")
	ac.es.AppendEqual(t, "key-set", "value")
	ac.es.AppendEqual(t, "append", "value")
	ac.es.StrlenEqual(t, "key-set")
	ac.es.MSetEqual(t, "key-set", "value")
	ac.es.MGetEqual(t, "key-not-exist")
	ac.es.IncrEqual(t, "incr")
	ac.es.IncrEqual(t, "incr")
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

	var key []string
	for i := 0; i < 4000; i++ {
		num := strconv.Itoa(i)
		key = append(key, "v", num)
	}
	ac.el.LpushEqual(t, "zkey-list")
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
	ac.ek.ScanEqual(t, "", 3)
	ac.ek.ExistsEqual(t, 3, "key-set", "incr", "foo", "append")
	ac.ek.DelEqual(t, 3, "key-set", "incr", "foo", "append")
	ac.ek.TTLEqual(t, "key-set", -2)
	ac.ek.RandomKeyEqual(t)
	ac.ek.ScanEqual(t, "", 0)

	ac.es.SetEqual(t, "key-set", "value")
	ac.ek.TypeEqual(t, "ket-set", "string")
	ac.ek.ObjectEqual(t, "key-set", "embstr")
	ac.ek.ExpireEqual(t, "key-set", 2, 1)
	ac.ek.TTLEqual(t, "key-set", 1)
	time.Sleep(time.Second * 2)
	ac.ek.ExpireEqual(t, "key-set", 1, 0)
	ac.ek.ExpireEqual(t, "key-set", 0, 0)

	//test PExpire
	ac.el.LpushEqual(t, "key-set", "value")
	ac.ek.TypeEqual(t, "ket-set", "list")
	ac.ek.PExpireEqual(t, "key-set", 2000, 1)
	ac.ek.TTLEqual(t, "key-set", 1)
	time.Sleep(time.Second * 2)
	ac.ek.PExpireEqual(t, "key-set", 1, 0)
	ac.ek.PExpireEqual(t, "key-set", 0, 0)

	at := time.Now().Unix() + int64(2*time.Second)
	var key []string
	for i := 0; i < 4000; i++ {
		num := strconv.Itoa(i)
		key = append(key, "v", num)
	}
	ac.el.LpushEqual(t, "zkey-list")
	ac.ek.TypeEqual(t, "ket-set", "list")
	ac.ek.ObjectEqual(t, "key-set", "quicklist")
	ac.ek.ExpireAtEqual(t, "zkey-list", int(at), 1)
	time.Sleep(time.Second * 2)
	ac.ek.ExpireAtEqual(t, "zkey-list", int(at), 0)
	ac.ek.ExpireAtEqual(t, "zkey-list", int(at), 0)

	//test PExpire
	at = time.Now().UnixNano()/1000 + int64(2*time.Second)
	ac.el.LpushEqual(t, "key-set", "value")
	ac.ek.TypeEqual(t, "ket-set", "list")
	ac.ek.ObjectEqual(t, "key-set", "ziplist")
	ac.ek.PExpireAtEqual(t, "key-set", 2000, 1)
	ac.ek.TTLEqual(t, "key-set", 1)
	time.Sleep(time.Second * 2)
	ac.ek.PExpireAtEqual(t, "key-set", 1, 0)
	ac.ek.PExpireAtEqual(t, "key-set", 0, 0)
}

//SystemCase check system case
func (ac *AutoClient) SystemCase(t *testing.T) {
	//auth
	ac.AuthEqual(t, "test-1541501672-1-98d9882bb7a8ba2c16974e")
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
