package kafkamdm

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLagLogger(t *testing.T) {
	logger := newLagLogger(5)
	now := time.Now()

	Convey("with 0 measurements", t, func() {
		So(logger.Min(), ShouldEqual, -1)
	})
	Convey("with 1 measurements", t, func() {
		logger.Store(0, 10, now.Add(time.Second*time.Duration(1)))
		So(logger.Min(), ShouldEqual, 10)
		So(logger.Rate(), ShouldEqual, 0)
	})
	Convey("with 2 measurements", t, func() {
		logger.Store(10, 15, now.Add(time.Second*time.Duration(2)))
		So(logger.Min(), ShouldEqual, 5)
		So(logger.Rate(), ShouldEqual, 5)
	})
	Convey("with a negative measurement", t, func() {
		logger.Store(10, 5, now.Add(time.Second*time.Duration(3)))

		// Negative measuremnets are discarded, should be same as last time.
		So(logger.Min(), ShouldEqual, 5)
		So(logger.Rate(), ShouldEqual, 5)
	})
	Convey("with lots of measurements", t, func() {
		for i := 0; i < 100; i++ {
			logger.Store(int64(10+i), int64(15+2*i), now.Add(time.Second*time.Duration(3+i)))
		}
		So(logger.Min(), ShouldEqual, 100)
		So(logger.Rate(), ShouldEqual, 2)
	})
}

/*
func TestRateLogger(t *testing.T) {
	logger := newRateLogger()
	now := time.Now()
	Convey("with 0 measurements", t, func() {
		So(logger.Rate(), ShouldEqual, 0)
	})
	Convey("after 1st measurements", t, func() {
		logger.Store(10, now)
		So(logger.Rate(), ShouldEqual, 0)
	})
	Convey("with 2nd measurements", t, func() {
		logger.Store(15, now.Add(time.Second))
		So(logger.Rate(), ShouldEqual, 5)
	})
	Convey("with old ts", t, func() {
		logger.Store(25, now)
		So(logger.Rate(), ShouldEqual, 5)
	})
	Convey("with less then 1per second", t, func() {
		logger.Store(30, now.Add(time.Second*10))
		So(logger.Rate(), ShouldEqual, 0)
	})
}

func TestRateLoggerSmallIncrements(t *testing.T) {
	logger := newRateLogger()
	now := time.Now()
	Convey("after 1st measurements", t, func() {
		logger.Store(10, now)
		So(logger.Rate(), ShouldEqual, 0)
	})
	Convey("with 2nd measurements", t, func() {
		logger.Store(20, now.Add(200*time.Millisecond))
		So(logger.Rate(), ShouldEqual, 0)
	})
	Convey("with 3rd measurements", t, func() {
		logger.Store(30, now.Add(400*time.Millisecond))
		So(logger.Rate(), ShouldEqual, 0)
	})
	Convey("with 4th measurements", t, func() {
		logger.Store(40, now.Add(600*time.Millisecond))
		So(logger.Rate(), ShouldEqual, 0)
	})
	Convey("with 5th measurements", t, func() {
		logger.Store(50, now.Add(800*time.Millisecond))
		So(logger.Rate(), ShouldEqual, 0)
	})
	Convey("with 6th measurements", t, func() {
		logger.Store(60, now.Add(1000*time.Millisecond))
		So(logger.Rate(), ShouldEqual, 60-10)
	})
	Convey("with 7th measurements", t, func() {
		logger.Store(80, now.Add(1200*time.Millisecond))
		So(logger.Rate(), ShouldEqual, 60-10)
	})
	Convey("with 8th measurements", t, func() {
		logger.Store(100, now.Add(1400*time.Millisecond))
		So(logger.Rate(), ShouldEqual, 60-10)
	})
	Convey("with 9th measurements", t, func() {
		logger.Store(120, now.Add(1600*time.Millisecond))
		So(logger.Rate(), ShouldEqual, 60-10)
	})
	Convey("with 10th measurements", t, func() {
		logger.Store(140, now.Add(1800*time.Millisecond))
		So(logger.Rate(), ShouldEqual, 60-10)
	})
	Convey("with 11th measurements", t, func() {
		logger.Store(160, now.Add(2000*time.Millisecond))
		So(logger.Rate(), ShouldEqual, 160-60)
	})
}
*/

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
		So(mon.Metric(), ShouldEqual, 90)
	})
	Convey("metric should be worst partition", t, func() {
		now := time.Now()
		for part := range mon.monitors {
			mon.StoreOffsets(part, int64(part), int64(2*part+10), now.Add(time.Second*time.Duration(part)))
		}
		So(mon.Metric(), ShouldEqual, 13)
	})
}
