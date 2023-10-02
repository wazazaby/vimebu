package vimebu

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func handleTestCaseFn(t *testing.T, tc testCase) {
	labels := make([]LabelCallback, 0, len(tc.input.labels))

	for _, label := range tc.input.labels {
		if label.shouldQuote {
			labels = append(labels, WithLabelQuote(label.name, label.value))
		} else {
			labels = append(labels, WithLabel(label.name, label.value))
		}
	}

	result := BuilderFunc(tc.input.name, labels...)
	require.Equal(t, tc.expected, result)
}

func TestBuilderFn(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mustPanic {
				require.Panics(t, func() {
					handleTestCaseFn(t, tc)
				})
			} else {
				require.NotPanics(t, func() {
					handleTestCaseFn(t, tc)
				})
			}
		})
	}
}

func TestLabelCond(t *testing.T) {
	result := BuilderFunc("test_cond",
		WithLabelCond(func() bool { return true }, "should", "appear"),
		WithLabelCond(func() bool { return false }, "should", "not"),
		WithLabelQuoteCond(func() bool { return true }, "present", `/and/"quoted"`),
		WithLabelQuoteCond(func() bool { return false }, "absent", `/and/"quoted"`),
	)
	require.Equal(t, `test_cond{should="appear",present="/and/\"quoted\""}`, result)
}

func BenchmarkBuilderFnTestCases(b *testing.B) {
	for _, tc := range testCases {
		if tc.mustPanic { // Skip test cases that panics as they will break the benchmarks.
			continue
		}
		b.Run(tc.name, func(b *testing.B) {
			labels := make([]LabelCallback, 0, len(tc.input.labels))

			for _, label := range tc.input.labels {
				if label.shouldQuote {
					labels = append(labels, WithLabelQuote(label.name, label.value))
				} else {
					labels = append(labels, WithLabel(label.name, label.value))
				}
			}

			for n := 0; n < b.N; n++ {
				_ = BuilderFunc(tc.input.name, labels...)
			}
		})
	}
}
