package exporter_test

import (
	"github.com/AP-Hunt/FicsitRemoteMonitoringCompanion/m/v2/exporter"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"time"
)

func updateTrain(station string) {
	FRMServer.ReturnsTrainData([]exporter.TrainDetails{
		{
			TrainName:     "Train1",
			PowerConsumed: 0,
			TrainStation:  station,
			Derailed:      false,
			Status:        "TS_SelfDriving",
			TimeTable: []exporter.TimeTable{
				{StationName: "First"},
				{StationName: "Second"},
				{StationName: "Third"},
			},
		},
	})
}

func advanceTime(now time.Time, increment time.Duration) time.Time {
	exporter.Now = func() time.Time {
		return now.Add(increment)
	}
	return now.Add(increment)
}

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

			now, _ := time.Parse(time.RFC3339, "2022-12-21T15:04:05Z")
			increment, _ := time.ParseDuration("5s")
			exporter.Now = func() time.Time {
				return now
			}
			updateTrain("First")

			collector.Collect()
			val, err := gaugeValue(exporter.TrainRoundTrip, "Train1")
			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal(float64(0)))
			val, err = gaugeValue(exporter.TrainSegmentTrip, "Train1", "First", "Second")
			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal(float64(0)))
			now = advanceTime(now, increment)
			collector.Collect()
			val, err = gaugeValue(exporter.TrainSegmentTrip, "Train1", "First", "Second")
			Expect(val).To(Equal(float64(0)))
			now = advanceTime(now, increment)
			collector.Collect()
			now = advanceTime(now, increment)
			collector.Collect()
			now = advanceTime(now, increment)
			collector.Collect()
			now = advanceTime(now, increment)
			collector.Collect()
			now = advanceTime(now, increment)

			// Start timing the trains here - No metrics yet because we just got our first "start" marker from the station change.
			updateTrain("Second")
			collector.Collect()
			val, err = gaugeValue(exporter.TrainRoundTrip, "Train1")
			Expect(val).To(Equal(float64(0)))
			val, err = gaugeValue(exporter.TrainSegmentTrip, "Train1", "First", "Second")
			Expect(val).To(Equal(float64(0)))

			now = advanceTime(now, increment)
			collector.Collect()
			now = advanceTime(now, increment)
			collector.Collect()
			now = advanceTime(now, increment)
			collector.Collect()
			now = advanceTime(now, increment)
			collector.Collect()
			now = advanceTime(now, increment)
			collector.Collect()
			// No stats again since train is still "en route"
			val, err = gaugeValue(exporter.TrainSegmentTrip, "Train1", "First", "Second")
			Expect(val).To(Equal(float64(0)))

			now = advanceTime(now, increment)

			// Train reaches third station - can record elapsed time between Second and Third station only
			updateTrain("Third")
			collector.Collect()
			val, err = gaugeValue(exporter.TrainRoundTrip, "Train1")
			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal(float64(0)))
			val, err = gaugeValue(exporter.TrainSegmentTrip, "Train1", "Second", "Third")
			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal(float64(30)))

			now = advanceTime(now, increment)
			collector.Collect()
			now = advanceTime(now, increment)
			collector.Collect()
			now = advanceTime(now, increment)
			collector.Collect()
			now = advanceTime(now, increment)
			collector.Collect()
			now = advanceTime(now, increment)
			collector.Collect()
			now = advanceTime(now, increment)

			updateTrain("First")
			collector.Collect()
			val, err = gaugeValue(exporter.TrainRoundTrip, "Train1")
			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal(float64(0)))
			val, err = gaugeValue(exporter.TrainSegmentTrip, "Train1", "Third", "First")
			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal(float64(30)))

			now = advanceTime(now, increment)
			collector.Collect()
			now = advanceTime(now, increment)
			collector.Collect()
			now = advanceTime(now, increment)
			collector.Collect()
			now = advanceTime(now, increment)
			collector.Collect()
			now = advanceTime(now, increment)
			collector.Collect()
			now = advanceTime(now, increment)
			updateTrain("Second")
			collector.Collect()

			val, err = gaugeValue(exporter.TrainRoundTrip, "Train1")
			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal(float64(90)))
			val, err = gaugeValue(exporter.TrainSegmentTrip, "Train1", "First", "Second")
			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal(float64(30)))

		})
	})
})
