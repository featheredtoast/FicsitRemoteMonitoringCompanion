package exporter_test

import (
	"github.com/AP-Hunt/FicsitRemoteMonitoringCompanion/m/v2/exporter"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("VehicleCollector", func() {
	var collector *exporter.VehicleCollector

	BeforeEach(func() {
		FRMServer.Reset()
		collector = exporter.NewVehicleCollector("http://localhost:9080/getVehicles")

		FRMServer.ReturnsVehicleData([]exporter.VehicleDetails{
			{
				Id: "1",
				VehicleType: "Truck",
				AutoPilot: true,
				FuelType: "Coal",
				FuelInventory: 23,
				PathName: "Path",
			},
		})
	})

	AfterEach(func() {
		collector = nil
	})

	Describe("Vehicle metrics collection", func() {
		It("sets the 'vehicle_fuel' metric with the right labels", func() {
			collector.Collect()

			val, err := gaugeValue(exporter.VehicleFuel, "1", "Truck", "Coal")

			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal(float64(23)))
		})
	})
})
