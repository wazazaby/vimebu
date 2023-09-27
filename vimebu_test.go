package vimebu

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type label struct {
	name, value string
}
type input struct {
	name   string
	labels []label
}
type testCase struct {
	name     string
	expected string
	input    input
}

var testCases = []testCase{
	{
		name: "metric with labels",
		input: input{
			labels: []label{{"cluster", "guava"}, {"host", "1.2.3.4"}},
			name:   "cassandra_query_count",
		},
		expected: `cassandra_query_count{cluster="guava",host="1.2.3.4"}`,
	},
	{
		name: "metric with single label",
		input: input{
			labels: []label{{"type", "std"}},
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
		name:     "no name",
		expected: ``,
	},
	{
		name: "some empty labels and values",
		input: input{
			labels: []label{{"operation", ""}, {"", "1.2.3.4"}, {"status", "OK"}, {"node", ""}},
			name:   "api_http_requests_total",
		},
		expected: `api_http_requests_total{status="OK"}`,
	},
	{
		name: "values contain double quotes",
		input: input{
			labels: []label{{"error", `something went "horribly" wrong`}, {"path", `some/path/"with"/quo"tes`}},
			name:   "api_http_requests_total",
		},
		expected: `api_http_requests_total{error="something went \"horribly\" wrong",path="some/path/\"with\"/quo\"tes"}`,
	},
}

func TestBuilder(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var b Builder
			b.Metric(tc.input.name)
			for _, label := range tc.input.labels {
				b.Label(label.name, label.value)
			}
			result := b.String()
			require.Equal(t, tc.expected, result)

			t.Run("reset", func(t *testing.T) {
				b.Reset()
				require.False(t, b.flName)
				require.False(t, b.flLabel)
			})
		})
	}
}

func BenchmarkBuilder(b *testing.B) {
	for _, tc := range testCases {
		for n := 0; n < b.N; n++ {
			b := Metric(tc.input.name)
			for _, label := range tc.input.labels {
				b.Label(label.name, label.value)
			}
			_ = b.String()
		}
	}
}
