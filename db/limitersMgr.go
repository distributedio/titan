package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/distributedio/titan/conf"
	"github.com/distributedio/titan/metrics"
	sdk_kv "github.com/pingcap/tidb/kv"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	LIMITDATA_NAMESPACE     = "sys_ratelimit"
	LIMITDATA_DBID          = 0
	ALL_NAMESPACE           = "*"
	NAMESPACE_COMMAND_TOKEN = "@"
	QPS_PREFIX              = "qps:"
	RATE_PREFIX             = "rate:"
	LIMIT_BURST_TOKEN       = " "
	TITAN_STATUS_TOKEN      = "titan_status:"
	TIME_FORMAT             = "2006-01-02 15:04:05"
)

type CommandLimiter struct {
	localIp     string
	limiterName string

	qpsl         *rate.Limiter
	ratel        *rate.Limiter
	localPercent float64

	lock               sync.Mutex
	lastTime           time.Time
	totalCommandsCount int64
	totalCommandsSize  int64
}

type LimitData struct {
	limit float64
	burst int
}

type LimitersMgr struct {
	limitDatadb         *DB
	globalBalancePeriod time.Duration
	titanStatusLifeTime time.Duration
	syncSetPeriod       time.Duration
	localIp             string
	localPercent        float64

	limiters          sync.Map
	qpsAllmatchLimit  sync.Map
	rateAllmatchLimit sync.Map
	lock              sync.Mutex
}

func getAllmatchLimiterName(limiterName string) string {
	strs := strings.Split(limiterName, NAMESPACE_COMMAND_TOKEN)
	if len(strs) < 2 {
		return ""
	}
	return fmt.Sprintf("%s%s%s", ALL_NAMESPACE, NAMESPACE_COMMAND_TOKEN, strs[1])
}

func NewLimitersMgr(store *RedisStore, rateLimit conf.RateLimit) (*LimitersMgr, error) {
	var addrs []net.Addr
	var err error
	if rateLimit.InterfaceName != "" {
		iface, err := net.InterfaceByName(rateLimit.InterfaceName)
		if err != nil {
			return nil, err
		}

		addrs, err = iface.Addrs()
		if err != nil {
			return nil, err
		}
	} else {
		addrs, err = net.InterfaceAddrs()
		if err != nil {
			return nil, err
		}
	}

	localIp := ""
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			localIp = ipnet.IP.String()
			break
		}
	}
	if localIp == "" {
		return nil, errors.New(rateLimit.InterfaceName + " adds is empty")
	}

	l := &LimitersMgr{
		limitDatadb:         store.DB(LIMITDATA_NAMESPACE, LIMITDATA_DBID),
		globalBalancePeriod: rateLimit.GlobalBalancePeriod,
		titanStatusLifeTime: rateLimit.TitanStatusLifetime,
		syncSetPeriod:       rateLimit.SyncSetPeriod,
		localIp:             localIp,
		localPercent:        1,
	}

	go l.startBalanceLimit()
	go l.startSyncNewLimit()
	return l, nil
}

func (l *LimitersMgr) init(limiterName string) *CommandLimiter {
	//lock is just prevent many new connection of same namespace to getlimit from tikv in same time
	l.lock.Lock()
	defer l.lock.Unlock()

	v, ok := l.limiters.Load(limiterName)
	if ok {
		return v.(*CommandLimiter)
	}

	allmatchLimiterName := getAllmatchLimiterName(limiterName)
	l.qpsAllmatchLimit.LoadOrStore(allmatchLimiterName, (*LimitData)(nil))
	l.rateAllmatchLimit.LoadOrStore(allmatchLimiterName, (*LimitData)(nil))

	qpsLimit, qpsBurst := l.getLimit(limiterName, true)
	rateLimit, rateBurst := l.getLimit(limiterName, false)
	if (qpsLimit > 0 && qpsBurst > 0) ||
		(rateLimit > 0 && rateBurst > 0) {
		newCl := NewCommandLimiter(l.localIp, limiterName, qpsLimit, qpsBurst, rateLimit, rateBurst, l.localPercent)
		v, _ := l.limiters.LoadOrStore(limiterName, newCl)
		return v.(*CommandLimiter)
	} else {
		l.limiters.LoadOrStore(limiterName, (*CommandLimiter)(nil))
		return nil
	}
}

