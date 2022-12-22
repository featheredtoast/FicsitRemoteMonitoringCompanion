package exporter_test

import (
	"github.com/AP-Hunt/FicsitRemoteMonitoringCompanion/m/v2/exporter"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"time"
)

func updateLocation (x float64, y float64, rotation int) {
	FRMServer.ReturnsVehicleData([]exporter.VehicleDetails{
		{
			Id:           "1",
			VehicleType:  "Truck",
			ForwardSpeed: 0,
			Location: exporter.Location{
				X:        x,
				Y:        y,
				Z:        0,
				Rotation: rotation,
			},
			AutoPilot:     true,
			FuelType:      "Coal",
			FuelInventory: 23,
			PathName:      "Path",
		},
	})
}

var _ = Describe("VehicleCollector", func() {
	var collector *exporter.VehicleCollector

	BeforeEach(func() {
		FRMServer.Reset()
		collector = exporter.NewVehicleCollector("http://localhost:9080/getVehicles")

		FRMServer.ReturnsVehicleData([]exporter.VehicleDetails{
			{
				Id:           "1",
				VehicleType:  "Truck",
				ForwardSpeed: 80,
				Location: exporter.Location{
					X:        1000,
					Y:        2000,
					Z:        1000,
					Rotation: 60,
				},
				AutoPilot:     true,
				FuelType:      "Coal",
				FuelInventory: 23,
				PathName:      "Path",
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

		It("sets the 'vehicle_round_trip_seconds' metric with the right labels", func() {

			now, _ := time.Parse(time.RFC3339, "2022-12-21T15:04:05Z")
			exporter.Now = func() time.Time {
				return now
			}
			// truck will be too fast here, nothing recorded
			collector.Collect()
			val, err := gaugeValue(exporter.VehicleRoundTrip, "1", "Truck", "Path")
			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal(float64(0)))

			exporter.Now = func() time.Time {
				d, _ := time.ParseDuration("30s")
				return now.Add(d)
			}

			updateLocation(0,0,0)
			// first time collecting stats, nothing yet but it does set start location to 0,0,0
			collector.Collect()
			val, err = gaugeValue(exporter.VehicleRoundTrip, "1", "Truck", "Path")
			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal(float64(0)))

			exporter.Now = func() time.Time {
				d, _ := time.ParseDuration("60s")
				return now.Add(d)
			}
			updateLocation(8000, 2000, 0)
			//go to a far away location now, star the timer
			collector.Collect()
			val, err = gaugeValue(exporter.VehicleRoundTrip, "1", "Truck", "Path")
			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal(float64(0)))

			exporter.Now = func() time.Time {
				d, _ := time.ParseDuration("70s")
				return now.Add(d)
			}
			updateLocation(1000, 2000, 180)
			//We are near but not facing the right way. Do not record this until we face near the right direction
			collector.Collect()
			val, err = gaugeValue(exporter.VehicleRoundTrip, "1", "Truck", "Path")
			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal(float64(0)))

			exporter.Now = func() time.Time {
				d, _ := time.ParseDuration("90s")
				return now.Add(d)
			}
			updateLocation(1000, 2000, 0)
			//Now we are back near enough to where we began recording, and facing near the same way end recording
			collector.Collect()
			val, err = gaugeValue(exporter.VehicleRoundTrip, "1", "Truck", "Path")
			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal(float64(30)))
		})
	})
})
