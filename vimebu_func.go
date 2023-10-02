package vimebu

import (
	"strconv"
	"strings"
)

// LabelCallback TODO:
type LabelCallback func() (name, value string, escapeQuote bool)

// WithLabel TODO:
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

// WithLabelCond TODO:
func WithLabelCond(fn func() bool, name, value string) LabelCallback {
	if fn() {
		return WithLabel(name, value)
	}
	return WithLabel("", "")
}

// WithLabelQuote TODO:
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

// WithLabelQuoteCond TODO:
func WithLabelQuoteCond(fn func() bool, name, value string) LabelCallback {
	if fn() {
		return WithLabelQuote(name, value)
	}
	return WithLabelQuote("", "")
}

// BuilderFunc is used to efficiently build a VictoriaMetrics metric.
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
