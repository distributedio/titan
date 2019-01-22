package metrics

import (
	"net/http"

	"go.uber.org/zap/zapcore"

	sdk_metrics "github.com/pingcap/tidb/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	//promethus default namespace
	namespace = "titan"

	//promethues default label key
	command   = "command"
	biz       = "biz"
	leader    = "leader"
	ztinfo    = "ztinfo"
	labelName = "level"
	gckeys    = "gckeys"
	expire    = "expire"
	tikvGC    = "tikvgc"
)

var (
	//Label value slice when creating prometheus object
	commandLabel = []string{command}
	bizLabel     = []string{biz}
	leaderLabel  = []string{leader}
	multiLabel   = []string{biz, command}
	ztInfoLabel  = []string{ztinfo}
	gcKeysLabel  = []string{gckeys}
	expireLabel  = []string{expire}
	tikvGCLabel  = []string{tikvGC}

	// global prometheus object
	gm *Metrics
)

// Metrics prometheus statistics
type Metrics struct {
	//biz
	ConnectionOnlineGaugeVec *prometheus.GaugeVec

	//command
	ZTInfoCounterVec    *prometheus.CounterVec
	IsLeaderGaugeVec    *prometheus.GaugeVec
	LRangeSeekHistogram prometheus.Histogram
	GCKeysCounterVec    *prometheus.CounterVec

	//expire
	ExpireKeysTotal *prometheus.CounterVec

	//tikvGC
	TikvGCTotal *prometheus.CounterVec

	//command biz
	CommandCallHistogramVec *prometheus.HistogramVec
	TxnCommitHistogramVec   *prometheus.HistogramVec
	TxnRetriesCounterVec    *prometheus.CounterVec
	TxnConflictsCounterVec  *prometheus.CounterVec
	TxnFailuresCounterVec   *prometheus.CounterVec

	//logger
	LogMetricsCounterVec *prometheus.CounterVec
}

// init create global object
func init() {
	gm = &Metrics{}

	gm.CommandCallHistogramVec = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "command_duration_seconds",
			Buckets:   prometheus.ExponentialBuckets(0.0005, 2, 20),
			Help:      "The cost times of command call",
		}, multiLabel)
	prometheus.MustRegister(gm.CommandCallHistogramVec)

	gm.TxnRetriesCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "txn_retries_total",
			Help:      "The total of txn retries",
		}, multiLabel)
	prometheus.MustRegister(gm.TxnRetriesCounterVec)

	gm.TxnConflictsCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "txn_conflicts_total",
			Help:      "The total of txn conflicts",
		}, multiLabel)
	prometheus.MustRegister(gm.TxnConflictsCounterVec)

	gm.TxnCommitHistogramVec = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "txn_commit_seconds",
			Buckets:   prometheus.ExponentialBuckets(0.0005, 2, 20),
			Help:      "The cost times of txn commit",
		}, multiLabel)
	prometheus.MustRegister(gm.TxnCommitHistogramVec)

	gm.TxnFailuresCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "txn_failures_total",
			Help:      "The total of txn failures",
		}, multiLabel)
	prometheus.MustRegister(gm.TxnFailuresCounterVec)

	gm.ConnectionOnlineGaugeVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "connect_online_number",
			Help:      "The number of online connection",
		}, bizLabel)
	prometheus.MustRegister(gm.ConnectionOnlineGaugeVec)

	gm.LRangeSeekHistogram = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "lrange_seek_duration_seconds",
			Buckets:   prometheus.ExponentialBuckets(0.0005, 2, 20),
			Help:      "The cost times of list lrange seek",
		})
	prometheus.MustRegister(gm.LRangeSeekHistogram)

	gm.ZTInfoCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "zt_info_total",
			Help:      "zlist transfer worker summary",
		}, ztInfoLabel)
	prometheus.MustRegister(gm.ZTInfoCounterVec)

	gm.GCKeysCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "gc_keys_total",
			Help:      "the number of gc keys added or deleted",
		}, gcKeysLabel)
	prometheus.MustRegister(gm.GCKeysCounterVec)

	gm.ExpireKeysTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "expire_keys_total",
			Help:      "the number of expire keys added or expired",
		}, expireLabel)
	prometheus.MustRegister(gm.ExpireKeysTotal)

	gm.TikvGCTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "tikv_gc_total",
			Help:      "the number of tikv gc total by exec",
		}, tikvGCLabel)
	prometheus.MustRegister(gm.TikvGCTotal)

	gm.IsLeaderGaugeVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "is_leader",
			Help:      "mark titan is leader for gc/expire/zt/tikvgc",
		}, leaderLabel)
	prometheus.MustRegister(gm.IsLeaderGaugeVec)

	gm.LogMetricsCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "logs_entries_total",
			Help:      "Number of logs of certain level",
		},
		[]string{labelName},
	)
	prometheus.MustRegister(gm.LogMetricsCounterVec)
	sdk_metrics.RegisterMetrics()
}

// GetMetrics return metrics object
func GetMetrics() *Metrics {
	return gm
}

// MetricsHandle register the metrics handle
func MetricsHandle() {
	http.Handle("/metrics", prometheus.Handler())
}

// Measure logger level rate
func Measure(e zapcore.Entry) error {
	label := e.LoggerName + "_" + e.Level.String()
	gm.LogMetricsCounterVec.WithLabelValues(label).Inc()
	return nil
}
