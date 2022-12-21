package exporter

import (
	"log"
)

type TrainCollector struct {
	FRMAddress string
}

type TimeTable struct {
	StationName string `json:"StationName"`
}
type TrainDetails struct {
	TrainName     string    `json:"TrainName"`
	PowerConsumed float64   `json:"PowerConsumed"`
	TrainStation  string    `json:"TrainStation"`
	Derailed      bool      `json:"Derailed"`
	Status        string    `json:"Status"` //"TS_SelfDriving",
	TimeTable     []TimeTable `json:"TimeTable"`
}

func NewTrainCollector(frmAddress string) *TrainCollector {
	return &TrainCollector{
		FRMAddress: frmAddress,
	}
}

func (c *TrainCollector) Collect() {
	details := []TrainDetails{}
	err := retrieveData(c.FRMAddress, &details)
	if err != nil {
		log.Printf("error reading train statistics from FRM: %s\n", err)
		return
	}

	for _, d := range details {
		TrainPower.WithLabelValues(d.TrainName).Set(d.PowerConsumed)

		isDerailed := parseBool(d.Derailed)
		TrainDerailed.WithLabelValues(d.TrainName).Set(isDerailed)

		//TODO: calculate round trip and segment trip time
	}
}
