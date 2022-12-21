package exporter_test

import (
	"context"

	"github.com/AP-Hunt/FicsitRemoteMonitoringCompanion/m/v2/exporter"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"time"
)

type TestCollector struct {
	counter int
}

func NewTestCollector() *TestCollector {
	return &TestCollector{
		counter: 0,
	}
}
func (t *TestCollector) Collect() {
	t.counter = t.counter + 1
}

var _ = Describe("CollectorRunner", func() {
	Describe("Basic Functionality", func() {
		It("runs on init and on each timeout", func() {
			ctx, cancel := context.WithCancel(context.Background())
			run := make(chan time.Time)
			exporter.AfterInterval = func(d time.Duration) <-chan time.Time {
				return run
			}

			c1 := NewTestCollector()
			c2 := NewTestCollector()
			runner := exporter.NewCollectorRunner(ctx, c1, c2)
			go runner.Start()
			run <- time.Now()
			cancel()
			Expect(c1.counter).To(Equal(2))
			Expect(c2.counter).To(Equal(2))
		})
	})
})
