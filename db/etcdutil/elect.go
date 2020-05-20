package etcdutil

import (
	"context"
	"sync"

	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/clientv3/concurrency"
	"go.uber.org/zap"
)

type Elect struct {
	rw     sync.RWMutex
	leader bool
	c      *clientv3.Client
	ctx    context.Context
	key    []byte
	val    []byte
	ttl    int
}

func RegisterElect(ctx context.Context, cli *clientv3.Client, key, val []byte, ttl int) *Elect {
	e := &Elect{
		c:   cli,
		ctx: ctx,
		key: key,
		val: val,
		ttl: ttl,
	}
	go e.Campaign()
	return e
}

func (e *Elect) Campaign() {
	for {
		s, err := concurrency.NewSession(e.c, concurrency.WithTTL(e.ttl))
		if err != nil {
			zap.L().Error("elect create session err", zap.Error(err))
			continue
		}
		key := append(etcdPrefix, e.key...)
		elec := concurrency.NewElection(s, string(key))

		if err = elec.Campaign(e.ctx, string(e.val)); err != nil {
			zap.L().Error("elect campaign err", zap.Error(err))
			continue
		}
		if logEnv := zap.L().Check(zap.DebugLevel, "Elect leader success"); logEnv != nil {
			logEnv.Write(zap.ByteString("key", e.key), zap.Int("ttl", e.ttl))
		}

		e.setLeader(true)
		select {
		case <-s.Done():
			e.setLeader(false)
			if logEnv := zap.L().Check(zap.DebugLevel, "Elect session done"); logEnv != nil {
				logEnv.Write(zap.ByteString("key", e.key), zap.Int("ttl", e.ttl))
			}
		}
	}
}

func (e *Elect) setLeader(leader bool) {
	e.rw.Lock()
	defer e.rw.Unlock()
	e.leader = leader
}

func (e *Elect) IsLeader() bool {
	e.rw.RLock()
	defer e.rw.RUnlock()
	return e.leader
}

func (e *Elect) Key() []byte {
	return e.key
}

func (e *Elect) Val() []byte {
	return e.val
}
