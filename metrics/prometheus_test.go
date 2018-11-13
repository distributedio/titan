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

	gm.TxnConflictsCounterVec.WithLabelValues(defaultLabel, defaultlabel).Inc()
	gm.TxnConflictsCounterVec.WithLabelValues(defaultLabel, defaultlabel).Inc()
	gm.TxnRetriesCounterVec.WithLabelValues(defaultLabel, defaultlabel).Inc()
	gm.TxnFailuresCounterVec.WithLabelValues(defaultLabel, defaultlabel).Desc()
	gm.TxnCommitHistogramVec.WithLabelValues(defaultLabel, defaultlabel).Desc()
	gm.CommandCallHistogramVec.WithLabelValues(defaultLabel, defaultlabel).Desc()

	gm.LRangeSeekHistogram.Desc()
	gm.LogMetricsCounterVec.WithLabelValues("INFO").Inc()
}
