package vimebu

import (
	"bytes"
	"strconv"
	"strings"
)

// Builder is used to efficiently build a VictoriaMetrics metric.
// It's backed by a bytes.Buffer to minimize memory copying.
//
// The zero value is ready to use.
type Builder struct {
	underlying      *bytes.Buffer
	flName, flLabel bool
}

func (b *Builder) init() {
	if b.underlying == nil {
		b.underlying = getBuffer()
	}
}

// Metric creates a new Builder.
// It can be useful if you want to create a metric in a single line.
func Metric(name string) *Builder {
	return (&Builder{}).Metric(name)
}

// Metric sets the metric name of the Builder.
//
// NoOp if called more than once for the same builder.
//
// Panics if the name is empty, contains more than [vimebu.MetricNameMaxLen] bytes or if it contains a double quote.
func (b *Builder) Metric(name string) *Builder {
	if b.flName {
		return b
	}

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

	b.init()

	b.underlying.WriteString(name)
	b.underlying.WriteByte(leftBracketByte)
	b.flName = true

	return b
}

// LabelQuote appends a pair of label name and label value to the Builder. Quotes inside the label value will be escaped.
//
// Panics if the label name or label value contain more than [vimebu.LabelNameMaxLen] or [vimebu.LabelValueMaxLen] bytes respectively.
//
// NoOp if the label name or label value are empty.
func (b *Builder) LabelQuote(name, value string) *Builder {
	return b.label(name, value, true)
}

// Label appends a pair of label name and label value to the Builder.
// Unlike [vimebu.Builder.LabelQuote], quotes inside the label value will not be escaped.
// It's better suited for a label value where you control the input (either it is already sanitized, or it comes from a const or an enum for example).
//
// Panics if the label name or label value contain more than [vimebu.LabelNameMaxLen] or [vimebu.LabelValueMaxLen] bytes respectively.
//
// NoOp if the label name or label value are empty.
func (b *Builder) Label(name, value string) *Builder {
	return b.label(name, value, false)
}

func (b *Builder) label(name, value string, escapeQuote bool) *Builder {
	if !b.flName || name == "" || value == "" {
		return b
	}

	if len(name) > LabelNameMaxLen {
		panic("label name contains too many bytes")
	}

	if len(value) > LabelValueLen {
		panic("label value contains too many bytes")
	}

	if b.flLabel { // If we already wrote a label, start writing commas before label names.
		b.underlying.WriteByte(commaByte)
	} else { // Otherwise, mark flag as true for next pass.
		b.flLabel = true
	}

	b.underlying.WriteString(name)
	b.underlying.WriteByte(equalByte)
	if escapeQuote && strings.Contains(value, `"`) { // If we need to escape quotes in the label value.
		buf := b.underlying.AvailableBuffer()
		quoted := strconv.AppendQuote(buf, value)
		b.underlying.Write(quoted)
	} else { // Otherwise, just wrap the label value inside a pair of double quotes.
		b.underlying.WriteByte(doubleQuotesByte)
		b.underlying.WriteString(value)
		b.underlying.WriteByte(doubleQuotesByte)
	}

	return b
}

// String builds the metric by returning the accumulated string.
func (b *Builder) String() string {
	defer putBuffer(b.underlying)
	if !b.flName {
		return ""
	}

	b.underlying.WriteByte(rightBracketByte)
	return b.underlying.String()
}

// Reset resets the Builder to be empty.
func (b *Builder) Reset() {
	b.flName, b.flLabel = false, false
	putBuffer(b.underlying)
}

// Grow exposes the underlying buffer's Grow method for preallocation purposes.
//
// It can be useful is you already know the size of your metric, including labels.
//
// Please see [bytes.Buffer.Grow].
func (b *Builder) Grow(n int) *Builder {
	b.init()
	b.underlying.Grow(n)
	return b
}