func (l *LimitersMgr) getLimit(limiterName string, isQps bool) (float64, int) {
	limit := float64(0)
	burst := int64(0)

	txn, err := l.limitDatadb.Begin()
	if err != nil {
		zap.L().Error("[Limit] transection begin failed", zap.String("limiterName", limiterName), zap.Bool("isQps", isQps), zap.Error(err))
		return 0, 0
	}
	defer func() {
		if err := txn.t.Commit(context.Background()); err != nil {
			zap.L().Error("[Limit] commit after get limit failed", zap.String("limiterName", limiterName), zap.Error(err))
			txn.t.Rollback()
		}
	}()

	var limiterKey string
	if isQps {
		limiterKey = QPS_PREFIX + limiterName
	} else {
		limiterKey = RATE_PREFIX + limiterName
	}

	str, err := txn.String([]byte(limiterKey))
	if err != nil {
		zap.L().Error("[Limit] get limit's value failed", zap.String("key", limiterKey), zap.Error(err))
		return 0, 0
	}
	val, err := str.Get()
	if err != nil {
		return 0, 0
	}

	limitStrs := strings.Split(string(val), LIMIT_BURST_TOKEN)
	if len(limitStrs) < 2 {
		zap.L().Error("[Limit] limit hasn't enough parameters, should be: <limit>[K|k|M|m] <burst>", zap.String("key", limiterKey), zap.ByteString("val", val))
		return 0, 0
	}
	limitStr := limitStrs[0]
	burstStr := limitStrs[1]
	if len(limitStr) < 1 {
		zap.L().Error("[Limit] limit part's length isn't enough, should be: <limit>[K|k|M|m] <burst>", zap.String("key", limiterKey), zap.ByteString("val", val))
		return 0, 0
	}
	var strUnit uint8
	var unit float64
	strUnit = limitStr[len(limitStr)-1]
	if strUnit == 'k' || strUnit == 'K' {
		unit = 1024
		limitStr = limitStr[:len(limitStr)-1]
	} else if strUnit == 'm' || strUnit == 'M' {
		unit = 1024 * 1024
		limitStr = limitStr[:len(limitStr)-1]
	} else {
		unit = 1
	}
	if limit, err = strconv.ParseFloat(limitStr, 64); err != nil {
		zap.L().Error("[Limit] limit can't be decoded to integer", zap.String("key", limiterKey), zap.ByteString("val", val), zap.Error(err))
		return 0, 0
	}
	limit *= unit
	if burst, err = strconv.ParseInt(burstStr, 10, 64); err != nil {
		zap.L().Error("[Limit] burst can't be decoded to integer", zap.String("key", limiterKey), zap.ByteString("val", val), zap.Error(err))
		return 0, 0
	}

	if logEnv := zap.L().Check(zap.DebugLevel, "[Limit] got limit"); logEnv != nil {
		logEnv.Write(zap.String("key", limiterKey), zap.Float64("limit", limit), zap.Int64("burst", burst))
	}

	return limit, int(burst)
}

func (l *LimitersMgr) CheckLimit(namespace string, cmdName string, cmdArgs []string) {
	if namespace == LIMITDATA_NAMESPACE {
		return
	}
	limiterName := fmt.Sprintf("%s%s%s", namespace, NAMESPACE_COMMAND_TOKEN, cmdName)
	v, ok := l.limiters.Load(limiterName)
	var commandLimiter *CommandLimiter
	if !ok {
		commandLimiter = l.init(limiterName)
	} else {
		commandLimiter = v.(*CommandLimiter)
	}

	if commandLimiter != nil {
		now := time.Now()
		commandLimiter.checkLimit(cmdName, cmdArgs)
		cost := time.Since(now).Seconds()
		metrics.GetMetrics().LimitCostHistogramVec.WithLabelValues(namespace, cmdName).Observe(cost)
	}
}

func (l *LimitersMgr) startBalanceLimit() {
	ticker := time.NewTicker(l.globalBalancePeriod)
	defer ticker.Stop()
	for range ticker.C {
		l.runBalanceLimit()
	}
}

