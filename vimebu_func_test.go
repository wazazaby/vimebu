package vimebu

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func handleTestCaseFunc(t *testing.T, tc testCase) {
	labels := make([]Label, 0, len(tc.input.labels))

	for _, label := range tc.input.labels {
		switch v := label.value.(type) {
		case string:
			if label.shouldQuote {
				labels = append(labels, LabelStringQuote(label.name, v))
			} else {
				labels = append(labels, LabelString(label.name, v))
			}
		case bool:
			labels = append(labels, LabelBool(label.name, v))
		case uint8:
			labels = append(labels, LabelUint(label.name, v))
		case uint16:
			labels = append(labels, LabelUint(label.name, v))
		case uint32:
			labels = append(labels, LabelUint(label.name, v))
		case uint64:
			labels = append(labels, LabelUint(label.name, v))
		case uint:
			labels = append(labels, LabelUint(label.name, v))
		case int8:
			labels = append(labels, LabelInt(label.name, v))
		case int16:
			labels = append(labels, LabelInt(label.name, v))
		case int32:
			labels = append(labels, LabelInt(label.name, v))
		case int64:
			labels = append(labels, LabelInt(label.name, v))
		case int:
			labels = append(labels, LabelInt(label.name, v))
		case float32:
			labels = append(labels, LabelFloat(label.name, v))
		case float64:
			labels = append(labels, LabelFloat(label.name, v))
		case fmt.Stringer:
			if label.shouldQuote {
				labels = append(labels, LabelStringerQuote(label.name, v))
			} else {
				labels = append(labels, LabelStringer(label.name, v))
			}
		default:
			panic(fmt.Sprintf("unsupported type %T", v))
		}
	}

	result := BuilderFunc(tc.input.name, labels...)
	require.Equal(t, tc.expected, result)
}

func TestBuilderFunc(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.NotPanics(t, func() {
				handleTestCaseFunc(t, tc)
			})
		})
	}
}

func TestLabelCond(t *testing.T) {
	result := BuilderFunc("test_cond",
		LabelStringCond("should", "appear", func() bool { return true }),
		LabelStringCond("should", "not", func() bool { return false }),
		LabelStringQuoteCond("present", `/and/"quoted"`, func() bool { return true }),
		LabelStringQuoteCond("absent", `/and/"quoted"`, func() bool { return false }),
		LabelFloatCond("nofloat", 66.7, func() bool { return false }),
		LabelBoolCond("nobool", true, func() bool { return false }),
		LabelIntCond("int", 1234, func() bool { return true }),
	)
	require.Equal(t, `test_cond{should="appear",present="/and/\"quoted\"",int="1234"}`, result)
}

func BenchmarkBuilderFuncFast(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_ = BuilderFunc("test_with_builder_func",
			LabelString("test", "teddy"),
			LabelStringQuote("quote", `"test"`),
			LabelBool("compressed", true),
			LabelFloat("float", 12.3),
			LabelUint("uint", uint8(69)),
			LabelInt("int", 667),
			LabelInt2("int2", 667),
		)
	}
}

func BenchmarkBuilderFuncTestCases(b *testing.B) {
	for _, tc := range testCases {
		if tc.skipBench {
			continue
		}
		b.Run(tc.name, func(b *testing.B) {
			labels := make([]Label, 0, len(tc.input.labels))

			for _, label := range tc.input.labels {
				switch v := label.value.(type) {
				case string:
					if label.shouldQuote {
						labels = append(labels, LabelStringQuote(label.name, v))
					} else {
						labels = append(labels, LabelString(label.name, v))
					}
				case bool:
					labels = append(labels, LabelBool(label.name, v))
				case uint8:
					labels = append(labels, LabelUint(label.name, v))
				case uint16:
					labels = append(labels, LabelUint(label.name, v))
				case uint32:
					labels = append(labels, LabelUint(label.name, v))
				case uint64:
					labels = append(labels, LabelUint(label.name, v))
				case uint:
					labels = append(labels, LabelUint(label.name, v))
				case int8:
					labels = append(labels, LabelInt(label.name, v))
				case int16:
					labels = append(labels, LabelInt(label.name, v))
				case int32:
					labels = append(labels, LabelInt(label.name, v))
				case int64:
					labels = append(labels, LabelInt(label.name, v))
				case int:
					labels = append(labels, LabelInt(label.name, v))
				case float32:
					labels = append(labels, LabelFloat(label.name, v))
				case float64:
					labels = append(labels, LabelFloat(label.name, v))
				case fmt.Stringer:
					if label.shouldQuote {
						labels = append(labels, LabelStringerQuote(label.name, v))
					} else {
						labels = append(labels, LabelStringer(label.name, v))
					}
				default:
					panic(fmt.Sprintf("unsupported type %T", v))
				}
			}

			for n := 0; n < b.N; n++ {
				_ = BuilderFunc(tc.input.name, labels...)
			}
		})
	}
}
