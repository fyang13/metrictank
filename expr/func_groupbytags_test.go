package expr

import (
	"errors"
	"math"
	"sort"
	"strconv"
	"testing"

	"github.com/grafana/metrictank/api/models"
	"github.com/grafana/metrictank/test"
	"github.com/raintank/schema"
)

func getModel(name string, data []schema.Point) models.Series {
	series := models.Series{
		Target:     name,
		QueryPatt:  name,
		Datapoints: getCopy(data),
	}
	series.SetTags()
	return series
}

// Test error cases
func TestNoTags(t *testing.T) {
	in := []models.Series{
		getModel("name1;tag1=val1", a),
	}
	expected := errors.New("No tags specified")

	testGroupByTags("ErrNoTags", in, nil, "sum", []string{}, expected, t)
}

// Test normal cases
func TestGroupByTagsSingleSeries(t *testing.T) {
	in := []models.Series{
		getModel("name1;tag1=val1", a),
	}
	out := []models.Series{
		getModel("name1;tag1=val1", a),
	}

	aggs := []string{"average", "median", "sum", "min", "max", "stddev", "diff", "range", "multiply"}

	for _, agg := range aggs {
		out[0].Datapoints = out[0].Datapoints[:0]
		aggFunc := getCrossSeriesAggFunc(agg)
		aggFunc(in, &out[0].Datapoints)

		testGroupByTags("SingleSeries"+agg, in, out, agg, []string{"tag1"}, nil, t)
	}
}

func TestGroupByTagsMultipleSeriesSingleResult(t *testing.T) {
	in := []models.Series{
		getModel("name1;tag1=val1;tag2=val2_0", a),
		getModel("name1;tag1=val1;tag2=val2_1", b),
	}
	out := []models.Series{
		getModel("name1;tag1=val1", sumab),
	}

	testGroupByTags("MultipleSeriesSingleResult", in, out, "sum", []string{"tag1"}, nil, t)
}

func TestGroupByTagsMultipleSeriesMultipleResults(t *testing.T) {
	in := []models.Series{
		getModel("name1;tag1=val1;tag2=val2_0", a),
		getModel("name1;tag1=val1;tag2=val2_1", b),
		getModel("name1;tag1=val1_1;tag2=val2_0", c),
		getModel("name1;tag1=val1_1;tag2=val2_1", d),
	}
	out := []models.Series{
		getModel("name1;tag1=val1", sumab),
		getModel("name1;tag1=val1_1", sumcd),
	}

	testGroupByTags("MultipleSeriesMultipleResult", in, out, "sum", []string{"tag1"}, nil, t)
}
func TestGroupByTagsMultipleSeriesMultipleResultsMultipleNames(t *testing.T) {
	in := []models.Series{
		getModel("name1;tag1=val1;tag2=val2_0", a),
		getModel("name1;tag1=val1;tag2=val2_1", b),
		getModel("name2;tag1=val1_1;tag2=val2_0", c),
		getModel("name2;tag1=val1_1;tag2=val2_1", d),
	}
	out := []models.Series{
		getModel("sum;tag1=val1", sumab),
		getModel("sum;tag1=val1_1", sumcd),
	}

	testGroupByTags("MultipleSeriesMultipleResultsMultipleNames", in, out, "sum", []string{"tag1"}, nil, t)
}

func TestGroupByTagsMultipleSeriesMultipleResultsMultipleNamesMoreTags(t *testing.T) {
	in := []models.Series{
		getModel("name1;tag1=val1;tag2=val2_0;tag3=3", a),
		getModel("name1;tag1=val1;tag2=val2_1;tag3=3", b),
		getModel("name2;tag1=val1_1;tag2=val2_0;tag3=3", c),
		getModel("name2;tag1=val1_1;tag2=val2_1;tag3=3", d),
	}
	out := []models.Series{
		getModel("sum;tag1=val1;tag3=3", sumab),
		getModel("sum;tag1=val1_1;tag3=3", sumcd),
	}

	testGroupByTags("MultipleSeriesMultipleResultsMultipleNamesMoreTags", in, out, "sum", []string{"tag1", "tag3"}, nil, t)
}

func TestGroupByTagsMultipleSeriesMultipleResultsGroupByName(t *testing.T) {
	in := []models.Series{
		getModel("name1;tag1=val1;tag2=val2_0", a),
		getModel("name1;tag1=val1;tag2=val2_1", b),
		getModel("name2;tag1=val1_1;tag2=val2_0", c),
		getModel("name2;tag1=val1_1;tag2=val2_1", d),
	}
	out := []models.Series{
		getModel("name1;tag1=val1", sumab),
		getModel("name2;tag1=val1_1", sumcd),
	}

	testGroupByTags("MultipleSeriesMultipleResultsGroupByName", in, out, "sum", []string{"tag1", "name"}, nil, t)
}