func (l *LimitersMgr) runBalanceLimit() {
	txn, err := l.limitDatadb.Begin()
	if err != nil {
		zap.L().Error("[Limit] transection begin failed", zap.String("titan", l.localIp), zap.Error(err))
		return
	}

	prefix := MetaKey(l.limitDatadb, []byte(TITAN_STATUS_TOKEN))
	endPrefix := sdk_kv.Key(prefix).PrefixNext()
	iter, err := txn.t.Iter(prefix, endPrefix)
	if err != nil {
		zap.L().Error("[Limit] seek failed", zap.ByteString("prefix", prefix), zap.Error(err))
		txn.Rollback()
		return
	}
	defer iter.Close()

	activeNum := float64(1)
	prefixLen := len(prefix)
	for ; iter.Valid() && iter.Key().HasPrefix(prefix); err = iter.Next() {
		if err != nil {
			zap.L().Error("[Limit] next failed", zap.ByteString("prefix", prefix), zap.Error(err))
			txn.Rollback()
			return
		}

		key := iter.Key()
		if len(key) <= prefixLen {
			zap.L().Error("ip is null", zap.ByteString("key", key))
			continue
		}
		ip := key[prefixLen:]
		obj := NewString(txn, key)
		if err = obj.decode(iter.Value()); err != nil {
			zap.L().Error("[Limit] Strings decoded value error", zap.ByteString("key", key), zap.Error(err))
			continue
		}

		lastActive := obj.Meta.Value
		lastActiveT, err := time.ParseInLocation(TIME_FORMAT, string(lastActive), time.Local)
		if err != nil {
			zap.L().Error("[Limit] value can't decoded into a time", zap.ByteString("key", key), zap.ByteString("lastActive", lastActive), zap.Error(err))
			continue
		}

		diff := time.Since(lastActiveT).Seconds()
		zap.L().Info("[Limit] last active time", zap.ByteString("ip", ip), zap.ByteString("lastActive", lastActive), zap.Float64("activePast", diff))
		if string(ip) != l.localIp && time.Since(lastActiveT) <= l.titanStatusLifeTime {
			activeNum++
		}
	}
	newPercent := 1 / activeNum
	var oldPercent float64
	l.lock.Lock()
	oldPercent = l.localPercent
	l.lock.Unlock()
	if oldPercent != newPercent {
		zap.L().Info("[Limit] balance limit in all titan server", zap.Float64("active server num", activeNum),
			zap.Float64("oldPercent", oldPercent), zap.Float64("newPercent", newPercent))
		l.limiters.Range(func(k, v interface{}) bool {
			commandLimiter := v.(*CommandLimiter)
			if commandLimiter != nil {
				commandLimiter.updateLimitPercent(newPercent)
			}
			return true
		})

		l.lock.Lock()
		l.localPercent = newPercent
		l.lock.Unlock()
	}

	key := []byte(TITAN_STATUS_TOKEN + l.localIp)
	s := NewString(txn, key)
	now := time.Now()
	value := now.Format(TIME_FORMAT)
	if err := s.Set([]byte(value), 0); err != nil {
		txn.Rollback()
		return
	}
	if err := txn.t.Commit(context.Background()); err != nil {
		zap.L().Error("[Limit] commit after balance limit failed", zap.String("titan", l.localIp))
		txn.Rollback()
		return
	}
}

func (l *LimitersMgr) startSyncNewLimit() {
	ticker := time.NewTicker(l.syncSetPeriod)
	defer ticker.Stop()
	for range ticker.C {
		l.runSyncNewLimit()
	}
}

