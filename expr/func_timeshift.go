package expr

import (
	"fmt"
	"time"

	"github.com/grafana/metrictank/api/models"
	"github.com/raintank/dur"
	"github.com/raintank/worldping-api/pkg/log"
	schema "gopkg.in/raintank/schema.v1"
)

type FuncTimeShift struct {
	in        GraphiteFunc
	timeShift string
	resetEnd  bool
	alignDST  bool

	shiftOffset int
}

func NewTimeShift() GraphiteFunc {
	return &FuncTimeShift{resetEnd: true, alignDST: false}
}

func (s *FuncTimeShift) Signature() ([]Arg, []Arg) {
	return []Arg{
		ArgSeriesList{val: &s.in},
		ArgString{val: &s.timeShift, validator: []Validator{IsIntervalString}},
		ArgBool{key: "resetEnd", opt: true, val: &s.resetEnd},
		ArgBool{key: "alignDST", opt: true, val: &s.alignDST},
	}, []Arg{ArgSeriesList{}}
}

func (s *FuncTimeShift) Context(context Context) Context {
	var err error
	s.shiftOffset, err = s.getTimeShiftSeconds(context)
	if err != nil {
		// panic??? TODO
		return context
	}

	context.from = addOffset(context.from, s.shiftOffset)
	context.to = addOffset(context.to, s.shiftOffset)
	return context
}

func (s *FuncTimeShift) Exec(cache map[Req][]models.Series) ([]models.Series, error) {
	series, err := s.in.Exec(cache)
	if err != nil {
		return nil, err
	}

	negativeOffset := -s.shiftOffset

	newName := func(oldName string) string {
		return fmt.Sprintf("timeShift(%s, \"%s\")", oldName, s.timeShift)
	}

	// TODO resetEnd - doesn't seem to make sense ??? I don't fully understand this option.

	outputs := make([]models.Series, 0, len(series))
	for _, serie := range series {
		out := pointSlicePool.Get().([]schema.Point)

		for _, v := range serie.Datapoints {
			log.Info("orig=%d, new=%d", v.Ts, addOffset(v.Ts, negativeOffset))
			out = append(out, schema.Point{Val: v.Val, Ts: addOffset(v.Ts, negativeOffset)})
		}

		output := models.Series{
			Target:     newName(serie.Target),
			QueryPatt:  newName(serie.QueryPatt),
			Tags:       serie.Tags,
			Datapoints: out,
			Interval:   serie.Interval,
		}
		output.Tags["timeShift"] = s.timeShift

		outputs = append(outputs, output)
		cache[Req{}] = append(cache[Req{}], output)
	}
	return outputs, nil
}

func (s *FuncTimeShift) getTimeShiftSeconds(context Context) (int, error) {
	// Trim off the sign (if there is one)
	sign := -1
	durStr := s.timeShift
	if durStr[0] == '-' {
		durStr = durStr[1:]
	} else if durStr[0] == '+' {
		sign = 1
		durStr = durStr[1:]
	}

	interval, err := dur.ParseDuration(s.timeShift)

	if err != nil {
		return 0, err
	}

	signedOffset := int(interval) * sign

	if s.alignDST {
		timezoneOffset := func(epochTime uint32) int {
			// User server timezone to determine DST
			localTime := time.Now()
			thenTime := time.Unix(int64(epochTime), 0)

			_, offset := time.Date(thenTime.Year(), thenTime.Month(),
				thenTime.Day(), thenTime.Hour(), thenTime.Minute(),
				thenTime.Second(), thenTime.Nanosecond(), localTime.Location()).Zone()

			return offset
		}

		reqStartOffset := timezoneOffset(context.from)
		reqEndOffset := timezoneOffset(context.to)
		myStartOffset := timezoneOffset(addOffset(context.from, signedOffset))
		myEndOffset := timezoneOffset(addOffset(context.to, signedOffset))

		dstOffset := 0

		// Both requests ranges are entirely in the one offset (either in or out of DST)
		if reqStartOffset == reqEndOffset && myStartOffset == myEndOffset {
			dstOffset = reqStartOffset - myStartOffset
		}

		signedOffset += dstOffset
	}

	return signedOffset, nil
}

func addOffset(orig uint32, offset int) uint32 {
	if offset < 0 {
		orig -= uint32(offset * -1)
	} else {
		orig += uint32(offset)
	}

	return orig
}
