package vimebu

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type label struct {
	name, value string
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
		expected: `http_request_duration_seconds{}`,
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
		expected:  `api_http_requests_total{}`,
		skipBench: true,
	},
	{
		name: "label value contains too many bytes",
		input: input{
			name:   "api_http_requests_total",
			labels: []label{{"test", strings.Repeat("b", 2048), false}},
		},
		expected:  `api_http_requests_total{}`,
		skipBench: true,
	},
}

func handleTestCase(t *testing.T, tc testCase) {
	var b Builder

	b.Metric(tc.input.name)

	for _, label := range tc.input.labels {
		if label.shouldQuote {
			b.LabelQuote(label.name, label.value)
		} else {
			b.Label(label.name, label.value)
		}
	}

	result := b.String()
	require.Equal(t, tc.expected, result)

	t.Run("reset", func(t *testing.T) {
		b.Reset()
		require.False(t, b.flName)
		require.False(t, b.flLabel)
	})
}

func TestBuilder(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.NotPanics(t, func() {
				handleTestCase(t, tc)
			})
		})
	}
}

var (
	status  = "Bad Request"
	path    = `some/path/"with"/quo"tes`
	host    = "1.2.3.4"
	cluster = "guava"
)

func BenchmarkBuilder(b *testing.B) {
	for n := 0; n < b.N; n++ {
		var builder Builder
		_ = builder.Metric("http_request_duration_seconds").
			Label("status", status).
			LabelQuote("path", path).
			Label("host", host).
			Label("cluster", cluster).
			String()
	}
}

func BenchmarkBuilderAppendQuoteNone(b *testing.B) {
	pathSafe := strconv.Quote(path)
	for n := 0; n < b.N; n++ {
		var builder Builder
		_ = builder.Metric("http_request_duration_seconds").
			Label("status", status).
			Label("path", pathSafe).
			Label("host", host).
			Label("cluster", cluster).
			String()
	}
}

func BenchmarkBuilderAppendQuoteOnly(b *testing.B) {
	for n := 0; n < b.N; n++ {
		var builder Builder
		_ = builder.Metric("http_request_duration_seconds").
			LabelQuote("status", status).
			LabelQuote("path", path).
			LabelQuote("host", host).
			LabelQuote("cluster", cluster).
			String()
	}
}

func BenchmarkBuilderTestCases(b *testing.B) {
	for _, tc := range testCases {
		if tc.skipBench {
			continue
		}
		b.Run(tc.name, func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				var builder Builder
				builder.Metric(tc.input.name)
				for _, label := range tc.input.labels {
					if label.shouldQuote {
						builder.LabelQuote(label.name, label.value)
					} else {
						builder.Label(label.name, label.value)
					}
				}
				_ = builder.String()
			}
		})
	}
}
