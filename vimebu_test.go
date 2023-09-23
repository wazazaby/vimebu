package vimebu

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuilder(t *testing.T) {
	type input struct {
		labels map[string]string
		name   string
	}
	type testCase struct {
		name     string
		input    input
		expected string
	}

	testCases := []testCase{
		{
			name: "metric with labels",
			input: input{
				labels: map[string]string{
					"cluster": "guava",
					"host":    "1.2.3.4",
				},
				name: "cassandra_query_count",
			},
			expected: `cassandra_query_count{cluster="guava",host="1.2.3.4"}`,
		},
		{
			name: "metric with single label",
			input: input{
				labels: map[string]string{"type": "std"},
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
				labels: map[string]string{
					"operation": "",
					"":          "1.2.3.4",
					"status":    "OK",
					"node":      "",
				},
				name: "api_http_requests_total",
			},
			expected: `api_http_requests_total{status="OK"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Run("one line", func(t *testing.T) {
				result := NewBuilder().Metric(tc.input.name).Labels(tc.input.labels).String()
				// Comparing len of produced string because the labels map order is not guaranteed.
				// Meaning comparing the strings would result in an error on random occasions because of the labels :(
				// Don't know how to improve this, please let me know if you have an idea.
				require.Len(t, result, len(tc.expected))
			})
			t.Run("verbose", func(t *testing.T) {
				var b Builder
				b.Metric(tc.input.name)
				for label, value := range tc.input.labels {
					b.Label(label, value)
				}
				result := b.String()
				// See comment above.
				require.Len(t, result, len(tc.expected))

				t.Run("reset", func(t *testing.T) {
					b.Reset()
					require.Empty(t, b.name)
					require.Empty(t, b.labels)
					require.Empty(t, b.underlying.Cap())
					require.Empty(t, b.underlying.Len())
				})
			})
		})
	}
}
