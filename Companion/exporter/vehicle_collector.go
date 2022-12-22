package exporter

import (
	"log"
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
	DepartTime    time.Time
	Departed      bool
}

func (v *VehicleDetails) recordElapsedTime() {
	now := Clock.Now()
	tripSeconds := now.Sub(v.DepartTime).Seconds()
	VehicleRoundTrip.WithLabelValues(v.Id, v.VehicleType, v.PathName).Set(tripSeconds)
}

func (d *VehicleDetails) handleTimingUpdates(trackedVehicles map[string]*VehicleDetails) {
	if d.AutoPilot {
		vehicle, exists := trackedVehicles[d.Id]
		if exists && vehicle.Departed && vehicle.Location.isNearby(d.Location) && vehicle.Location.isSameDirection(d.Location) {
			// vehicle near first tracked location facing roughly the same way
			// record elapsed time.
			vehicle.recordElapsedTime()
			vehicle.Departed = false
		} else if exists && !vehicle.Departed && !vehicle.Location.isNearby(d.Location) {
			// vehicle departed from first tracked location - start counter
			vehicle.Departed = true
			vehicle.DepartTime = Clock.Now()
		} else if !exists && d.ForwardSpeed < 10 {
			// start tracking the vehicle at low speeds

			trackedVehicle := VehicleDetails{
				Id:            d.Id,
				Location: d.Location,
				VehicleType:   d.VehicleType,
				PathName:      d.PathName,
				Departed:      false,
			}
			trackedVehicles[d.Id] = &trackedVehicle
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

		d.handleTimingUpdates(c.TrackedVehicles)
	}
}
