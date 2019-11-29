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
	LIMITDATA_DBID             = 0
	ALL_NAMESPACE              = "*"
	NAMESPACE_COMMAND_TOKEN    = "@"
	QPS_PREFIX                 = "qps:"
	RATE_PREFIX                = "rate:"
	LIMIT_VALUE_TOKEN          = " "
	LIMITER_STATUS_PREFIX      = "limiter_status:"
	LIMITER_STATUS_VALUE_TOKEN = ","
	TIME_FORMAT                = "2006-01-02 15:04:05"
	MAXIMUM_WEIGHT             = 1
	MINIMUM_WEIGHT             = 0.1
)

type CommandLimiter struct {
	localIp     string
	limiterName string

	qpsl         *rate.Limiter
	ratel        *rate.Limiter
	localPercent float64
	weight       float64
	skipBalance  bool

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
	limitDatadb *DB
	conf        *conf.RateLimit
	localIp     string

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

func getLimiterKey(limiterName string) []byte {
	var key []byte
	key = append(key, []byte(LIMITER_STATUS_PREFIX)...)
	key = append(key, []byte(limiterName)...)
	key = append(key, ':')
	return key
}

func getNamespaceAndCmd(limiterName string) []string {
	strs := strings.Split(limiterName, NAMESPACE_COMMAND_TOKEN)
	if len(strs) < 2 {
		return nil
	}
	return strs

}

func NewLimitersMgr(store *RedisStore, rateLimit *conf.RateLimit) (*LimitersMgr, error) {
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

	if rateLimit.LimiterNamespace == "" {
		return nil, errors.New("limiter-namespace is configured with empty")
	}
	if rateLimit.WeightChangeFactor <= 1 {
		return nil, errors.New("weight-change-factor should > 1")
	}
	if !(rateLimit.UsageToDivide > 0 && rateLimit.UsageToDivide < rateLimit.UsageToMultiply && rateLimit.UsageToMultiply < 1) {
		return nil, errors.New("should config 0 < usage-to-divide < usage-to-multiply < 1")
	}
	if rateLimit.InitialPercent > 1 || rateLimit.InitialPercent <= 0 {
		return nil, errors.New("initial-percent should in (0, 1]")
	}

	l := &LimitersMgr{
		limitDatadb: store.DB(rateLimit.LimiterNamespace, LIMITDATA_DBID),
		conf:        rateLimit,
		localIp:     localIp,
	}

	go l.startSyncNewLimit()
	go l.startReportAndBalance()
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
		newCl := NewCommandLimiter(l.localIp, limiterName, qpsLimit, qpsBurst, rateLimit, rateBurst, l.conf.InitialPercent)
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

	limitStrs := strings.Split(string(val), LIMIT_VALUE_TOKEN)
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

func (l *LimitersMgr) startReportAndBalance() {
	ticker := time.NewTicker(l.conf.GlobalBalancePeriod)
	defer ticker.Stop()
	for range ticker.C {
		l.runReportAndBalance()
	}
}

func (l *LimitersMgr) runReportAndBalance() {
	l.limiters.Range(func(k, v interface{}) bool {
		limiterName := k.(string)
		commandLimiter := v.(*CommandLimiter)
		if commandLimiter != nil {
			averageQps := commandLimiter.reportLocalStat(l.conf.GlobalBalancePeriod)
			commandLimiter.balanceLimit(averageQps, l.limitDatadb, l.conf.TitanStatusLifetime, l.conf.UsageToDivide, l.conf.UsageToMultiply, l.conf.WeightChangeFactor)

		} else {
			namespaceAndCmd := getNamespaceAndCmd(limiterName)
			metrics.GetMetrics().LimiterQpsVec.WithLabelValues(namespaceAndCmd[0], namespaceAndCmd[1], l.localIp).Set(0)
			metrics.GetMetrics().LimiterRateVec.WithLabelValues(namespaceAndCmd[0], namespaceAndCmd[1], l.localIp).Set(0)
		}
		return true
	})
}

func (l *LimitersMgr) startSyncNewLimit() {
	ticker := time.NewTicker(l.conf.SyncSetPeriod)
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
				newCl := NewCommandLimiter(l.localIp, limiterName, qpsLimit, qpsBurst, rateLimit, rateBurst, l.conf.InitialPercent)
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

func NewCommandLimiter(localIp string, limiterName string, qpsLimit float64, qpsBurst int, rateLimit float64, rateBurst int, initialPercent float64) *CommandLimiter {
	if !(qpsLimit > 0 && qpsBurst > 0) &&
		!(rateLimit > 0 && rateBurst > 0) {
		return nil
	}
	if initialPercent <= 0 {
		return nil
	}
	var qpsl, ratel *rate.Limiter
	if qpsLimit > 0 && qpsBurst > 0 {
		qpsl = rate.NewLimiter(rate.Limit(qpsLimit*initialPercent), qpsBurst)
	}
	if rateLimit > 0 && rateBurst > 0 {
		ratel = rate.NewLimiter(rate.Limit(rateLimit*initialPercent), rateBurst)
	}
	cl := &CommandLimiter{
		limiterName:  limiterName,
		localIp:      localIp,
		qpsl:         qpsl,
		ratel:        ratel,
		localPercent: initialPercent,
		weight:       MAXIMUM_WEIGHT,
		skipBalance:  true,
		lastTime:     time.Now(),
	}
	return cl
}

func (cl *CommandLimiter) updateLimit(qpsLimit float64, qpsBurst int, rateLimit float64, rateBurst int) {
	cl.lock.Lock()
	defer cl.lock.Unlock()

	////when limit is changed, the qps can't be used to balanceLimit
	var orgQpsLimit, orgRateLimit float64
	var orgQpsBurst, orgRateBurst int
	if cl.qpsl != nil {
		orgQpsLimit = float64(cl.qpsl.Limit()) / cl.localPercent
		orgQpsBurst = cl.qpsl.Burst()
	}
	if cl.ratel != nil {
		orgRateLimit = float64(cl.ratel.Limit()) / cl.localPercent
		orgRateBurst = cl.ratel.Burst()
	}
	if orgQpsLimit != qpsLimit || orgQpsBurst != qpsBurst || orgRateLimit != rateLimit || orgRateBurst != rateBurst {
		cl.skipBalance = true
	}
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
}

func (cl *CommandLimiter) reportLocalStat(globalBalancePeriod time.Duration) float64 {
	var qpsLocal, rateLocal float64
	cl.lock.Lock()
	defer cl.lock.Unlock()
	seconds := time.Since(cl.lastTime).Seconds()
	if seconds >= 0 {
		qpsLocal = float64(cl.totalCommandsCount) / seconds
		rateLocal = float64(cl.totalCommandsSize) / 1024 / seconds
	} else {
		qpsLocal = 0
		rateLocal = 0
	}
	cl.totalCommandsCount = 0
	cl.totalCommandsSize = 0
	cl.lastTime = time.Now()

	namespaceCmd := getNamespaceAndCmd(cl.limiterName)
	metrics.GetMetrics().LimiterQpsVec.WithLabelValues(namespaceCmd[0], namespaceCmd[1], cl.localIp).Set(qpsLocal)
	metrics.GetMetrics().LimiterRateVec.WithLabelValues(namespaceCmd[0], namespaceCmd[1], cl.localIp).Set(rateLocal)

	return qpsLocal
}

func (cl *CommandLimiter) balanceLimit(averageQps float64, limitDatadb *DB, titanStatusLifetime time.Duration,
	devideUsage float64, multiplyUsage float64, weightChangeFactor float64) {
	cl.lock.Lock()
	defer cl.lock.Unlock()

	if cl.qpsl == nil {
		return
	}
	if cl.skipBalance {
		cl.skipBalance = false
		return
	}

	txn, err := limitDatadb.Begin()
	if err != nil {
		zap.L().Error("[Limit] transection begin failed", zap.String("titan", cl.localIp), zap.Error(err))
		return
	}

	weights, qpss, err := cl.scanStatusInOtherTitan(limitDatadb, txn, titanStatusLifetime)
	if err != nil {
		txn.Rollback()
		return
	}

	totalWeight := cl.weight
	for i := range weights {
		totalWeight += weights[i]
	}

	originalLimit := float64(cl.qpsl.Limit()) / cl.localPercent
	selfLimitInTarget := originalLimit * (cl.weight / totalWeight)
	if averageQps < selfLimitInTarget*devideUsage {
		otherFull := false
		for i := range qpss {
			otherLimitInTarget := originalLimit * (weights[i] / totalWeight)
			if qpss[i] >= otherLimitInTarget*multiplyUsage {
				otherFull = true
				break
			}
		}
		if otherFull {
			cl.weight /= weightChangeFactor
			if cl.weight < MINIMUM_WEIGHT {
				cl.weight = MINIMUM_WEIGHT
			}
		}
	} else if averageQps >= selfLimitInTarget*multiplyUsage {
		cl.weight *= weightChangeFactor
		if cl.weight > MAXIMUM_WEIGHT {
			cl.weight = MAXIMUM_WEIGHT
		}
	}

	totalWeight = cl.weight
	for i := range weights {
		totalWeight += weights[i]
	}
	newPercent := cl.weight / totalWeight

	key := getLimiterKey(cl.limiterName)
	key = append(key, []byte(cl.localIp)...)
	s := NewString(txn, key)
	now := time.Now()
	strTime := now.Format(TIME_FORMAT)
	value := fmt.Sprintf("%f%s%f%s%s", cl.weight, LIMITER_STATUS_VALUE_TOKEN, averageQps, LIMITER_STATUS_VALUE_TOKEN, strTime)
	if err := s.Set([]byte(value), 0); err != nil {
		txn.Rollback()
		return
	}
	if err := txn.t.Commit(context.Background()); err != nil {
		zap.L().Error("[Limit] commit after balance limit failed", zap.String("titan", cl.localIp))
		txn.Rollback()
		return
	}
	zap.L().Info("[Limit] balance limit", zap.String("limiterName", cl.limiterName),
		zap.Float64("qps", averageQps), zap.Float64("newWeight", cl.weight), zap.Float64("newPercent", newPercent))
	cl.updateLimitPercent(newPercent)
}

func (cl *CommandLimiter) scanStatusInOtherTitan(limitDatadb *DB, txn *Transaction, titanStatusLifetime time.Duration) ([]float64, []float64, error) {
	key := getLimiterKey(cl.limiterName)
	prefix := MetaKey(limitDatadb, key)
	endPrefix := sdk_kv.Key(prefix).PrefixNext()
	iter, err := txn.t.Iter(prefix, endPrefix)
	if err != nil {
		zap.L().Error("[Limit] seek failed", zap.ByteString("prefix", prefix), zap.Error(err))
		return nil, nil, err
	}
	defer iter.Close()

	prefixLen := len(prefix)
	var weights, qpss []float64
	var weight, qps float64
	for ; iter.Valid() && iter.Key().HasPrefix(prefix); err = iter.Next() {
		if err != nil {
			zap.L().Error("[Limit] next failed", zap.ByteString("prefix", prefix), zap.Error(err))
			return nil, nil, err
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

		val := string(obj.Meta.Value)
		vals := strings.Split(val, LIMITER_STATUS_VALUE_TOKEN)
		if len(vals) < 3 {
			zap.L().Error("[Limit] short of values(should 3 values)", zap.ByteString("key", key), zap.String("value", val))
			continue
		}
		sWeight := vals[0]
		sQps := vals[1]
		lastActive := vals[2]

		if weight, err = strconv.ParseFloat(sWeight, 64); err != nil {
			zap.L().Error("[Limit] weight can't be decoded to float", zap.ByteString("key", key), zap.String("weight", sWeight), zap.Error(err))
			continue
		}
		if qps, err = strconv.ParseFloat(sQps, 64); err != nil {
			zap.L().Error("[Limit] qps can't be decoded to float", zap.ByteString("key", key), zap.String("qps", sQps), zap.Error(err))
			continue
		}

		lastActiveT, err := time.ParseInLocation(TIME_FORMAT, lastActive, time.Local)
		if err != nil {
			zap.L().Error("[Limit] value can't decoded into a time", zap.ByteString("key", key), zap.String("lastActive", lastActive), zap.Error(err))
			continue
		}

		zap.L().Info("[Limit] other titan status", zap.ByteString("key", key), zap.Float64("weight", weight), zap.Float64("qps", qps), zap.String("lastActive", lastActive))
		if string(ip) != cl.localIp && time.Since(lastActiveT) <= titanStatusLifetime {
			weights = append(weights, weight)
			qpss = append(qpss, qps)
		}
	}
	return weights, qpss, nil
}

func (cl *CommandLimiter) updateLimitPercent(newPercent float64) {
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
