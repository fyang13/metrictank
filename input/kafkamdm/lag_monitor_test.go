package kafkamdm

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

type OffsetAdjuster struct {
	readOffset, highWaterMark int64
	ts                        time.Time
	lag                       *lagLogger
}

func (o *OffsetAdjuster) add(msgsProcessed, msgsAdded int64, secondsPassed int) {
	o.ts = o.ts.Add(time.Second * time.Duration(secondsPassed))
	o.readOffset += msgsProcessed
	o.highWaterMark += msgsAdded
	o.lag.Store(o.readOffset, o.highWaterMark, o.ts)
}

func TestLagLogger(t *testing.T) {
	logger := newLagLogger(5)

	Convey("with 0 measurements", t, func() {
		So(logger.Min(), ShouldEqual, -1)
	})

	adjuster := OffsetAdjuster{
		readOffset:    10 * 1000 * 1000,
		highWaterMark: 10 * 1000 * 1000,
		ts:            time.Now(),
		lag:           logger,
	}
	Convey("with 1 measurements", t, func() {
		adjuster.add(0, 10, 1)
		So(logger.Min(), ShouldEqual, 10)
		So(logger.Rate(), ShouldEqual, 0)
	})
	Convey("with 2 measurements", t, func() {
		adjuster.add(10, 5, 1)
		So(logger.Min(), ShouldEqual, 5)
		So(logger.Rate(), ShouldEqual, 5)
	})
	Convey("with a negative measurement", t, func() {
		// Negative measurements are discarded, should be same as last time.
		// Add directly to not mess with the adjusters offsets
		logger.Store(adjuster.readOffset, adjuster.highWaterMark-100, adjuster.ts)
		So(logger.Min(), ShouldEqual, 5)
		So(logger.Rate(), ShouldEqual, 5)
	})
	Convey("with lots of measurements", t, func() {
		// fall behind by 1 each step (we started behind by 5)
		for i := 0; i < 100; i++ {
			adjuster.add(19, 20, 1)
		}
		So(logger.Min(), ShouldEqual, 101)
		So(logger.Rate(), ShouldEqual, 20)
	})
}

func TestLagWithShortProcessingPause(t *testing.T) {
	logger := newLagLogger(5)

	adjuster := OffsetAdjuster{
		readOffset:    10 * 1000 * 1000,
		highWaterMark: 10 * 1000 * 1000,
		ts:            time.Now(),
		lag:           logger,
	}

	// start with 50 lag
	adjuster.highWaterMark += 50

	// Simulate being almost in sync
	for i := 0; i < 100; i++ {
		adjuster.add(5000, 5000, 5)
	}

	Convey("should be almost in sync", t, func() {
		So(logger.Min(), ShouldEqual, 50)
		So(logger.Rate(), ShouldEqual, 1000)
	})

	// Short pause with no msgs processed
	adjuster.add(0, 5*1000, 5)

	Convey("Short pause should not cause large lag estimate", t, func() {
		So(logger.Min(), ShouldEqual, 50)
		So(logger.Rate(), ShouldEqual, 1000)
	})
}

func TestLagMonitor(t *testing.T) {
	mon := NewLagMonitor(10, []int32{0, 1, 2, 3})
	Convey("with 0 measurements", t, func() {
		So(mon.Metric(), ShouldEqual, 10000)
	})
	Convey("with lots of measurements", t, func() {
		now := time.Now()
		for part := range mon.monitors {
			for i := 0; i < 100; i++ {
				mon.StoreOffsets(part, int64(i), int64(2*i), now.Add(time.Second*time.Duration(i)))
			}
		}
		So(mon.Metric(), ShouldEqual, 45)
	})
	Convey("metric should be worst partition", t, func() {
		now := time.Now()
		for part := range mon.monitors {
			mon.StoreOffsets(part, int64(part+200), int64(2*part+210), now.Add(time.Second*time.Duration(part+100)))
		}
		So(mon.Metric(), ShouldEqual, 6)
	})
}
