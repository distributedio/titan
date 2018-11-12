package metrics

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap/zapcore"

	"github.com/prometheus/client_golang/prometheus"
)

//MetricsType object type
type metricsType int

const (
	ConnectionOnlineType metricsType = iota
	ZTInfoType
	IsLeaderType
	LRangeSeekType
	CommandCallType
	RecycleInfoType
	TxnCommitType
	TxnRetriesType
	TxnConflictsType
	TxnFailuresType
	LogMetricsType
)

//MetricsTypeValue export metric msg
var MetricsTypeValue = map[metricsType]string{
	ConnectionOnlineType: "ConnectionOnlineGaugeVec",
	ZTInfoType:           "ZTInfoCounterVec",
	IsLeaderType:         "IsLeaderGaugeVec",
	LRangeSeekType:       "LRangeSeekHistogramVec",
	CommandCallType:      "CommandTransferHistogram",
	RecycleInfoType:      "GCInfoCounterType",
	TxnCommitType:        "TxnCommitHistogramVec",
	TxnRetriesType:       "TxnRetriesCounterVec",
	TxnConflictsType:     "TxnConflictsCounterVec",
	TxnFailuresType:      "TxnFailuresCounterVec",
	LogMetricsType:       "LogMetrics",
}

const (
	//promethus default namespace
	namespace = "thanos"

	//promethues default label key
	command     = "command"
	biz         = "biz"
	leader      = "leader"
	ztinfo      = "ztinfo"
	labelName   = "level"
	recycleinfo = "recycleinfo"
)

var (
	//Label value slice when creating prometheus object
	commandLabel     = []string{command}
	bizLabel         = []string{biz}
	leaderLabel      = []string{leader}
	multiLabel       = []string{biz, command}
	ztInfoLabel      = []string{ztinfo}
	recycleInfoLabel = []string{recycleinfo}

	// global prometheus object
	gm *Metrics
)

//Metrics prometheus statistics
type Metrics struct {
	//biz
	ConnectionOnlineGaugeVec *prometheus.GaugeVec

	//command
	ZTInfoCounterVec      *prometheus.CounterVec
	IsLeaderGaugeVec      *prometheus.GaugeVec
	LRangeSeekHistogram   prometheus.Histogram
	RecycleInfoCounterVec *prometheus.CounterVec

	//command biz
	CommandCallHistogramVec *prometheus.HistogramVec
	TxnCommitHistogramVec   *prometheus.HistogramVec
	TxnRetriesCounterVec    *prometheus.CounterVec
	TxnConflictsCounterVec  *prometheus.CounterVec
	TxnFailuresCounterVec   *prometheus.CounterVec

	//logger
	LogMetricsCounterVec *prometheus.CounterVec
}

//init create global object
func init() {
	gm = &Metrics{}

	gm.CommandCallHistogramVec = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "thanos_command_call_second",
			Buckets:   prometheus.ExponentialBuckets(0.0005, 2, 20),
			Help:      "The cost times of command call",
		}, multiLabel)
	prometheus.MustRegister(gm.CommandCallHistogramVec)

	gm.TxnRetriesCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "thanos_txn_retries_total",
			Help:      "The total of txn retries",
		}, multiLabel)
	prometheus.MustRegister(gm.TxnRetriesCounterVec)

	gm.TxnConflictsCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "thanos_txn_conflicts_total",
			Help:      "The total of txn conflicts",
		}, multiLabel)
	prometheus.MustRegister(gm.TxnConflictsCounterVec)

	gm.TxnCommitHistogramVec = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "thanos_txn_commit_second",
			Buckets:   prometheus.ExponentialBuckets(0.0005, 2, 20),
			Help:      "The cost times of txn commit",
		}, multiLabel)
	prometheus.MustRegister(gm.TxnCommitHistogramVec)

	gm.TxnFailuresCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "thanos_txn_failures_total",
			Help:      "The total of txn failures",
		}, multiLabel)
	prometheus.MustRegister(gm.TxnFailuresCounterVec)

	gm.ConnectionOnlineGaugeVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "thanos_connect_online_number",
			Help:      "The number of online connection",
		}, bizLabel)
	prometheus.MustRegister(gm.ConnectionOnlineGaugeVec)

	gm.LRangeSeekHistogram = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "thanos_lrange_seek_second",
			Buckets:   prometheus.ExponentialBuckets(0.0005, 2, 20),
			Help:      "The cost times of list lrange seek",
		})
	prometheus.MustRegister(gm.LRangeSeekHistogram)

	gm.ZTInfoCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "thanos_zt_info",
			Help:      "zlist transfer worker summary",
		}, ztInfoLabel)
	prometheus.MustRegister(gm.ZTInfoCounterVec)

	gm.RecycleInfoCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "thanos_recycle_info",
			Help:      "the number of recycle data",
		}, ztInfoLabel)
	prometheus.MustRegister(gm.RecycleInfoCounterVec)

	gm.IsLeaderGaugeVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "thanos_is_leader",
			Help:      "mark titan is leader for gc/expire/zt",
		}, leaderLabel)
	prometheus.MustRegister(gm.IsLeaderGaugeVec)

	gm.LogMetricsCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "thanos_logs_entries_total",
			Help:      "Number of logs of certain level",
		},
		[]string{labelName},
	)

	http.Handle("/thanos/metrics", prometheus.Handler())
}

//GetMetrics return metrics object
func GetMetrics() *Metrics {
	return gm
}

//String export gm msg
func (mt *Metrics) String() string {
	if msg, err := json.Marshal(MetricsTypeValue); err != nil {
		return string(msg)
	}
	return ""
}

//Measure logger level rate
func Measure(e zapcore.Entry) error {
	label := e.LoggerName + "_" + e.Level.String()
	gm.LogMetricsCounterVec.WithLabelValues(label).Inc()
	return nil
}
