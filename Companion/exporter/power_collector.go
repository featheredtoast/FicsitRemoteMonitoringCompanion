package exporter

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

type PowerCollector struct {
	FRMAddress string
}

type PowerDetails struct {
	CircuitId           float64 `json:"CircuitID"`
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

func (pd *PowerDetails) parseBatteryTimeEmptySeconds() *float64 {
	matched, params := parseTimeSeconds(pd.BatteryTimeEmpty)
	if !matched {
		return nil
	}
	return &params
}

func (pd *PowerDetails) parseBatteryTimeFullSeconds() *float64 {
	matched, params := parseTimeSeconds(pd.BatteryTimeFull)
	if !matched {
		return nil
	}
	return &params
}

func (pd *PowerDetails) parseFuseTriggered() float64 {
	if pd.FuseTriggered {
		return 1
	} else {
		return 0
	}
}

func NewPowerCollector(frmAddress string) *PowerCollector {
	return &PowerCollector{
		FRMAddress: frmAddress,
	}
}

func (c *PowerCollector) Collect() {
	resp, err := http.Get(c.FRMAddress)

	if err != nil {
		log.Printf("error fetching power statistics from FRM: %s\n", err)
		return
	}

	defer resp.Body.Close()

	details := []PowerDetails{}
	decoder := json.NewDecoder(resp.Body)

	err = decoder.Decode(&details)
	if err != nil {
		log.Printf("error reading power statistics from FRM: %s\n", err)
		return
	}

	for _, d := range details {
		circuitId := strconv.FormatFloat(d.CircuitId, 'f', -1, 64)
		PowerConsumed.WithLabelValues(circuitId).Set(d.PowerConsumed)
		PowerCapacity.WithLabelValues(circuitId).Set(d.PowerCapacity)
		PowerMaxConsumed.WithLabelValues(circuitId).Set(d.PowerMaxConsumed)
		BatteryDifferential.WithLabelValues(circuitId).Set(d.BatteryDifferential)
		BatteryPercent.WithLabelValues(circuitId).Set(d.BatteryPercent)
		BatteryCapacity.WithLabelValues(circuitId).Set(d.BatteryCapacity)
		batterySecondsRemaining := d.parseBatteryTimeEmptySeconds()
		if batterySecondsRemaining != nil {
			BatterySecondsEmpty.WithLabelValues(circuitId).Set(*batterySecondsRemaining)
		}
		batterySecondsFull := d.parseBatteryTimeFullSeconds()
		if batterySecondsFull != nil {
			BatterySecondsFull.WithLabelValues(circuitId).Set(*batterySecondsFull)
		}
		fuseTriggered := d.parseFuseTriggered()
		FuseTriggered.WithLabelValues(circuitId).Set(fuseTriggered)
	}
}
