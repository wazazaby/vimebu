package vimebu

import (
	"bytes"
	"strconv"
	"strings"
)

const (
	MetricNameMaxLen = 256  // MetricNameMaxLen is the maximum len in bytes allowed for the metric name.
	LabelNameMaxLen  = 128  // LabelNameMaxLen is the maximum len in bytes allowed for a label name.
	LabelValueLen    = 1024 // LabelValueLen is the maximum len in bytes allowed for a label value.
)

// Builder is used to efficiently build a VictoriaMetrics or Prometheus metric.
// It's backed by a bytes.Buffer to minimize memory copying.
// The zero value is ready to use.
type Builder struct {
	underlying      bytes.Buffer
	flName, flLabel bool
}

// Metric creates a new Builder.
// It can be useful if you want to create a metric in a single line.
func Metric(name string) *Builder {
	return (&Builder{}).Metric(name)
}

// Metric sets the metric name of the Builder.
// NoOp if called more than once for the same builder or if the name is empty.
// Panics if the name contains more than [vimebu.MetricNameMaxLen] bytes or if it contains a double quote.
func (b *Builder) Metric(name string) *Builder {
	if b.flName || name == "" {
		return b
	}

	if len(name) > MetricNameMaxLen {
		panic("metric name contains too many bytes")
	}

	if strings.Contains(name, `"`) {
		panic("metric name should not contain double quotes")
	}

	b.underlying.WriteString(name + "{")
	b.flName = true

	return b
}

// Label appends a pair of label and label value to the Builder.
// NoOp if the label or label value are empty.
// Panics if the label name or label value contains more than [vimebu.LabelNameMaxLen] or [vimebu.LabelValueMaxLen] bytes respectively.
func (b *Builder) Label(label, value string) *Builder {
	if !b.flName || label == "" || value == "" {
		return b
	}

	if len(label) > LabelNameMaxLen {
		panic("label name contains too many bytes")
	}

	if len(value) > LabelValueLen {
		panic("label value contains too many bytes")
	}

	if b.flLabel {
		b.underlying.WriteString("," + label + "=")
	} else {
		b.underlying.WriteString(label + "=")
		b.flLabel = true
	}

	buf := b.underlying.AvailableBuffer()
	quoted := strconv.AppendQuote(buf, value)
	b.underlying.Write(quoted)

	return b
}

// String builds the metric by returning the accumulated string.
func (b *Builder) String() string {
	if !b.flName {
		return ""
	}
	b.underlying.WriteString("}")
	return b.underlying.String()
}

// Reset resets the Builder to be empty.
func (b *Builder) Reset() {
	b.flName, b.flLabel = false, false
	b.underlying.Reset()
}

// Grow exposes the underlying builder's Grow method for preallocation purposes.
//
// Please see [bytes.Buffer.Grow].
func (b *Builder) Grow(n int) {
	b.underlying.Grow(n)
}
