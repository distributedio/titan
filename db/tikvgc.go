package db

import (
	"context"
	"time"

	"github.com/meitu/titan/conf"
	"github.com/pingcap/tidb/store/tikv"
	"github.com/pingcap/tidb/store/tikv/gcworker"
	"github.com/pingcap/tidb/store/tikv/oracle"
	"go.uber.org/zap"
)

var (
	sysTikvGCLeader        = []byte("$sys:0:TGC:GCLeader")
	sysTikvGCLastSafePoint = []byte("$sys:0:TGC:LastSafePoint")
)

const (
	tikvGcTimeFormat = "20060102-15:04:05 -0700 MST"
)

// StartTikvGC start tikv gcwork
func StartTikvGC(db *DB, tikvCfg *conf.TikvGC) {
	ticker := time.NewTicker(tikvCfg.Interval)
	defer ticker.Stop()
	uuid := UUID()
	ctx := context.Background()
	for range ticker.C {
		isLeader, err := isLeader(db, sysTikvGCLeader, uuid, tikvCfg.LeaderLifeTime)
		if err != nil {
			zap.L().Error("[TikvGC] check TikvGC leader failed", zap.Error(err))
			continue
		}
		if !isLeader {
			zap.L().Debug("[TikvGC] not TikvGC leader")
			continue
		}
		if err := runTikvGC(ctx, db, uuid, tikvCfg.SafePointLifeTime, tikvCfg.Concurrency); err != nil {
			zap.L().Error("[TikvGC] do TikvGC failed", zap.Error(err))
			continue
		}
	}
}

func runTikvGC(ctx context.Context, db *DB, uuid []byte, lifeTime time.Duration, concurrency int) error {
	newPoint, err := getNewSafePoint(db, lifeTime)
	if err != nil {
		return err
	}

	lastPoint, err := getLastSafePoint(db)
	if err != nil {
		return err
	}

	if lastPoint != nil && newPoint.Before(*lastPoint) {
		zap.L().Info("[TikvGC] last safe point is later than current on,no need to gc.",
			zap.Time("last", *lastPoint), zap.Time("current", *newPoint))
		return nil
	}

	if lastPoint == nil {
		zap.L().Info("[TikvGC] current safe point ", zap.Time("current", *newPoint))
	} else {
		zap.L().Info("[TikvGC] current safe point ", zap.Time("current", *newPoint), zap.Time("last", *lastPoint))
	}

	if err := saveLastSafePoint(ctx, db, newPoint); err != nil {
		zap.L().Error("[TikvGC] save last safe point err ", zap.Time("current", *newPoint))
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
	if err := txn.t.Set(sysTikvGCLastSafePoint, []byte(safePoint.Format(tikvGcTimeFormat))); err != nil {
		return err
	}
	if err := txn.t.Commit(ctx); err != nil {
		txn.Rollback()
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
	val, err := txn.t.Get(sysTikvGCLastSafePoint)
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