func TestGroupByTagsSingleGroupByName(t *testing.T) {
	in := []models.Series{
		getModel("name1;tag1=val1;tag2=val2_0;tag3=3", a),
		getModel("name1;tag1=val1;tag2=val2_1;tag3=3", b),
		getModel("name1;tag1=val1_1;tag2=val2_0;tag3=3", c),
	}
	out := []models.Series{
		getModel("name1", sumabc),
	}

	testGroupByTags("SingleGroupByName", in, out, "sum", []string{"name"}, nil, t)
}

func TestGroupByTagsMultipleGroupByName(t *testing.T) {
	in := []models.Series{
		getModel("name1;tag1=val1;tag2=val2_0;tag3=3", a),
		getModel("name1;tag1=val1;tag2=val2_1;tag3=3", b),
		getModel("name2;tag1=val1_1;tag2=val2_0;tag3=3", c),
		getModel("name2;tag1=val1_1;tag2=val2_1;tag3=3", d),
	}
	out := []models.Series{
		getModel("name1", sumab),
		getModel("name2", sumcd),
	}

	testGroupByTags("MultipleGroupByName", in, out, "sum", []string{"name"}, nil, t)
}

func TestGroupByTagsMultipleSeriesMissingTag(t *testing.T) {
	in := []models.Series{
		getModel("name1;tag1=val1;tag2=val2_0", a),
		getModel("name1;tag1=val1;tag2=val2_1", b),
		getModel("name2;tag1=val1_1;tag2=val2_0", c),
		getModel("name2;tag1=val1_1;tag2=val2_1", d),
	}
	out := []models.Series{
		getModel("name1;missingTag=;tag1=val1", sumab),
		getModel("name2;missingTag=;tag1=val1_1", sumcd),
	}

	testGroupByTags("MultipleSeriesMissingTag", in, out, "sum", []string{"tag1", "name", "missingTag"}, nil, t)
}

func TestGroupByTagsAllAggregators(t *testing.T) {
	aggregators := []struct {
		name             string
		result1, result2 []schema.Point
	}{
		{name: "sum", result1: sumab, result2: sumabc},
		{name: "avg", result1: avgab, result2: avgabc},
		{name: "average", result1: avgab, result2: avgabc},
		{name: "max", result1: maxab, result2: maxabc},
		{name: "median", result1: medianab, result2: medianabc},
		{name: "multiply", result1: multab, result2: multabc},
		{name: "stddev", result1: stddevab, result2: stddevabc},
		{name: "diff", result1: diffab, result2: diffabc},
		{name: "range", result1: rangeab, result2: rangeabc},
	}

	for _, agg := range aggregators {
		in := []models.Series{
			getModel("name1;tag1=val1;tag2=val2_0", a),
			getModel("name1;tag1=val1;tag2=val2_1", b),
			getModel("name2;tag1=val1_1;tag2=val2_0", a),
			getModel("name2;tag1=val1_1;tag2=val2_1", b),
			getModel("name2;tag1=val1_1;tag2=val2_2", c),
		}
		out := []models.Series{
			getModel("name1;tag1=val1", agg.result1),
			getModel("name2;tag1=val1_1", agg.result2),
		}

		testGroupByTags("AllAggregators:"+agg.name, in, out, agg.name, []string{"tag1", "name"}, nil, t)
	}
}

func testGroupByTags(name string, in []models.Series, out []models.Series, agg string, tags []string, expectedErr error, t *testing.T) {
	f := NewGroupByTags()
	gby := f.(*FuncGroupByTags)
	gby.in = NewMock(in)
	gby.aggregator = agg
	gby.tags = tags

	got, err := f.Exec(make(map[Req][]models.Series))
	if err != expectedErr {
		if expectedErr == nil {
			t.Fatalf("case %q: expected no error but got %q", name, err)
		} else if err == nil || err.Error() != expectedErr.Error() {
			t.Fatalf("case %q: expected error %q but got %q", name, expectedErr, err)
		}
	}
	if len(got) != len(out) {
		t.Fatalf("case %q: GroupByTags output expected to be %d but actually %d", name, len(out), len(got))
	}

	// Make sure got and out are in the same order
	sort.Slice(got, func(i, j int) bool {
		return got[i].Target < got[j].Target
	})
	sort.Slice(out, func(i, j int) bool {
		return out[i].Target < out[j].Target
	})
	for i, g := range got {
		o := out[i]
		if g.Target != o.Target {
			t.Fatalf("case %q: expected target %q, got %q", name, o.Target, g.Target)
		}
		if len(g.Datapoints) != len(o.Datapoints) {
			t.Fatalf("case %q: len output expected %d, got %d", name, len(o.Datapoints), len(g.Datapoints))
		}
		for j, p := range g.Datapoints {
			bothNaN := math.IsNaN(p.Val) && math.IsNaN(o.Datapoints[j].Val)
			if (bothNaN || p.Val == o.Datapoints[j].Val) && p.Ts == o.Datapoints[j].Ts {
				continue
			}
			t.Fatalf("case %q: output point %d - expected %v got %v", name, j, o.Datapoints[j], p)
		}
		if len(g.Tags) != len(o.Tags) {
			t.Fatalf("case %q: len tags expected %d, got %d", name, len(o.Tags), len(g.Tags))
		}
		for k, v := range g.Tags {
			expectedVal, ok := o.Tags[k]

			if !ok {
				t.Fatalf("case %q: Got unknown tag key '%s'", name, k)
			}

			if v != expectedVal {
				t.Fatalf("case %q: Key '%s' had wrong value: expected '%s', got '%s'", name, k, expectedVal, v)
			}
		}
	}
}

