package logbunny

import (
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap/zapcore"
)

const labelName string = "level"

var logMetrics *prometheus.CounterVec

func init() {
	logMetrics = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "logbunny_logs_total",
			Help: "Number of logs of certain level",
		},
		[]string{labelName},
	)
	prometheus.MustRegister(logMetrics)
}

func measure(e zapcore.Entry) error {
	label := e.LoggerName + "_" + e.Level.String()
	logMetrics.WithLabelValues(label).Inc()
	return nil
}
