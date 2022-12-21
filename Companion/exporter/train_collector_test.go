package exporter_test

import (
	"github.com/AP-Hunt/FicsitRemoteMonitoringCompanion/m/v2/exporter"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TrainCollector", func() {
	var collector *exporter.TrainCollector

	BeforeEach(func() {
		FRMServer.Reset()
		collector = exporter.NewTrainCollector("http://localhost:9080/getTrains")

		FRMServer.ReturnsTrainData([]exporter.TrainDetails{
			{
				TrainName:     "Train1",
				PowerConsumed: 67,
				TrainStation:  "NextStation",
				Derailed:      false,
				Status:        "TS_SelfDriving",
				TimeTable: []exporter.TimeTable{
					{StationName: "First"},
					{StationName: "Second"},
				},
			},
			{
				TrainName:     "DerailedTrain",
				PowerConsumed: 0,
				TrainStation:  "NextStation",
				Derailed:      true,
				Status:        "Derailed",
			},
		})
	})

	AfterEach(func() {
		collector = nil
	})

	Describe("Train metrics collection", func() {
		It("sets the 'train_derailed' metric with the right labels", func() {
			collector.Collect()

			val, err := gaugeValue(exporter.TrainDerailed, "DerailedTrain")

			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal(float64(1)))
		})
		It("sets the 'train_power_consumed' metric with the right labels", func() {
			collector.Collect()

			val, err := gaugeValue(exporter.TrainPower, "Train1")

			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal(float64(67)))
		})
	})
})
