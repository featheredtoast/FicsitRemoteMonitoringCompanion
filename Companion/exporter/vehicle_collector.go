package exporter

import (
	"log"
)

type VehicleCollector struct {
	FRMAddress string
}

type VehicleDetails struct {
	Id            string  `json:"ID"`
	VehicleType   string  `json:"VehicleType"`
	AutoPilot     bool    `json:"AutoPilot"`
	FuelType      string  `json:"FuelType"`
	FuelInventory float64 `json"FuelInventory"`
	PathName      string  `json:"PathName"`
}

func NewVehicleCollector(frmAddress string) *VehicleCollector {
	return &VehicleCollector{
		FRMAddress: frmAddress,
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
		VehicleFuel.WithLabelValues(d.Id, d.FuelType).Set(d.FuelInventory)
	}
}
