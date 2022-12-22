package exporter

import (
	"log"
	"time"
)

type TrainCollector struct {
	FRMAddress    string
	TrackedTrains map[string]*TrainDetails
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
	StationArrivalTimestamp      time.Time
	StationCounter               int
	FirstStationArrivalTimestamp time.Time
}

func NewTrainCollector(frmAddress string) *TrainCollector {
	return &TrainCollector{
		FRMAddress:    frmAddress,
		TrackedTrains: make(map[string]*TrainDetails),
	}
}

func (t *TrainDetails) markNextStation(d *TrainDetails) {
	if t.TrainStation != d.TrainStation {
		t.StationCounter = t.StationCounter + 1
		now := Clock.Now()
		tripSeconds := now.Sub(t.StationArrivalTimestamp).Seconds()
		TrainSegmentTrip.WithLabelValues(t.TrainName, t.TrainStation, d.TrainStation).Set(tripSeconds)
		if len(t.TimeTable) <= t.StationCounter {
			roundTripSeconds := now.Sub(t.FirstStationArrivalTimestamp).Seconds()
			TrainRoundTrip.WithLabelValues(t.TrainName).Set(roundTripSeconds)
			t.StationCounter = 0
			t.FirstStationArrivalTimestamp = now
		}
		t.StationArrivalTimestamp = now
		t.TrainStation = d.TrainStation
	}
}

func (t *TrainDetails) markFirstStation(d *TrainDetails) {
	if t.TrainStation != d.TrainStation {
		t.StationCounter = 0
		t.FirstStationArrivalTimestamp = Clock.Now()
		t.StationArrivalTimestamp = Clock.Now()
		t.TrainStation = d.TrainStation
	}
}

func (d *TrainDetails) handleTimingUpdates(trackedTrains map[string]*TrainDetails) {
	// add or update prev station and timestamp for automatic trains
	if d.Status == "TS_SelfDriving" {
		train, exists := trackedTrains[d.TrainName]
		if exists && !train.FirstStationArrivalTimestamp.IsZero() {
			train.markNextStation(d)
		} else if exists {
			train.markFirstStation(d)
		} else if !exists {
			trackedTrain := TrainDetails{
				TrainName:      d.TrainName,
				TrainStation:   d.TrainStation,
				StationCounter: 0,
				TimeTable:      d.TimeTable,
			}
			trackedTrains[d.TrainName] = &trackedTrain
		}
	} else {
		//remove manual trains, nothing to mark
		_, exists := trackedTrains[d.TrainName]
		if exists {
			delete(trackedTrains, d.TrainName)
		}
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

		d.handleTimingUpdates(c.TrackedTrains)
	}
}
