package exporter

import (
	"log"
)

type ProductionCollector struct {
	endpoint string
}

type ProductionDetails struct {
	ItemName           string  `json:"Name"`
	ProdPercent        float64 `json:"ProdPercent"`
	ConsPercent        float64 `json:"ConsPercent"`
	CurrentProduction  float64 `json:"CurrentProd"`
	CurrentConsumption float64 `json:"CurrentConsumed"`
	MaxProd            float64 `json:"MaxProd"`
	MaxConsumed        float64 `json:"MaxConsumed"`
}

func NewProductionCollector(endpoint string) *ProductionCollector {
	return &ProductionCollector{
		endpoint: endpoint,
	}
}

func (c *ProductionCollector) Collect(frmAddress string, sessionName string) {
	details := []ProductionDetails{}
	err := retrieveData(frmAddress+c.endpoint, &details)
	if err != nil {
		log.Printf("error reading production statistics from FRM: %s\n", err)
		return
	}

	for _, d := range details {
		GaugeWithLabelValues(ItemsProducedPerMin, d.ItemName, frmAddress, sessionName).Set(d.CurrentProduction)
		GaugeWithLabelValues(ItemsConsumedPerMin, d.ItemName, frmAddress, sessionName).Set(d.CurrentConsumption)

		GaugeWithLabelValues(ItemProductionCapacityPercent, d.ItemName, frmAddress, sessionName).Set(d.ProdPercent)
		GaugeWithLabelValues(ItemConsumptionCapacityPercent, d.ItemName, frmAddress, sessionName).Set(d.ConsPercent)
		GaugeWithLabelValues(ItemProductionCapacityPerMinute, d.ItemName, frmAddress, sessionName).Set(d.MaxProd)
		GaugeWithLabelValues(ItemConsumptionCapacityPerMinute, d.ItemName, frmAddress, sessionName).Set(d.MaxConsumed)
	}
}
