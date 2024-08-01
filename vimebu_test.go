package vimebu

import (
	"fmt"
	"strconv"
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
		name:      "no name",
		skipBench: true,
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
		name: "metric name contains too many bytes",
		input: input{
			name: strings.Repeat("b", 512),
		},
		skipBench: true,
	},
	{
		name: "label name contains too many bytes",
		input: input{
			name:   "api_http_requests_total",
			labels: []label{{strings.Repeat("b", 256), "test", false}},
		},
		expected:  `api_http_requests_total`,
		skipBench: true,
	},
	{
		name: "label value contains too many bytes",
		input: input{
			name:   "api_http_requests_total",
			labels: []label{{"test", strings.Repeat("b", 2048), false}},
		},
		expected:  `api_http_requests_total`,
		skipBench: true,
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
			},
		},
		expected: `cassandra_query_count{path="/\"some\"/path",is_bidule="true",is_tac="false",point="123.456",num="1234",stringer="spiderman",uint8="128",int="-42"}`,
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

func handleTestCase(t *testing.T, tc testCase) {
	var b Builder

	b.Metric(tc.input.name)

	for _, label := range tc.input.labels {
		switch v := label.value.(type) {
		case string:
			if label.shouldQuote {
				b.LabelQuote(label.name, v)
			} else {
				b.Label(label.name, v)
			}
		case bool:
			b.LabelBool(label.name, v)
		case uint8:
			b.LabelUint8(label.name, v)
		case uint16:
			b.LabelUint16(label.name, v)
		case uint32:
			b.LabelUint32(label.name, v)
		case uint64:
			b.LabelUint64(label.name, v)
		case uint:
			b.LabelUint(label.name, v)
		case int8:
			b.LabelInt8(label.name, v)
		case int16:
			b.LabelInt16(label.name, v)
		case int32:
			b.LabelInt32(label.name, v)
		case int64:
			b.LabelInt64(label.name, v)
		case int:
			b.LabelInt(label.name, v)
		case float32:
			b.LabelFloat32(label.name, v)
		case float64:
			b.LabelFloat64(label.name, v)
		case fmt.Stringer:
			if label.shouldQuote {
				b.LabelStringerQuote(label.name, v)
			} else {
				b.LabelStringer(label.name, v)
			}
		default:
			panic(fmt.Sprintf("unsupported type %T", v))
		}
	}

	result := b.String()
	require.Equal(t, tc.expected, result)
}

func TestBuilder(t *testing.T) {
	t.Parallel()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.NotPanics(t, func() {
				handleTestCase(t, tc)
			})
		})
	}
}

func TestBuilderParallel(t *testing.T) {
	var eg errgroup.Group
	for i := 0; i < 400; i++ {
		i := i
		name := fmt.Sprintf("foobar%d", i)
		eg.Go(func() error {
			require.NotPanics(t, func() {
				Metric(name).
					Label("host", "foobar").
					LabelBool("compressed", false).
					LabelUint8("port", 80).
					LabelFloat32("float", 12.3).
					GetOrCreateCounter().
					Add(300)
			})
			return nil
		})
	}
	require.NoError(t, eg.Wait())
}

var (
	status  = "Bad Request"
	path    = `some/path/"with"/quo"tes`
	host    = "1.2.3.4"
	cluster = "guava"
)

func BenchmarkBuilder(b *testing.B) {
	b.ReportAllocs()

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			var builder Builder
			_ = builder.Metric("http_request_duration_seconds").
				Label("status", status).
				LabelQuote("path", path).
				Label("host", host).
				Label("cluster", cluster).
				String()
		}
	})
}

func BenchmarkBuilderAppendQuoteNone(b *testing.B) {
	b.ReportAllocs()

	pathSafe := strconv.Quote(path)

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			var builder Builder
			_ = builder.Metric("http_request_duration_seconds").
				Label("status", status).
				Label("path", pathSafe).
				Label("host", host).
				Label("cluster", cluster).
				String()
		}
	})
}

func BenchmarkBuilderAppendQuoteOnly(b *testing.B) {
	b.ReportAllocs()

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			var builder Builder
			_ = builder.Metric("http_request_duration_seconds").
				LabelQuote("status", status).
				LabelQuote("path", path).
				LabelQuote("host", host).
				LabelQuote("cluster", cluster).
				String()
		}
	})
}

func BenchmarkBuilderTestCasesParallel(b *testing.B) {
	b.ReportAllocs()

	for _, tc := range testCases {
		if tc.skipBench {
			continue
		}
		b.Run(tc.name, func(b *testing.B) {
			b.RunParallel(func(p *testing.PB) {
				for p.Next() {
					doBenchmarkCase(tc.input)
				}
			})
		})
	}
}

func BenchmarkBuilderTestCasesSequential(b *testing.B) {
	b.ReportAllocs()

	for _, tc := range testCases {
		if tc.skipBench {
			continue
		}
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				doBenchmarkCase(tc.input)
			}
		})
	}
}

func doBenchmarkCase(in input) {
	var builder Builder
	builder.Metric(in.name)
	for _, label := range in.labels {
		switch v := label.value.(type) {
		case string:
			if label.shouldQuote {
				builder.LabelQuote(label.name, v)
			} else {
				builder.Label(label.name, v)
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
				builder.LabelStringerQuote(label.name, v)
			} else {
				builder.LabelStringer(label.name, v)
			}
		default:
			panic(fmt.Sprintf("unsupported type %T", v))
		}
	}
	_ = builder.String()
}
