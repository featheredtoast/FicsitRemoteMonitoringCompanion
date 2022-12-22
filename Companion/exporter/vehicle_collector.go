package exporter

import (
	"log"
	"math"
	"time"
)

type VehicleCollector struct {
	FRMAddress      string
	TrackedVehicles map[string]*VehicleDetails
}

type VehicleDetails struct {
	Id            string   `json:"ID"`
	VehicleType   string   `json:"VehicleType"`
	Location      Location `json:"location"`
	ForwardSpeed  float64  `json:"ForwardSpeed"`
	AutoPilot     bool     `json:"AutoPilot"`
	FuelType      string   `json:"FuelType"`
	FuelInventory float64  `json"FuelInventory"`
	PathName      string   `json:"PathName"`
	StartLocation Location
	DepartTime    time.Time
	Departed      bool
}

// Calculates if a location is nearby another.
// From observation, 5000 units is "good enough" to be considered nearby.
func (l *Location) isNearby(other Location) bool {
	x := l.X - other.X
	y := l.Y - other.Y
	z := l.Z - other.Z

	dist := math.Sqrt(math.Pow(x, 2) + math.Pow(y, 2) + math.Pow(z, 2))
	return dist <= 5000
}

// Calculates if this location is roughly facing the same way as another
func (l *Location) isSameDirection(other Location) bool {
	diff := math.Abs(float64(l.Rotation - other.Rotation))
	return diff <= 90
}

func (v *VehicleDetails) recordElapsedTime() {
	now := Now()
	tripSeconds := now.Sub(v.DepartTime).Seconds()
	VehicleRoundTrip.WithLabelValues(v.Id, v.VehicleType, v.PathName).Set(tripSeconds)
	v.Departed = false
}

func (d *VehicleDetails) handleTimingUpdates(trackedVehicles map[string]*VehicleDetails) {
	if d.AutoPilot {
		vehicle, exists := trackedVehicles[d.Id]
		if exists && vehicle.Departed && vehicle.Location.isNearby(d.Location) && vehicle.Location.isSameDirection(d.Location) {
			// vehicle arrived at a nearby location facing around the same way.
			// record elapsed time.
			vehicle.recordElapsedTime()
		} else if exists && !vehicle.StartLocation.isNearby(d.Location) {
			// vehicle departed - start counter
			vehicle.Departed = true
			vehicle.DepartTime = Now()
		} else if d.ForwardSpeed < 10 {
			// start tracking the vehicle at low speeds
			d.StartLocation = d.Location
			trackedVehicles[d.Id] = d
		}
	} else {
		//remove manual vehicles, nothing to mark
		_, exists := trackedVehicles[d.Id]
		if exists {
			delete(trackedVehicles, d.Id)
		}
	}
}

func NewVehicleCollector(frmAddress string) *VehicleCollector {
	return &VehicleCollector{
		FRMAddress:      frmAddress,
		TrackedVehicles: make(map[string]*VehicleDetails),
	}
}

func (c *VehicleCollector) Collect() {
	details := []VehicleDetails{}
	err := retrieveData(c.FRMAddress, &details)
	if err != nil {
		log.Printf("error reading vehicle statistics from FRM: %s\n", err)
		return
	}

	for _, d := range details {
		VehicleFuel.WithLabelValues(d.Id, d.VehicleType, d.FuelType).Set(d.FuelInventory)

		// TODO: round trip caluclations
		d.handleTimingUpdates(c.TrackedVehicles)
	}
}