func (l *LimitersMgr) runSyncNewLimit() {
	allmatchLimits := []*sync.Map{&l.qpsAllmatchLimit, &l.rateAllmatchLimit}
	for i, allmatchLimit := range allmatchLimits {
		allmatchLimit.Range(func(k, v interface{}) bool {
			limiterName := k.(string)
			limitData := v.(*LimitData)
			isQps := false
			if i == 0 {
				isQps = true
			}
			limit, burst := l.getLimit(limiterName, isQps)
			if limit > 0 && burst > 0 {
				if limitData == nil {
					limitData = &LimitData{limit, burst}
					allmatchLimit.Store(limiterName, limitData)
				} else {
					limitData.limit = limit
					limitData.burst = burst
				}
			} else {
				allmatchLimit.Store(limiterName, (*LimitData)(nil))
			}
			return true
		})
	}

	var localPercent float64
	l.lock.Lock()
	localPercent = l.localPercent
	l.lock.Unlock()
	l.limiters.Range(func(k, v interface{}) bool {
		limiterName := k.(string)
		commandLimiter := v.(*CommandLimiter)
		allmatchLimiterName := getAllmatchLimiterName(limiterName)
		qpsLimit, qpsBurst := l.getLimit(limiterName, true)
		if !(qpsLimit > 0 && qpsBurst > 0) {
			v, ok := l.qpsAllmatchLimit.Load(allmatchLimiterName)
			if ok {
				limitData := v.(*LimitData)
				if limitData != nil {
					qpsLimit = limitData.limit
					qpsBurst = limitData.burst
				}
			}
		}
		rateLimit, rateBurst := l.getLimit(limiterName, false)
		if !(rateLimit > 0 && rateBurst > 0) {
			v, ok := l.rateAllmatchLimit.Load(allmatchLimiterName)
			if ok {
				limitData := v.(*LimitData)
				if limitData != nil {
					rateLimit = limitData.limit
					rateBurst = limitData.burst
				}
			}
		}

		if (qpsLimit > 0 && qpsBurst > 0) ||
			(rateLimit > 0 && rateBurst > 0) {
			if logEnv := zap.L().Check(zap.DebugLevel, "[Limit] limit is set"); logEnv != nil {
				logEnv.Write(zap.String("limiter name", limiterName), zap.Float64("qps limit", qpsLimit), zap.Int("qps burst", qpsBurst),
					zap.Float64("rate limit", rateLimit), zap.Int("rate burst", rateBurst))
			}
			if commandLimiter == nil {
				newCl := NewCommandLimiter(l.localIp, limiterName, qpsLimit, qpsBurst, rateLimit, rateBurst, localPercent)
				l.limiters.Store(limiterName, newCl)
			} else {
				commandLimiter.updateLimit(qpsLimit, qpsBurst, rateLimit, rateBurst)
			}
		} else {
			if commandLimiter != nil {
				if logEnv := zap.L().Check(zap.DebugLevel, "[Limit] limit is cleared"); logEnv != nil {
					logEnv.Write(zap.String("limiter name", limiterName), zap.Float64("qps limit", qpsLimit), zap.Int("qps burst", qpsBurst),
						zap.Float64("rate limit", rateLimit), zap.Int("rate burst", rateBurst))
				}
				l.limiters.Store(limiterName, (*CommandLimiter)(nil))
			}
		}
		return true
	})
}

func NewCommandLimiter(localIp string, limiterName string, qpsLimit float64, qpsBurst int, rateLimit float64, rateBurst int, localPercent float64) *CommandLimiter {
	var qpsl, ratel *rate.Limiter
	if qpsLimit > 0 && qpsBurst > 0 {
		qpsl = rate.NewLimiter(rate.Limit(qpsLimit*localPercent), qpsBurst)
	}
	if rateLimit > 0 && rateBurst > 0 {
		ratel = rate.NewLimiter(rate.Limit(rateLimit*localPercent), rateBurst)
	}
	cl := &CommandLimiter{
		limiterName:  limiterName,
		localIp:      localIp,
		qpsl:         qpsl,
		ratel:        ratel,
		localPercent: localPercent,
		lastTime:     time.Now(),
	}
	return cl
}

