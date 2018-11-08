package metrics

import (
	"net/http"
	"testing"
)

const (
	defaultLabel = "test"
	defaultlabel = "label"
)

func TestMetrics(t *testing.T) {

	go func() {
		http.ListenAndServe(":8888", nil)
	}()

	gm.IsLeaderGaugeVec.WithLabelValues(defaultlabel).Inc()

	gm.ConnectionOnlineGaugeVec.WithLabelValues(defaultLabel).Inc()

	gm.TransactionConflictGauageVec.WithLabelValues(defaultLabel, defaultlabel).Inc()
	gm.TransactionConflictGauageVec.WithLabelValues(defaultLabel, defaultlabel).Inc()
	gm.TransactionRollbackGaugeVec.WithLabelValues(defaultLabel, defaultlabel).Inc()
	gm.TransactionFailureGaugeVec.WithLabelValues(defaultLabel, defaultlabel).Desc()
	gm.TransactionCommitHistogramVec.WithLabelValues(defaultLabel, defaultlabel).Desc()
	gm.CommandCallHistogramVec.WithLabelValues(defaultLabel, defaultlabel).Desc()

	gm.LRangeSeekHistogram.Desc()
	gm.LogMetricsCounterVec.WithLabelValues("INFO").Inc()
}
