package vimebu

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

type stringerValue struct {
	value string
}

func (v stringerValue) String() string {
	return v.value
}

type label struct {
	name        string
	value       any
	shouldQuote bool
}
type input struct {
	name   string
	labels []label
}
type testCase struct {
	name      string
	expected  string
	input     input
	skipBench bool
}

var testCases = []testCase{
	{
		name: "metric with labels",
		input: input{
			labels: []label{{"cluster", "guava", false}, {"host", "1.2.3.4", false}},
			name:   "cassandra_query_count",
		},
		expected: `cassandra_query_count{cluster="guava",host="1.2.3.4"}`,
	},
	{
		name: "metric with single label",
		input: input{
			labels: []label{{"type", "std", false}},
			name:   "produce_one_total",
		},
		expected: `produce_one_total{type="std"}`,
	},
	{
		name: "metric without label",
		input: input{
			name: "http_request_duration_seconds",
		},
		expected: `http_request_duration_seconds`,
	},

	{
		name: "metric with a lot of labels",
		input: input{
			labels: []label{
				{"method", "PUT", false},
				{"host", "1.2.3.4", false},
				{"status", "OK", false},
				{"node", "node--11-3", false},
				{"path", "/foo/bar", false},
				{"size", "667", false},
				{"auth", "basic", false},
				{"error", "nil", false},
				{"cached", "false", false},
				{"query", "select_boop", false},
			},
			name: "api_http_requests_total",
		},
		expected: `api_http_requests_total{method="PUT",host="1.2.3.4",status="OK",node="node--11-3",path="/foo/bar",size="667",auth="basic",error="nil",cached="false",query="select_boop"}`,
	},
	{
		name: "some empty labels and values",
		input: input{
			labels: []label{{"operation", "", false}, {"", "1.2.3.4", false}, {"status", "OK", false}, {"node", "", false}},
			name:   "api_http_requests_total",
		},
		expected:  `api_http_requests_total{status="OK"}`,
		skipBench: true,
	},
	{
		name: "values contain double quotes",
		input: input{
			labels: []label{{"error", `something went "horribly" wrong`, true}, {"path", `some/path/"with"/quo"tes`, true}},
			name:   "api_http_requests_total",
		},
		expected: `api_http_requests_total{error="something went \"horribly\" wrong",path="some/path/\"with\"/quo\"tes"}`,
	},
	{
		name: "values with and without double quotes",
		input: input{
			labels: []label{
				{"status", "Internal Server Error", false},
				{"error", `something went "horribly" wrong`, true},
				{"host", "1.2.3.4", false},
				{"path", `some/path/"with"/quo"tes`, true},
			},
			name: "api_http_requests_total",
		},
		expected: `api_http_requests_total{status="Internal Server Error",error="something went \"horribly\" wrong",host="1.2.3.4",path="some/path/\"with\"/quo\"tes"}`,
	},
	{
		name: "mixed label value types",
		input: input{
			name: "cassandra_query_count",
			labels: []label{
				{"path", `/"some"/path`, true},
				{"is_bidule", true, false},
				{"is_tac", false, false},
				{"point", float64(123.456), false},
				{"num", int64(1234), false},
				{"stringer", stringerValue{"spiderman"}, false},
				{"uint8", uint8(128), false},
				{"int", int(-42), false},
				{"error", fmt.Errorf("i/o timeout"), false},
			},
		},
		expected: `cassandra_query_count{path="/\"some\"/path",is_bidule="true",is_tac="false",point="123.456",num="1234",stringer="spiderman",uint8="128",int="-42",error="i/o timeout"}`,
	},
	{
		name: "bool label values",
		input: input{
			name: "cassandra_query_count",
			labels: []label{
				{"is_bidule", true, false},
				{"is_tac", false, false},
			},
		},
		expected: `cassandra_query_count{is_bidule="true",is_tac="false"}`,
	},
	{
		name: "int label values",
		input: input{
			name: "cassandra_query_count",
			labels: []label{
				{"a", int8(12), false},
				{"b", int16(5555), false},
				{"c", int32(69002), false},
				{"d", int64(80085), false},
				{"e", int(-1234), false},
			},
		},
		expected: `cassandra_query_count{a="12",b="5555",c="69002",d="80085",e="-1234"}`,
	},
	{
		name: "uint label values",
		input: input{
			name: "cassandra_query_count",
			labels: []label{
				{"a", uint8(12), false},
				{"b", uint16(5555), false},
				{"c", uint32(69002), false},
				{"d", uint64(80085), false},
				{"e", uint(1234), false},
			},
		},
		expected: `cassandra_query_count{a="12",b="5555",c="69002",d="80085",e="1234"}`,
	},
	{
		name: "float label values",
		input: input{
			name: "cassandra_query_count",
			labels: []label{
				{"a", float32(1), false},
				{"b", float32(0), false},
				{"c", float64(11111111.22222222), false},
				{"d", float64(1234.456789), false},
				{"e", float64(1234.4567890000), false},
			},
		},
		expected: `cassandra_query_count{a="1",b="0",c="11111111.22222222",d="1234.456789",e="1234.456789"}`,
	},
	{
		name: "fmt.Stringer label values",
		input: input{
			name: "external_hit_count",
			labels: []label{
				{"key", stringerValue{"value"}, false},
				{"key_quoted", stringerValue{`"yep"`}, true},
			},
		},
		expected: `external_hit_count{key="value",key_quoted="\"yep\""}`,
	},
}

