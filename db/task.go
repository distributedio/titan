package db

import (
	"context"
	"net/url"
	"strings"

	"github.com/distributedio/titan/conf"
	"github.com/distributedio/titan/db/store"
	"github.com/distributedio/titan/metrics"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/clientv3/concurrency"
	"go.uber.org/zap"
)

var (
	etcdPrefix = []byte("/titan:")
)

func RegisterTask(db *DB, conf *conf.TiKV) error {
	var (
		task_pool     *TaskPool
		register_list []TaskRegister
		err           error
	)
	if task_pool, err = NewTaskPool(db, conf); err != nil {
		return err
	}
	if !conf.GC.Disable {
		register_list = append(register_list, RegisterGCTask())
	}
	if !conf.Expire.Disable {
		register_list = append(register_list, RegisterExpireTask())
	}
	if !conf.TiKVGC.Disable {
		register_list = append(register_list, RegisterTikvGCTask())
	}
	if !conf.ZT.Disable {
		register_list = append(register_list, RegisterZT())
	}
	if len(register_list) == 0 {
		return nil
	}

	if err = task_pool.Regist(register_list...); err != nil {
		return err
	}
	task_pool.Start()
	return nil
}

func NewEtcdClient(addrs []string) (*clientv3.Client, error) {
	cli, err := clientv3.New(clientv3.Config{Endpoints: addrs})
	if err != nil {
		return nil, err
	}
	if logEnv := zap.L().Check(zap.DebugLevel, "new etcd client"); logEnv != nil {
		logEnv.Write(zap.Strings("endpoints", addrs))
	}
	return cli, nil
}

func PdAddrsToEtcd(pd string) []string {
	var etcdAddrs []string
	if strings.Contains(pd, store.MockAddr) {
		return etcdAddrs
	}
	u, err := url.Parse(pd)
	if err != nil {
		zap.L().Error("parase pd addrs to etcd err", zap.String("addrs", pd), zap.Error(err))
		return etcdAddrs
	}
	for _, v := range strings.Split(u.Host, ",") {
		etcdAddrs = append(etcdAddrs, "http://"+v)
	}
	return etcdAddrs
}

type TaskRegister func(db *DB, cli *clientv3.Client, conf *conf.TiKV) (*Task, error)
type TaskProc func(task *Task)

type TaskPool struct {
	c    *clientv3.Client
	db   *DB
	conf *conf.TiKV
	list []*Task
}

func NewTaskPool(db *DB, conf *conf.TiKV) (*TaskPool, error) {
	task_pool := &TaskPool{
		db:   db,
		conf: conf,
	}
	etcdAddrs := conf.EtcdAddrs
	if len(etcdAddrs) == 0 {
		etcdAddrs = PdAddrsToEtcd(conf.PdAddrs)
	}
	etcdClient, err := NewEtcdClient(etcdAddrs)
	if err != nil {
		return nil, err
	}
	task_pool.c = etcdClient

	return task_pool, nil
}

func (tp *TaskPool) Regist(registers ...TaskRegister) error {
	var err error
	var task *Task
	for _, register := range registers {
		if task, err = register(tp.db, tp.c, tp.conf); err == nil {
			tp.list = append(tp.list, task)
		}
	}
	return err
}

func (tp *TaskPool) Start() {
	for i, _ := range tp.list {
		task := tp.list[i]
		go func(task *Task) {
			for {
				if !task.Campaign() {
					continue
				}
				task.proc(task)
				metrics.GetMetrics().IsLeaderGaugeVec.WithLabelValues(task.lable).Set(0)
			}
		}(task)
	}
}

func RegisterZT() TaskRegister {
	return func(db *DB, cli *clientv3.Client, conf *conf.TiKV) (*Task, error) {
		return NewTask(db, cli, sysZTLeader, sysZTLeaderFlushInterval, conf.ZT, StartZT, "ZT")
	}
}

func RegisterExpireTask() TaskRegister {
	return func(db *DB, cli *clientv3.Client, conf *conf.TiKV) (*Task, error) {
		return NewTask(db, cli, sysExpireLeader, conf.Expire.LeaderTTL, conf.Expire, StartExpire, "EX")
	}
}

func RegisterTikvGCTask() TaskRegister {
	return func(db *DB, cli *clientv3.Client, conf *conf.TiKV) (*Task, error) {
		return NewTask(db, cli, sysTiKVGCLeader, conf.TiKVGC.LeaderTTL, conf.TiKVGC, StartTiKVGC, "TGC")
	}
}

func RegisterGCTask() TaskRegister {
	return func(db *DB, cli *clientv3.Client, conf *conf.TiKV) (*Task, error) {
		return NewTask(db, cli, sysGCLeader, conf.GC.LeaderTTL, conf.GC, StartGC, "GC")
	}
}

func NewTask(db *DB, cli *clientv3.Client, key []byte, ttl int, conf interface{}, proc TaskProc, lable string) (*Task, error) {
	uuid := UUID()
	t := &Task{
		db:    db,
		id:    uuid,
		key:   key,
		conf:  conf,
		proc:  proc,
		lable: lable,
	}
	session, err := concurrency.NewSession(cli, concurrency.WithTTL(ttl))
	if err != nil {
		zap.L().Error("create task err", zap.String("lable", lable), zap.Error(err))
		return nil, err
	}
	t.Session = session

	zap.L().Info("create task", zap.String("lable", lable))
	return t, nil
}

type Task struct {
	*concurrency.Session
	db    *DB
	id    []byte
	key   []byte
	conf  interface{}
	proc  TaskProc
	lable string
}

func (t *Task) Campaign() bool {
	key := append(etcdPrefix, t.key...)
	elec := concurrency.NewElection(t.Session, string(key))

	if err := elec.Campaign(context.Background(), string(t.id)); err != nil {
		zap.L().Error("elect campaign err", zap.Error(err))
		return false
	}

	if logEnv := zap.L().Check(zap.DebugLevel, "Elect leader success"); logEnv != nil {
		logEnv.Write(zap.ByteString("key", key), zap.ByteString("id", t.id), zap.String("lable", t.lable))
	}
	metrics.GetMetrics().IsLeaderGaugeVec.WithLabelValues(t.lable).Set(1)
	return true
}
