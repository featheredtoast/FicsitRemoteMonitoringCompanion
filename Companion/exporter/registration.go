package exporter

import (
	"context"
	"slices"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type MetricVectorDetails struct {
	Name   string
	Help   string
	Labels []string
}

var RegisteredMetricVectors = []MetricVectorDetails{}
var RegisteredMetrics = []*prometheus.GaugeVec{}
var MetricsRegister *RecordedMetricsRegister = nil

func GaugeWithLabelValues(metric *prometheus.GaugeVec, labelValues ...string) prometheus.Gauge {
	if (MetricsRegister != nil) {
		return MetricsRegister.WithLabelValues(metric, labelValues...)
	}
	return metric.WithLabelValues(labelValues...)
}

func RegisterNewGaugeVec(opts prometheus.GaugeOpts, labelNames []string) *prometheus.GaugeVec {
	// All metrics include url and session_name labels
	labelNames = append(labelNames, "url", "session_name")
	RegisteredMetricVectors = append(RegisteredMetricVectors, MetricVectorDetails{
		Name:   opts.Name,
		Help:   opts.Help,
		Labels: labelNames,
	})

	metric := promauto.NewGaugeVec(opts, labelNames)
	RegisteredMetrics = append(RegisteredMetrics, metric)
	return metric
}

// Item for metric detail: a gauge of a gauge vec
// Records metric, labels, and the last time it was updated
type RecordedMetricDetail struct {
	Metric      *prometheus.GaugeVec
	LabelValues []string
	LastUsed    time.Time
	DropLabels  prometheus.Labels
	DropData chan int
}

// A register for gauge vec metrics
// Allows for gauge vectors to delete stale labels from being recorded into prometheus
// updated metrics will appear as 'fresh' and will not be deleted.
type RecordedMetricsRegister struct {
	Ctx                   context.Context
	MetricsEvents         chan RecordedMetricDetail
	RecordedMetricDetails []RecordedMetricDetail

	//time in seconds until a gauge for a set of labels is considered stale and safe to remove
	TimeUntilStale time.Duration
}

// a call to GaugeVec.WithLabelValues, but updates the current set of label values as fresh
func (r *RecordedMetricsRegister) WithLabelValues(metric *prometheus.GaugeVec, labelValues ...string) prometheus.Gauge {
	r.MetricsEvents <- RecordedMetricDetail{Metric: metric, LabelValues: labelValues, LastUsed: Clock.Now()}
	return metric.WithLabelValues(labelValues...)
}

func (r *RecordedMetricsRegister) DeletePartialMatch(labels prometheus.Labels) int{
	dropData := make(chan int)
	r.MetricsEvents <- RecordedMetricDetail{DropLabels: labels, DropData: dropData}
	return <-dropData
}

func (r *RecordedMetricsRegister) dropStale() {
	staleTime := Clock.Now().Add(time.Duration(-r.TimeUntilStale))
	r.RecordedMetricDetails = slices.DeleteFunc(r.RecordedMetricDetails, func(record RecordedMetricDetail) bool {
		if record.LastUsed.Before(staleTime) {
			record.Metric.DeleteLabelValues(record.LabelValues...)
			return true
		}
		return false
	})
}

func (r *RecordedMetricsRegister) recordMetric(metricDetail RecordedMetricDetail) {

	if len(metricDetail.DropLabels) > 0 {
		val := 0
		for _, record := range r.RecordedMetricDetails {
			val += record.Metric.DeletePartialMatch(metricDetail.DropLabels)
		}
		metricDetail.DropData <-val
	} else if metricDetail.Metric != nil {
		r.RecordedMetricDetails = slices.DeleteFunc(r.RecordedMetricDetails, func(record RecordedMetricDetail) bool {
			return record.Metric == metricDetail.Metric && slices.Equal(record.LabelValues, metricDetail.LabelValues)
		})
		r.RecordedMetricDetails = append(r.RecordedMetricDetails, metricDetail)
	}
}

func (r *RecordedMetricsRegister) Start() error {
	timeout := Clock.After(5 * time.Second)
	for {
		select {
		case <-timeout:
			r.dropStale()
			timeout = Clock.After(5 * time.Second)
		case metricDetail := <-r.MetricsEvents:
			r.recordMetric(metricDetail)
		case <-r.Ctx.Done():
			return r.Ctx.Err()
		}
	}
}

func NewRecordedMetricsRegister(ctx context.Context, timeUntilStale time.Duration) *RecordedMetricsRegister {
	r := &RecordedMetricsRegister{
		Ctx:                   ctx,
		MetricsEvents:         make(chan RecordedMetricDetail, 100),
		RecordedMetricDetails: []RecordedMetricDetail{},
		TimeUntilStale:        timeUntilStale,
	}
	return r
}