func addLabelAnyToBuilder(builder *Builder, label label) {
	switch v := label.value.(type) {
	case string:
		if label.shouldQuote {
			builder.LabelString(label.name, v)
		} else {
			builder.LabelTrustedString(label.name, v)
		}
	case bool:
		builder.LabelBool(label.name, v)
	case uint8:
		builder.LabelUint8(label.name, v)
	case uint16:
		builder.LabelUint16(label.name, v)
	case uint32:
		builder.LabelUint32(label.name, v)
	case uint64:
		builder.LabelUint64(label.name, v)
	case uint:
		builder.LabelUint(label.name, v)
	case int8:
		builder.LabelInt8(label.name, v)
	case int16:
		builder.LabelInt16(label.name, v)
	case int32:
		builder.LabelInt32(label.name, v)
	case int64:
		builder.LabelInt64(label.name, v)
	case int:
		builder.LabelInt(label.name, v)
	case float32:
		builder.LabelFloat32(label.name, v)
	case float64:
		builder.LabelFloat64(label.name, v)
	case fmt.Stringer:
		if label.shouldQuote {
			builder.LabelStringer(label.name, v)
		} else {
			builder.LabelTrustedStringer(label.name, v)
		}
	case error:
		if label.shouldQuote {
			builder.LabelNamedError(label.name, v)
		} else {
			builder.LabelNamedTrustedError(label.name, v)
		}
	default:
		panic(fmt.Sprintf("unsupported type %T", v))
	}
}

func TestBuilderMetricEmptyName(t *testing.T) {
	require.Panics(t, func() {
		Metric("")
	})

	require.Panics(t, func() {
		var builder Builder
		builder.Metric("")
	})
}

func TestBuilderMetricAlreadyCalled(t *testing.T) {
	var builder Builder
	builder.Metric("test_metric")

	require.Panics(t, func() {
		builder.Metric("another_metric")
	})
}

func TestBuilderMetricNotCalled(t *testing.T) {
	var builder Builder

	require.Panics(t, func() {
		builder.LabelString("host", "1.2.3.4")
	})
}

func TestBuilder(t *testing.T) {
	t.Parallel()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.NotPanics(t, func() {
				builder := Metric(tc.input.name)
				for _, label := range tc.input.labels {
					addLabelAnyToBuilder(builder, label)
				}
				result := builder.String()
				require.Equal(t, tc.expected, result)
			})
		})
	}
}

func TestBuilderParallel(t *testing.T) {
	var eg errgroup.Group
	for i := range 400 {
		i := i
		name := fmt.Sprintf("foobar%d", i)
		eg.Go(func() error {
			require.NotPanics(t, func() {
				Metric(name).
					LabelTrustedString("host", "foobar").
					LabelBool("compressed", false).
					LabelUint8("port", 80).
					LabelFloat32("float", 12.3).
					LabelNamedError("err", nil).
					GetOrCreateCounter().
					Add(300)
			})
			return nil
		})
	}
	require.NoError(t, eg.Wait())
}

func captureLogOutput(f func()) []string {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	f()
	log.SetOutput(os.Stderr)
	return strings.FieldsFunc(buf.String(), func(c rune) bool {
		return c == '\n' || c == '\r'
	})
}

func TestBuilderOptionsWithLabelNameMaxLen(t *testing.T) {
	logLines := captureLogOutput(func() {
		metric := Metric("test_options", WithLabelNameMaxLen(3)).
			LabelTrustedString("one", "one").
			LabelTrustedString("two", "two").
			LabelTrustedString("three", "three").                  // Will be skipped.
			LabelString("four", `fo"ur`).                          // Will be skipped.
			LabelTrustedError(fmt.Errorf("something went wrong")). // Will be skipped.
			LabelError(fmt.Errorf(`"something" went wrong`)).      // Will be skipped.
			LabelNamedTrustedError("err", fmt.Errorf("mayday")).
			String()
		require.Equal(t, `test_options{one="one",two="two",err="mayday"}`, metric)
	})

	require.Len(t, logLines, 4) // One log line for each label name exceeding the limit of 3 bytes, thus getting skipped.
}

