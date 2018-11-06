package metrics

import (
	"encoding/json"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
)

//MetricsType object type
type MetricsType int

const (
	ConnectionOnlineType MetricsType = iota
	ZTInfoType
	IsLeaderType
	LRangeSeekType
	CommandCallType
	TransactionCommitType
	TransactionRollbackType
	TransactionConflictType
	TransactionFailureType
)

//MetricsTypeValue export metric msg
var MetricsTypeValue = map[MetricsType]string{
	ConnectionOnlineType:    "ConnectionOnlineGaugeVec",
	ZTInfoType:              "ZTInfoCounterVec",
	IsLeaderType:            "IsLeaderGaugeVec",
	LRangeSeekType:          "LRangeSeekHistogramVec",
	CommandCallType:         "CommandTransferHistogram",
	TransactionCommitType:   "TransactionCommitHistogramVec",
	TransactionRollbackType: "TransactionRollbackGaugeVec",
	TransactionConflictType: "TransactionConflictGauageVec",
	TransactionFailureType:  "TransactionFailureGaugeVec",
}

const (
	//promethus default namespace
	namespace = "titan"

	//promethues default label key
	command = "command"
	biz     = "biz"
	leader  = "leader"
	ztinfo  = "ztinfo"
)

var (
	//Label value slice when creating prometheus object
	commandLabel = []string{command}
	bizLabel     = []string{biz}
	leaderLabel  = []string{leader}
	multiLabel   = []string{biz, command}
	ztInfoLabel  = []string{ztinfo}

	// global prometheus object
	gm *Metrics
)

//Metrics
type Metrics struct {
	//biz
	ConnectionOnlineGaugeVec *prometheus.GaugeVec

	//command
	ZTInfoCounterVec    *prometheus.CounterVec
	IsLeaderGaugeVec    *prometheus.GaugeVec
	LRangeSeekHistogram prometheus.Histogram

	//command biz
	CommandCallHistogramVec       *prometheus.HistogramVec
	TransactionCommitHistogramVec *prometheus.HistogramVec
	TransactionRollbackGaugeVec   *prometheus.GaugeVec
	TransactionConflictGauageVec  *prometheus.GaugeVec
	TransactionFailureGaugeVec    *prometheus.GaugeVec
}

//init create global object
func init() {
	gm = &Metrics{}

	gm.CommandCallHistogramVec = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "command_call_times",
			Buckets:   prometheus.ExponentialBuckets(0.0005, 2, 20),
			Help:      "The cost times of command call",
		}, multiLabel)
	prometheus.MustRegister(gm.CommandCallHistogramVec)

	gm.TransactionRollbackGaugeVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "transaction_rollback_count",
			Help:      "The count of transaction rollback",
		}, multiLabel)
	prometheus.MustRegister(gm.TransactionRollbackGaugeVec)

	gm.TransactionConflictGauageVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "transaction_conflict_count",
			Help:      "The count of transaction conflict",
		}, multiLabel)
	prometheus.MustRegister(gm.TransactionConflictGauageVec)

	gm.TransactionCommitHistogramVec = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "transaction_commit_times",
			Buckets:   prometheus.ExponentialBuckets(0.0005, 2, 20),
			Help:      "The cost times of transaction commit",
		}, multiLabel)
	prometheus.MustRegister(gm.TransactionCommitHistogramVec)

	gm.TransactionFailureGaugeVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "transaction_failure_count",
			Help:      "The count of transaction failure",
		}, multiLabel)
	prometheus.MustRegister(gm.TransactionFailureGaugeVec)

	gm.ConnectionOnlineGaugeVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "connect_online_count",
			Help:      "The count of online connection",
		}, bizLabel)
	prometheus.MustRegister(gm.ConnectionOnlineGaugeVec)

	gm.LRangeSeekHistogram = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "lrange_seek_times",
			Buckets:   prometheus.ExponentialBuckets(0.0005, 2, 20),
			Help:      "The cost times of list lrange seek",
		})
	prometheus.MustRegister(gm.LRangeSeekHistogram)

	gm.ZTInfoCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "zt_info",
			Help:      "zlist transfer worker summary",
		}, ztInfoLabel)
	prometheus.MustRegister(gm.ZTInfoCounterVec)

	gm.IsLeaderGaugeVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "is_leader",
			Help:      "mark titan is leader for gc/expire/zt",
		}, leaderLabel)
	prometheus.MustRegister(gm.IsLeaderGaugeVec)

	http.Handle("/titan/gm", prometheus.Handler())
}

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
