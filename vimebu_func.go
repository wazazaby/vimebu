package vimebu

import (
	"strconv"
	"strings"
)

type LabelCallback func() (name, value string, escapeQuote bool)

// WithLabel appends a pair of label name and label value to the builder.
// Unlike [vimebu.WithLabelQuote], quotes inside the label value will not be escaped.
// It's better suited for a label value where you control the input (either it is already sanitized, or it comes from a const or an enum for example).
//
// Panics if the label name or label value contain more than [vimebu.LabelNameMaxLen] or [vimebu.LabelValueMaxLen] bytes respectively.
func WithLabel(name, value string) LabelCallback {
	if len(name) > LabelNameMaxLen {
		panic("label name contains too many bytes")
	}

	if len(value) > LabelValueLen {
		panic("label value contains too many bytes")
	}

	return func() (string, string, bool) {
		return name, value, false
	}
}

// WithLabelCond is a wrapper around [vimebu.WithLabel].
// It allows you to conditionally add a label depending on the output of a predicate function.
//
// Panics if the label name or label value contain more than [vimebu.LabelNameMaxLen] or [vimebu.LabelValueMaxLen] bytes respectively.
func WithLabelCond(fn func() bool, name, value string) LabelCallback {
	if fn() {
		return WithLabel(name, value)
	}
	return WithLabel("", "")
}

// WithLabelQuote appends a pair of label name and label value to the builder. Quotes inside the label value will be escaped.
//
// Panics if the label name or label value contain more than [vimebu.LabelNameMaxLen] or [vimebu.LabelValueMaxLen] bytes respectively.
func WithLabelQuote(name, value string) LabelCallback {
	if len(name) > LabelNameMaxLen {
		panic("label name contains too many bytes")
	}

	if len(value) > LabelValueLen {
		panic("label value contains too many bytes")
	}

	return func() (string, string, bool) {
		return name, value, true
	}
}

// WithLabelQuoteCond is a wrapper around [vimebu.WithLabelQuote].
// It allows you to conditionally add a label depending on the output of a predicate function.
//
// Panics if the label name or label value contain more than [vimebu.LabelNameMaxLen] or [vimebu.LabelValueMaxLen] bytes respectively.
func WithLabelQuoteCond(fn func() bool, name, value string) LabelCallback {
	if fn() {
		return WithLabelQuote(name, value)
	}
	return WithLabelQuote("", "")
}

// BuilderFunc is used to efficiently build a VictoriaMetrics metric.
// It's backed by a bytes.Buffer to minimize memory copying.
//
// Panics if the name is empty, contains more than [vimebu.MetricNameMaxLen] bytes or if it contains a double quote.
func BuilderFunc(name string, labels ...LabelCallback) string {
	ln := len(name)
	if ln == 0 {
		panic("metric name must not be empty")
	}

	if ln > MetricNameMaxLen {
		panic("metric name contains too many bytes")
	}

	if strings.Contains(name, `"`) {
		panic("metric name should not contain double quotes")
	}

	// In case there are no labels, using a strings.Builder is actually
	// about ~9% faster than concatenating the name and brackets.
	if len(labels) == 0 {
		var b strings.Builder
		b.Grow(ln + 2)
		b.WriteString(name)
		b.WriteString("{}")
		return b.String()
	}

	b := getBuffer()
	defer putBuffer(b)

	b.WriteString(name)
	b.WriteByte(leftBracketByte)
	var flLabel bool
	for _, callback := range labels {
		name, value, escapeQuote := callback()
		if name == "" || value == "" {
			continue
		}

		if flLabel {
			b.WriteByte(commaByte)
		} else {
			flLabel = true
		}

		b.WriteString(name)
		b.WriteByte(equalByte)
		if escapeQuote && strings.Contains(value, `"`) {
			buf := b.AvailableBuffer()
			quoted := strconv.AppendQuote(buf, value)
			b.Write(quoted)
		} else {
			b.WriteByte(doubleQuotesByte)
			b.WriteString(value)
			b.WriteByte(doubleQuotesByte)
		}
	}

	b.WriteByte(rightBracketByte)
	return b.String()
}