func TestBuilderOptionsWithLabelValueMaxLen(t *testing.T) {
	logLines := captureLogOutput(func() {
		metric := Metric("test_options", WithLabelValueMaxLen(5)).
			LabelTrustedString("one", "one").
			LabelTrustedString("two", "two").
			LabelTrustedString("three", "three").
			LabelString("four", `fo"ur`).
			LabelTrustedError(fmt.Errorf("something went wrong")). // Will be skipped.
			LabelError(fmt.Errorf(`"something" went wrong`)).      // Will be skipped.
			LabelNamedTrustedError("err", fmt.Errorf("mayday")).   // Will be skipped.
			String()
		require.Equal(t, `test_options{one="one",two="two",three="three",four="fo\"ur"}`, metric)
	})

	require.Len(t, logLines, 3) // One log line for each label value exceeding the limit of 5 bytes, thus getting skipped.
}

func TestBuilderReset(t *testing.T) {
	options := []BuilderOption{WithLabelNameMaxLen(64), WithLabelValueMaxLen(256)}
	builder := Metric("test_reset", options...).LabelString("test", "something")

	require.NotNil(t, builder.pool)
	require.True(t, builder.hasMetricName)
	require.True(t, builder.hasLabel)
	require.Equal(t, 64, builder.labelNameMaxLen)
	require.Equal(t, 256, builder.labelValueMaxLen)

	builder.Reset()

	require.Nil(t, builder.pool)
	require.False(t, builder.hasMetricName)
	require.False(t, builder.hasLabel)
	require.Equal(t, 0, builder.labelNameMaxLen)
	require.Equal(t, 0, builder.labelValueMaxLen)
}

func BenchmarkBuilderTestCasesParallel(b *testing.B) {
	for _, tc := range testCases {
		if tc.skipBench {
			continue
		}
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.RunParallel(func(p *testing.PB) {
				for p.Next() {
					doBenchmarkCase(tc.input)
				}
			})
		})
	}
}

func BenchmarkBuilderTestCasesSequential(b *testing.B) {
	for _, tc := range testCases {
		if tc.skipBench {
			continue
		}
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				doBenchmarkCase(tc.input)
			}
		})
	}
}

func doBenchmarkCase(in input) {
	builder := Metric(in.name, WithLabelNameMaxLen(100), WithLabelValueMaxLen(100))
	for _, label := range in.labels {
		addLabelAnyToBuilder(builder, label)
	}
	_ = builder.String()
}

func BenchmarkCompareSequentialFmt(b *testing.B) {
	b.ReportAllocs()

	var (
		host    = "255.255.255.255"
		version = 3
		err     = fmt.Errorf("mayday")
		test    bool
	)

	for range b.N {
		_ = fmt.Sprintf(`some_metric_name_one{host="%s",version="%d",err=%q,test="%t"}`, host, version, err, test)
		_ = fmt.Sprintf(`some_metric_name_two{host="%s",version="%d",err=%q,test="%t"}`, host, version, err, test)
		_ = fmt.Sprintf(`some_metric_name_three{host="%s",version="%d",err=%q,test="%t"}`, host, version, err, test)
		_ = fmt.Sprintf(`some_metric_name_four{host="%s",version="%d",err=%q,test="%t"}`, host, version, err, test)
		_ = fmt.Sprintf(`some_loooooooooooooooooooooooooooooooooonger_metric_name_one{host="%s",version="%d",err=%q,test="%t"}`, host, version, err, test)
		_ = fmt.Sprintf(`some_loooooooooooooooooooooooooooooooooonger_metric_name_two{host="%s",version="%d",err=%q,test="%t"}`, host, version, err, test)
		_ = fmt.Sprintf(`some_loooooooooooooooooooooooooooooooooonger_metric_name_three{host="%s",version="%d",err=%q,test="%t"}`, host, version, err, test)
		_ = fmt.Sprintf(`some_loooooooooooooooooooooooooooooooooonger_metric_name_four{host="%s",version="%d",err=%q,test="%t"}`, host, version, err, test)
	}
}

