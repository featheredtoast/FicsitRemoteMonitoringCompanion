package exporter_test

import (
	"context"
	"time"

	"github.com/AP-Hunt/FicsitRemoteMonitoringCompanion/Companion/exporter"
	"github.com/benbjohnson/clock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"
)

func expectGauge(metric *prometheus.GaugeVec, labels ...string) Assertion {
	val, err := gaugeValue(metric, labels...)
	Expect(err).ToNot(HaveOccurred())
	return Expect(val)
}

func eventuallyExpectGauge(metric *prometheus.GaugeVec, labels ...string) AsyncAssertion {
	val, err := gaugeValue(metric, labels...)
	Expect(err).ToNot(HaveOccurred())
	return Eventually(val)
}

var _ = Describe("RecordedMetricRegistration", func() {
	var ctx context.Context
	var cancel context.CancelFunc
	var metric *prometheus.GaugeVec
	var r *exporter.RecordedMetricsRegister
	var testTime *clock.Mock
	var globalReg *prometheus.Registry
	BeforeEach(func() {
		testTime = clock.NewMock()
		exporter.Clock = testTime
		ctx, cancel = context.WithCancel(context.Background())
		globalReg = prometheus.NewRegistry()
		metric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "test_metric",
			Help: "Test Metric",
		}, []string{"label1", "label2"})
		globalReg.MustRegister(metric)
		r = exporter.NewRecordedMetricsRegister(ctx, 2 * time.Second)
		go r.Start()

	})
	AfterEach(func() {
		globalReg.Unregister(metric)
		cancel()
	})

	It("records metric items like normal", func() {
		r.WithLabelValues(metric, "val1", "val2").Set(1)
		r.WithLabelValues(metric, "val1", "val2").Set(2)
		r.WithLabelValues(metric, "val1", "val3").Set(1)
		expectGauge(metric, "val1", "val2").To(Equal(2.0))
		expectGauge(metric, "val1", "val3").To(Equal(1.0))
	})

	It("drops old metrics", func() {
		r.WithLabelValues(metric, "val1", "val2").Set(1)
		testTime.Add(1 * time.Second)
		r.WithLabelValues(metric, "val1", "val3").Set(1)
		testTime.Add(5 * time.Second)
		r.WithLabelValues(metric, "val1", "val4").Set(1)
		expectGauge(metric, "val1", "val2").To(Equal(0.0))
		expectGauge(metric, "val1", "val3").To(Equal(0.0))
		expectGauge(metric, "val1", "val4").To(Equal(1.0))
		testTime.Add(1 * time.Second)
		r.WithLabelValues(metric, "val1", "val5").Set(1)
		testTime.Add(15 * time.Second)
		expectGauge(metric, "val1", "val4").To(Equal(0.0))
		expectGauge(metric, "val1", "val5").To(Equal(0.0))
	})

	It("allows metrics to get re-registered", func() {
		r.WithLabelValues(metric, "val1", "val2").Set(1)
		testTime.Add(5 * time.Second)
		r.WithLabelValues(metric, "val1", "val2").Set(1)
		r.WithLabelValues(metric, "val1", "val3").Set(1)
		expectGauge(metric, "val1", "val2").To(Equal(1.0))
	})

	It("allows metrics to get deleted by a partial match", func() {
		r.WithLabelValues(metric, "val1", "val2").Set(1)
		r.WithLabelValues(metric, "val1", "val3").Set(1)
		deleted := r.DeletePartialMatch(prometheus.Labels{"label1": "val1"})
		Expect(deleted).To(Equal(2))
	})
})
