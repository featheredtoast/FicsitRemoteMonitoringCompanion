package exporter

import (
	"log"
	"time"
)

type TrainCollector struct {
	FRMAddress string
}

type TimeTable struct {
	StationName string `json:"StationName"`
}
type TrainDetails struct {
	TrainName                    string      `json:"TrainName"`
	PowerConsumed                float64     `json:"PowerConsumed"`
	TrainStation                 string      `json:"TrainStation"`
	Derailed                     bool        `json:"Derailed"`
	Status                       string      `json:"Status"` //"TS_SelfDriving",
	TimeTable                    []TimeTable `json:"TimeTable"`
	StationArrivalTimestamp      *time.Time
	StationCounter               int
	FirstStationArrivalTimestamp *time.Time
}

var TrackedTrains map[string]*TrainDetails

func NewTrainCollector(frmAddress string) *TrainCollector {
	return &TrainCollector{
		FRMAddress: frmAddress,
	}
}

func (t *TrainDetails) markNextStation(d *TrainDetails) {
	if t.TrainStation != d.TrainStation {
		now := Now()
		tripSeconds := time.Since(*t.StationArrivalTimestamp).Seconds()
		TrainSegmentTrip.WithLabelValues(t.TrainName, t.TrainStation, d.TrainStation).Set(tripSeconds)
		if len(t.TimeTable) <= t.StationCounter {
			roundTripSeconds := time.Since(*t.FirstStationArrivalTimestamp).Seconds()
			TrainRoundTrip.WithLabelValues(t.TrainName).Set(roundTripSeconds)
			t.StationCounter = 0
			t.FirstStationArrivalTimestamp = &now
		} else {
			t.StationCounter = t.StationCounter + 1
		}
		t.StationArrivalTimestamp = &now
		t.TrainStation = d.TrainStation
	}
}

func (t *TrainDetails) markFirstStation(d *TrainDetails) {
	if t.TrainStation != d.TrainStation {
		t.StationCounter = 0
		now := Now()
		t.FirstStationArrivalTimestamp = &now
		t.TrainStation = d.TrainStation
	}
}

func (t *TrainDetails) markFirstSeen() {
	t.StationCounter = 0
	TrackedTrains[t.TrainName] = t
}

func (d *TrainDetails) handleTimingUpdates() {
	// add or update prev station and timestamp for automatic trains
	if d.Status == "TS_SelfDriving" {
		train, exists := TrackedTrains[d.TrainName]
		if exists && train.FirstStationArrivalTimestamp != nil {
			train.markNextStation(d)
		} else if exists {
			train.markFirstStation(d)
		} else {
			d.markFirstSeen()
		}
	} else {
		//remove manual trains, nothing to mark
		_, exists := TrackedTrains[d.TrainName]
		if exists {
			delete(TrackedTrains, d.TrainName)
		}
	}
}

func (c *TrainCollector) Collect() {
	TrackedTrains = make(map[string]*TrainDetails)
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
		d.handleTimingUpdates()
	}
}