func BenchmarkCompareParralelFmt(b *testing.B) {
	b.ReportAllocs()

	var (
		host    = "255.255.255.255"
		version = 3
		err     = fmt.Errorf("mayday")
		test    bool
	)

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			_ = fmt.Sprintf(`some_metric_name_one{host="%s",version="%d",err=%q,test="%t"}`, host, version, err, test)
			_ = fmt.Sprintf(`some_metric_name_two{host="%s",version="%d",err=%q,test="%t"}`, host, version, err, test)
			_ = fmt.Sprintf(`some_metric_name_three{host="%s",version="%d",err=%q,test="%t"}`, host, version, err, test)
			_ = fmt.Sprintf(`some_metric_name_four{host="%s",version="%d",err=%q,test="%t"}`, host, version, err, test)
			_ = fmt.Sprintf(`some_loooooooooooooooooooooooooooooooooonger_metric_name_one{host="%s",version="%d",err=%q,test="%t"}`, host, version, err, test)
			_ = fmt.Sprintf(`some_loooooooooooooooooooooooooooooooooonger_metric_name_two{host="%s",version="%d",err=%q,test="%t"}`, host, version, err, test)
			_ = fmt.Sprintf(`some_loooooooooooooooooooooooooooooooooonger_metric_name_three{host="%s",version="%d",err=%q,test="%t"}`, host, version, err, test)
			_ = fmt.Sprintf(`some_loooooooooooooooooooooooooooooooooonger_metric_name_four{host="%s",version="%d",err=%q,test="%t"}`, host, version, err, test)
		}
	})
}

func BenchmarkCompareSequentialVimebu(b *testing.B) {
	b.ReportAllocs()

	var (
		host    = "255.255.255.255"
		version = 3
		err     = fmt.Errorf("mayday")
		test    bool
	)

	for range b.N {
		_ = Metric("some_metric_name_one").LabelTrustedString("host", host).LabelInt("version", version).LabelError(err).LabelBool("test", test).String()
		_ = Metric("some_metric_name_two").LabelTrustedString("host", host).LabelInt("version", version).LabelError(err).LabelBool("test", test).String()
		_ = Metric("some_metric_name_three").LabelTrustedString("host", host).LabelInt("version", version).LabelError(err).LabelBool("test", test).String()
		_ = Metric("some_metric_name_four").LabelTrustedString("host", host).LabelInt("version", version).LabelError(err).LabelBool("test", test).String()
		_ = Metric("some_loooooooooooooooooooooooooooooooooonger_metric_name_one").LabelTrustedString("host", host).LabelInt("version", version).LabelError(err).LabelBool("test", test).String()
		_ = Metric("some_loooooooooooooooooooooooooooooooooonger_metric_name_two").LabelTrustedString("host", host).LabelInt("version", version).LabelError(err).LabelBool("test", test).String()
		_ = Metric("some_loooooooooooooooooooooooooooooooooonger_metric_name_three").LabelTrustedString("host", host).LabelInt("version", version).LabelError(err).LabelBool("test", test).String()
		_ = Metric("some_loooooooooooooooooooooooooooooooooonger_metric_name_four").LabelTrustedString("host", host).LabelInt("version", version).LabelError(err).LabelBool("test", test).String()
	}
}

func BenchmarkCompareParralelVimebu(b *testing.B) {
	b.ReportAllocs()

	var (
		host    = "255.255.255.255"
		version = 3
		err     = fmt.Errorf("mayday")
		test    bool
	)

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			_ = Metric("some_metric_name_one").LabelTrustedString("host", host).LabelInt("version", version).LabelError(err).LabelBool("test", test).String()
			_ = Metric("some_metric_name_two").LabelTrustedString("host", host).LabelInt("version", version).LabelError(err).LabelBool("test", test).String()
			_ = Metric("some_metric_name_three").LabelTrustedString("host", host).LabelInt("version", version).LabelError(err).LabelBool("test", test).String()
			_ = Metric("some_metric_name_four").LabelTrustedString("host", host).LabelInt("version", version).LabelError(err).LabelBool("test", test).String()
			_ = Metric("some_loooooooooooooooooooooooooooooooooonger_metric_name_one").LabelTrustedString("host", host).LabelInt("version", version).LabelError(err).LabelBool("test", test).String()
			_ = Metric("some_loooooooooooooooooooooooooooooooooonger_metric_name_two").LabelTrustedString("host", host).LabelInt("version", version).LabelError(err).LabelBool("test", test).String()
			_ = Metric("some_loooooooooooooooooooooooooooooooooonger_metric_name_three").LabelTrustedString("host", host).LabelInt("version", version).LabelError(err).LabelBool("test", test).String()
			_ = Metric("some_loooooooooooooooooooooooooooooooooonger_metric_name_four").LabelTrustedString("host", host).LabelInt("version", version).LabelError(err).LabelBool("test", test).String()
		}
	})
}
