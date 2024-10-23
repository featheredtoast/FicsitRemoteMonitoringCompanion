package exporter

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type CollectorRunner struct {
	collectors []Collector
	ctx        context.Context
	cancel     context.CancelFunc
	frmBaseUrl string
	saveName   string
}

type Collector interface {
	Collect(string, string)
}

func NewCollectorRunner(ctx context.Context, frmBaseUrl string, collectors ...Collector) *CollectorRunner {
	ctx, cancel := context.WithCancel(ctx)
	return &CollectorRunner{
		ctx:        ctx,
		cancel:     cancel,
		collectors: collectors,
		frmBaseUrl: frmBaseUrl,
		saveName:   "default",
	}
}

func (c *CollectorRunner) updateSaveName() {
	//TODO: update save name
	newSaveName := "default"
	if newSaveName != c.saveName {
		MetricsRegister.DeletePartialMatch(prometheus.Labels{"url": c.frmBaseUrl, "save_name": c.saveName})
		c.saveName = newSaveName
	}
}

func (c *CollectorRunner) Start() {
	c.updateSaveName()
	c.Collect(c.frmBaseUrl, c.saveName)
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-Clock.After(5 * time.Second):
			c.updateSaveName()
			c.Collect(c.frmBaseUrl, c.saveName)
		}
	}
}

func (c *CollectorRunner) Stop() {
	c.cancel()
}

func (c *CollectorRunner) Collect(server string, saveName string) {
	for _, collector := range c.collectors {
		collector.Collect(server, saveName)
	}
}
