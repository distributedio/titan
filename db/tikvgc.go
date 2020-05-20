package db

import (
	"context"
	"time"

	"github.com/distributedio/titan/conf"
	"github.com/distributedio/titan/db/etcdutil"
	"github.com/pingcap/tidb/store/tikv"
	"github.com/pingcap/tidb/store/tikv/gcworker"
	"github.com/pingcap/tidb/store/tikv/oracle"
	"go.etcd.io/etcd/clientv3"
	"go.uber.org/zap"
)

var (
	sysTiKVGCLeader        = []byte("$sys:0:TGC:GCLeader")
	sysTiKVGCLastSafePoint = []byte("$sys:0:TGC:LastSafePoint")
)

const (
	tikvGcTimeFormat = "20060102-15:04:05 -0700 MST"
)

// StartTiKVGC start tikv gcwork
func StartTiKVGC(db *DB, cli *clientv3.Client, tikvCfg *conf.TiKVGC) {
	ticker := time.NewTicker(tikvCfg.Interval)
	defer ticker.Stop()
	uuid := UUID()
	ctx := context.Background()
	e := etcdutil.RegisterElect(ctx, cli, sysTiKVGCLeader, uuid, tikvCfg.LeaderTTL)
	for range ticker.C {
		if tikvCfg.Disable {
			continue
		}
		if !isLeader(e) {
			if logEnv := zap.L().Check(zap.DebugLevel, "[TiKVGC]  not TiKVGC leader"); logEnv != nil {
				logEnv.Write(zap.ByteString("leader", sysTiKVGCLeader),
					zap.ByteString("uuid", uuid),
					zap.Int("leader-ttl", tikvCfg.LeaderTTL),
					zap.Duration("safe-point-life-time", tikvCfg.SafePointLifeTime),
					zap.Int("concurrency", tikvCfg.Concurrency))
			}
			continue
		}
		if err := runTiKVGC(ctx, db, uuid, tikvCfg.SafePointLifeTime, tikvCfg.Concurrency); err != nil {
			zap.L().Error("[TiKVGC] do TiKVGC failed", zap.Error(err))
			continue
		}
	}
}

func runTiKVGC(ctx context.Context, db *DB, uuid []byte, lifeTime time.Duration, concurrency int) error {
	newPoint, err := getNewSafePoint(db, lifeTime)
	if err != nil {
		return err
	}

	lastPoint, err := getLastSafePoint(db)
	if err != nil {
		return err
	}

	if lastPoint != nil && newPoint.Before(*lastPoint) {
		zap.L().Info("[TiKVGC] last safe point is later than current on,no need to gc.",
			zap.Time("last", *lastPoint), zap.Time("current", *newPoint))
		return nil
	}

	if lastPoint == nil {
		zap.L().Info("[TiKVGC] current safe point ", zap.Time("current", *newPoint))
	} else {
		zap.L().Info("[TiKVGC] current safe point ", zap.Time("current", *newPoint), zap.Time("last", *lastPoint))
	}

	if err := saveLastSafePoint(ctx, db, newPoint); err != nil {
		zap.L().Error("[TiKVGC] save last safe point err ", zap.Time("current", *newPoint))
		return err
	}
	safePoint := oracle.ComposeTS(oracle.GetPhysical(*newPoint), 0)
	if err := gcworker.RunGCJob(ctx, db.kv.Storage.(tikv.Storage), safePoint, UUIDString(uuid), concurrency); err != nil {
		return err
	}
	return nil

}

func saveLastSafePoint(ctx context.Context, db *DB, safePoint *time.Time) error {
	txn, err := db.Begin()
	if err != nil {
		return err
	}
	if err := txn.t.Set(sysTiKVGCLastSafePoint, []byte(safePoint.Format(tikvGcTimeFormat))); err != nil {
		return err
	}
	if err := txn.t.Commit(ctx); err != nil {
		if err := txn.Rollback(); err != nil {
			zap.L().Error("rollback failed", zap.Error(err))
		}
		return err
	}
	return nil
}

func getNewSafePoint(db *DB, lifeTime time.Duration) (*time.Time, error) {
	currentVer, err := db.kv.CurrentVersion()
	if err != nil {
		return nil, err
	}
	physical := oracle.ExtractPhysical(currentVer.Ver)
	sec, nsec := physical/1e3, (physical%1e3)*1e6
	now := time.Unix(sec, nsec)
	safePoint := now.Add(-lifeTime)
	return &safePoint, nil
}

func getLastSafePoint(db *DB) (*time.Time, error) {
	txn, err := db.Begin()
	if err != nil {
		return nil, err
	}
	val, err := txn.t.Get(sysTiKVGCLastSafePoint)
	if err != nil {
		if IsErrNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	str := string(val)
	if str == "" {
		return nil, nil
	}
	t, err := time.Parse(tikvGcTimeFormat, str)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
