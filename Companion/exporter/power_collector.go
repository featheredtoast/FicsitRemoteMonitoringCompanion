package exporter

import (
	"log"
	"strconv"
)

type PowerInfo struct {
	CircuitGroupId float64 `json:"CircuitGroupID"`
	PowerConsumed  float64 `json:"PowerConsumed"`
}

type PowerCollector struct {
	endpoint string
}

type PowerDetails struct {
	CircuitGroupId      float64 `json:"CircuitGroupID"`
	PowerConsumed       float64 `json:"PowerConsumed"`
	PowerCapacity       float64 `json:"PowerCapacity"`
	PowerMaxConsumed    float64 `json:"PowerMaxConsumed"`
	BatteryDifferential float64 `json:"BatteryDifferential"`
	BatteryPercent      float64 `json:"BatteryPercent"`
	BatteryCapacity     float64 `json:"BatteryCapacity"`
	BatteryTimeEmpty    string  `json:"BatteryTimeEmpty"`
	BatteryTimeFull     string  `json:"BatteryTimeFull"`
	FuseTriggered       bool    `json:"FuseTriggered"`
}

func NewPowerCollector(endpoint string) *PowerCollector {
	return &PowerCollector{
		endpoint: endpoint,
	}
}

func (c *PowerCollector) Collect(frmAddress string, saveName string) {
	details := []PowerDetails{}
	err := retrieveData(frmAddress+c.endpoint, &details)
	if err != nil {
		log.Printf("error reading power statistics from FRM: %s\n", err)
		return
	}

	for _, d := range details {
		circuitGroupId := strconv.FormatFloat(d.CircuitGroupId, 'f', -1, 64)
		GaugeWithLabelValues(PowerConsumed, circuitGroupId, frmAddress, saveName).Set(d.PowerConsumed)
		GaugeWithLabelValues(PowerCapacity, circuitGroupId, frmAddress, saveName).Set(d.PowerCapacity)
		GaugeWithLabelValues(PowerMaxConsumed, circuitGroupId, frmAddress, saveName).Set(d.PowerMaxConsumed)
		GaugeWithLabelValues(BatteryDifferential, circuitGroupId, frmAddress, saveName).Set(d.BatteryDifferential)
		GaugeWithLabelValues(BatteryPercent, circuitGroupId, frmAddress, saveName).Set(d.BatteryPercent)
		GaugeWithLabelValues(BatteryCapacity, circuitGroupId, frmAddress, saveName).Set(d.BatteryCapacity)
		batterySecondsRemaining := parseTimeSeconds(d.BatteryTimeEmpty)
		if batterySecondsRemaining != nil {
			GaugeWithLabelValues(BatterySecondsEmpty, circuitGroupId, frmAddress, saveName).Set(*batterySecondsRemaining)
		}
		batterySecondsFull := parseTimeSeconds(d.BatteryTimeFull)
		if batterySecondsFull != nil {
			GaugeWithLabelValues(BatterySecondsFull, circuitGroupId, frmAddress, saveName).Set(*batterySecondsFull)
		}
		fuseTriggered := parseBool(d.FuseTriggered)
		GaugeWithLabelValues(FuseTriggered, circuitGroupId, frmAddress, saveName).Set(fuseTriggered)
	}
}
