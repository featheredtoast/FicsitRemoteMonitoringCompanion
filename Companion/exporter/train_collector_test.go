package exporter_test

import (
	"github.com/AP-Hunt/FicsitRemoteMonitoringCompanion/m/v2/exporter"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"time"
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
		It("sets the 'train_round_trip_seconds' and 'train_segment_trip_seconds' metric with the right labels", func() {

			now := time.Now()
			exporter.Now = func() time.Time {
				return now
			}

			FRMServer.ReturnsTrainData([]exporter.TrainDetails{
				{
					TrainName:     "Train1",
					PowerConsumed: 0,
					TrainStation:  "First",
					Derailed:      false,
					Status:        "TS_SelfDriving",
					TimeTable: []exporter.TimeTable{
						{StationName: "First"},
						{StationName: "Second"},
						{StationName: "Third"},
					},
				},
			})

			collector.Collect()
			val, err := gaugeValue(exporter.TrainRoundTrip, "Train1")
			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal(float64(0)))
			val, err = gaugeValue(exporter.TrainSegmentTrip, "Train1", "First", "Second")
			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal(float64(0)))

			exporter.Now = func() time.Time {
				d, _ := time.ParseDuration("30s")
				return now.Add(d)
			}

			FRMServer.ReturnsTrainData([]exporter.TrainDetails{
				{
					TrainName:     "Train1",
					PowerConsumed: 0,
					TrainStation:  "First",
					Derailed:      false,
					Status:        "TS_SelfDriving",
					TimeTable: []exporter.TimeTable{
						{StationName: "First"},
						{StationName: "Second"},
						{StationName: "Third"},
					},
				},
			})

			collector.Collect()
			val, err = gaugeValue(exporter.TrainRoundTrip, "Train1")
			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal(float64(0)))
			val, err = gaugeValue(exporter.TrainSegmentTrip, "Train1", "First", "Second")
			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal(float64(0)))

			exporter.Now = func() time.Time {
				d, _ := time.ParseDuration("1m")
				return now.Add(d)
			}

			FRMServer.ReturnsTrainData([]exporter.TrainDetails{
				{
					TrainName:     "Train1",
					PowerConsumed: 0,
					TrainStation:  "Second",
					Derailed:      false,
					Status:        "TS_SelfDriving",
					TimeTable: []exporter.TimeTable{
						{StationName: "First"},
						{StationName: "Second"},
						{StationName: "Third"},
					},
				},
			})

			collector.Collect()
			val, err = gaugeValue(exporter.TrainRoundTrip, "Train1")
			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal(float64(0)))
			val, err = gaugeValue(exporter.TrainSegmentTrip, "Train1", "First", "Second")
			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal(float64(60)))

			exporter.Now = func() time.Time {
				d, _ := time.ParseDuration("2m")
				return now.Add(d)
			}

			FRMServer.ReturnsTrainData([]exporter.TrainDetails{
				{
					TrainName:     "Train1",
					PowerConsumed: 0,
					TrainStation:  "Third",
					Derailed:      false,
					Status:        "TS_SelfDriving",
					TimeTable: []exporter.TimeTable{
						{StationName: "First"},
						{StationName: "Second"},
						{StationName: "Third"},
					},
				},
			})

			collector.Collect()
			val, err = gaugeValue(exporter.TrainRoundTrip, "Train1")
			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal(float64(0)))
			val, err = gaugeValue(exporter.TrainSegmentTrip, "Train1", "Second", "Third")
			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal(float64(60)))

			exporter.Now = func() time.Time {
				d, _ := time.ParseDuration("3m")
				return now.Add(d)
			}
			FRMServer.ReturnsTrainData([]exporter.TrainDetails{
				{
					TrainName:     "Train1",
					PowerConsumed: 0,
					TrainStation:  "First",
					Derailed:      false,
					Status:        "TS_SelfDriving",
					TimeTable: []exporter.TimeTable{
						{StationName: "First"},
						{StationName: "Second"},
						{StationName: "Third"},
					},
				},
			})

			collector.Collect()
			val, err = gaugeValue(exporter.TrainRoundTrip, "Train1")
			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal(float64(180)))
			val, err = gaugeValue(exporter.TrainSegmentTrip, "Train1", "Third", "First")
			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal(float64(60)))

		})
	})
})