func (cl *CommandLimiter) updateLimit(qpsLimit float64, qpsBurst int, rateLimit float64, rateBurst int) {
	cl.lock.Lock()
	defer cl.lock.Unlock()

	if qpsLimit > 0 && qpsBurst > 0 {
		if cl.qpsl != nil {
			if cl.qpsl.Burst() != qpsBurst {
				cl.qpsl = rate.NewLimiter(rate.Limit(qpsLimit*cl.localPercent), qpsBurst)
			} else if cl.qpsl.Limit() != rate.Limit(qpsLimit*cl.localPercent) {
				cl.qpsl.SetLimit(rate.Limit(qpsLimit * cl.localPercent))
			}
		} else {
			cl.qpsl = rate.NewLimiter(rate.Limit(qpsLimit*cl.localPercent), qpsBurst)
		}
	} else {
		cl.qpsl = nil
	}

	if rateLimit > 0 && rateBurst > 0 {
		if cl.ratel != nil {
			if cl.ratel.Burst() != rateBurst {
				cl.ratel = rate.NewLimiter(rate.Limit(rateLimit*cl.localPercent), rateBurst)
			} else if cl.ratel.Limit() != rate.Limit(rateLimit*cl.localPercent) {
				cl.ratel.SetLimit(rate.Limit(rateLimit * cl.localPercent))
			}
		} else {
			cl.ratel = rate.NewLimiter(rate.Limit(rateLimit*cl.localPercent), rateBurst)
		}
	} else {
		cl.ratel = nil
	}

	var qpsLocal, rateLocal float64
	seconds := time.Since(cl.lastTime).Seconds()
	if seconds >= 0 {
		qpsLocal = float64(cl.totalCommandsCount) / seconds
		rateLocal = float64(cl.totalCommandsSize) / 1024 / seconds
	} else {
		qpsLocal = 0
		rateLocal = 0
	}
	metrics.GetMetrics().LimiterQpsVec.WithLabelValues(cl.localIp, cl.limiterName).Set(qpsLocal)
	metrics.GetMetrics().LimiterRateVec.WithLabelValues(cl.localIp, cl.limiterName).Set(rateLocal)
	cl.totalCommandsCount = 0
	cl.totalCommandsSize = 0
	cl.lastTime = time.Now()
}

func (cl *CommandLimiter) updateLimitPercent(newPercent float64) {
	cl.lock.Lock()
	defer cl.lock.Unlock()

	if cl.localPercent != newPercent && cl.localPercent > 0 && newPercent > 0 {
		if cl.qpsl != nil {
			qpsLimit := (float64(cl.qpsl.Limit()) / cl.localPercent) * newPercent
			zap.L().Info("percent changed", zap.String("limiterName", cl.limiterName), zap.Float64("qps limit", qpsLimit), zap.Int("burst", cl.qpsl.Burst()))
			cl.qpsl.SetLimit(rate.Limit(qpsLimit))
		}
		if cl.ratel != nil {
			rateLimit := float64(cl.ratel.Limit()) / cl.localPercent * newPercent
			zap.L().Info("percent changed", zap.String("limiterName", cl.limiterName), zap.Float64("rate limit", rateLimit), zap.Int("burst", cl.ratel.Burst()))
			cl.ratel.SetLimit(rate.Limit(rateLimit))
		}
		cl.localPercent = newPercent
	}
}

func (cl *CommandLimiter) checkLimit(cmdName string, cmdArgs []string) {
	cl.lock.Lock()
	defer cl.lock.Unlock()

	if cl.qpsl != nil {
		r := cl.qpsl.Reserve()
		if !r.OK() {
			zap.L().Error("[Limit] request events num exceed limiter burst", zap.Int("qps limiter burst", cl.qpsl.Burst()))
		} else {
			d := r.Delay()
			if d > 0 {
				if logEnv := zap.L().Check(zap.DebugLevel, "[Limit] trigger qps limit"); logEnv != nil {
					logEnv.Write(zap.String("limiter name", cl.limiterName), zap.Int64("sleep us", int64(d/time.Microsecond)))
				}
				time.Sleep(d)
			}
		}
	}

	cmdSize := len(cmdName)
	for i := range cmdArgs {
		cmdSize += len(cmdArgs[i]) + 1
	}
	if cl.ratel != nil {
		r := cl.ratel.ReserveN(time.Now(), cmdSize)
		if !r.OK() {
			zap.L().Error("[Limit] request events num exceed limiter burst", zap.Int("rate limiter burst", cl.ratel.Burst()), zap.Int("command size", cmdSize))
		} else {
			d := r.Delay()
			if d > 0 {
				if logEnv := zap.L().Check(zap.DebugLevel, "[Limit] trigger rate limit"); logEnv != nil {
					logEnv.Write(zap.String("limiter name", cl.limiterName), zap.Strings("args", cmdArgs), zap.Int64("sleep us", int64(d/time.Microsecond)))
				}
				time.Sleep(d)
			}
		}
	}

	cl.totalCommandsCount++
	cl.totalCommandsSize += int64(cmdSize)
	if logEnv := zap.L().Check(zap.DebugLevel, "[Limit] limiter state"); logEnv != nil {
		logEnv.Write(zap.String("limiter name", cl.limiterName), zap.Time("last time", cl.lastTime), zap.Int64("command count", cl.totalCommandsCount), zap.Int64("command size", cl.totalCommandsSize))
	}
}