// Benchmarks:

// input series: 1, 10, 100, 1k, 10k, 100k
// output series: 1, same as input, then if applicable: 10, 100, 1k, 10k

// 1 input series
func BenchmarkGroupByTags1in1out(b *testing.B) {
	benchmarkGroupByTags(b, 1, 1)
}

// 10 input Series
func BenchmarkGroupByTags10in1out(b *testing.B) {
	benchmarkGroupByTags(b, 10, 1)
}

func BenchmarkGroupByTags10in10out(b *testing.B) {
	benchmarkGroupByTags(b, 10, 10)
}

// 100 input series
func BenchmarkGroupByTags100in1out(b *testing.B) {
	benchmarkGroupByTags(b, 100, 1)
}

func BenchmarkGroupByTags100in10out(b *testing.B) {
	benchmarkGroupByTags(b, 100, 10)
}

func BenchmarkGroupByTags100in100out(b *testing.B) {
	benchmarkGroupByTags(b, 100, 100)
}

// 1k input series
func BenchmarkGroupByTags1000in1out(b *testing.B) {
	benchmarkGroupByTags(b, 1000, 1)
}

func BenchmarkGroupByTags1000in10out(b *testing.B) {
	benchmarkGroupByTags(b, 1000, 10)
}

func BenchmarkGroupByTags1000in100out(b *testing.B) {
	benchmarkGroupByTags(b, 1000, 100)
}

func BenchmarkGroupByTags1000in1000out(b *testing.B) {
	benchmarkGroupByTags(b, 1000, 1000)
}

// 10k input series
func BenchmarkGroupByTags10000in1out(b *testing.B) {
	benchmarkGroupByTags(b, 10000, 1)
}

func BenchmarkGroupByTags10000in10out(b *testing.B) {
	benchmarkGroupByTags(b, 10000, 10)
}

func BenchmarkGroupByTags10000in100out(b *testing.B) {
	benchmarkGroupByTags(b, 10000, 100)
}

func BenchmarkGroupByTags10000in1000out(b *testing.B) {
	benchmarkGroupByTags(b, 10000, 1000)
}

func BenchmarkGroupByTags10000in10000out(b *testing.B) {
	benchmarkGroupByTags(b, 10000, 10000)
}

// 100k input series
func BenchmarkGroupByTags100000in1out(b *testing.B) {
	benchmarkGroupByTags(b, 100000, 1)
}

func BenchmarkGroupByTags100000in10out(b *testing.B) {
	benchmarkGroupByTags(b, 100000, 10)
}

func BenchmarkGroupByTags100000in100out(b *testing.B) {
	benchmarkGroupByTags(b, 100000, 100)
}

func BenchmarkGroupByTags100000in1000out(b *testing.B) {
	benchmarkGroupByTags(b, 100000, 1000)
}

func BenchmarkGroupByTags100000in10000out(b *testing.B) {
	benchmarkGroupByTags(b, 100000, 10000)
}

func BenchmarkGroupByTags100000in100000out(b *testing.B) {
	benchmarkGroupByTags(b, 100000, 100000)
}

func benchmarkGroupByTags(b *testing.B, numInputSeries, numOutputSeries int) {
	var input []models.Series
	tagValues := []string{"tag1", "tag2", "tag3", "tag4"}
	for i := 0; i < numInputSeries; i++ {
		series := models.Series{
			Target: strconv.Itoa(i),
		}

		for _, tag := range tagValues {
			series.Target += ";" + tag + "=" + strconv.Itoa(i%numOutputSeries)
		}

		series.Datapoints = test.RandFloats100()
		input = append(input, series)
	}
	b.ResetTimer()
	var err error
	for i := 0; i < b.N; i++ {
		f := NewGroupByTags()
		gby := f.(*FuncGroupByTags)
		gby.in = NewMock(input)
		gby.aggregator = "sum"
		gby.tags = []string{"tag1", "tag2"}
		results, err = f.Exec(make(map[Req][]models.Series))
		if err != nil {
			b.Fatalf("%s", err)
		}

		if len(results) != numOutputSeries {
			b.Fatalf("Expected %d groups, got %d", numOutputSeries, len(results))
		}

		if true {
			for _, serie := range results {
				pointSlicePool.Put(serie.Datapoints[:0])
			}
		}
	}
	b.SetBytes(int64(numInputSeries * len(results[0].Datapoints) * 12))
}
