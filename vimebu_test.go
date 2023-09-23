package vimebu

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuilder(t *testing.T) {
	type testCase struct {
		it    string
		input struct {
			labels map[string]string
			name   string
		}
		expected string
	}

	testCases := []testCase{
		{
			it: "metric with labels",
			input: struct {
				labels map[string]string
				name   string
			}{
				labels: map[string]string{
					"cluster": "guava",
					"host":    "1.2.3.4",
				},
				name: "cassandra_query_count",
			},
			expected: `cassandra_query_count{cluster="guava",host="1.2.3.4"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.it, func(t *testing.T) {
			t.Run("one line", func(t *testing.T) {
				result := NewBuilder().Metric(tc.input.name).Labels(tc.input.labels).String()
				require.Len(t, result, len(tc.expected))
			})
			t.Run("verbose", func(t *testing.T) {
				var b Builder
				b.Metric(tc.input.name)
				for label, value := range tc.input.labels {
					b.Label(label, value)
				}
				result := b.String()
				require.Len(t, result, len(tc.expected))
			})
		})
	}
}
