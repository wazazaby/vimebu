package vimebu

import (
	"fmt"
	"strings"
)

// Builder is used to efficiently build a VictoriaMetrics or Prometheus metric.
// It's backed by a strings.Builder to minimize memory copying.
// The zero value is ready to use.
type Builder struct {
	labels     map[string]string
	name       string
	underlying strings.Builder
	size       int // size is used as a counter to know how many bytes we need to preallocate to the strings.Builder buffer
}

// NewBuilder creates a new Builder.
// It can be useful if you want to create a metric in a single line.
func NewBuilder() *Builder {
	return &Builder{}
}

// Metric sets the metric name of the builder.
// NoOp if called more than once or if the name is empty.
func (b *Builder) Metric(name string) *Builder {
	if b.name != "" || name == "" {
		return b
	}
	b.size += len(name)
	b.name = name
	return b
}

// Label appends a pair of label and label value to the builder.
// NoOp if the metric name, the label or label value are empty.
func (b *Builder) Label(label, value string) *Builder {
	if b.name == "" || label == "" || value == "" {
		return b
	}
	if b.labels == nil {
		b.labels = make(map[string]string)
	}
	b.size += len(label + value)
	b.labels[label] = value
	return b
}

// String builds the metric by returning the accumulated string.
// Returns an empty string if the metric name is empty.
func (b *Builder) String() string {
	if b.name == "" {
		return ""
	}

	// Preallocate the underlying builder's buffer.
	b.underlying.Grow(b.calculatePrealloc())

	b.underlying.WriteString(b.name + "{")

	// Skip writing a comma after the label / value pair on the first iteration.
	first := true
	for label, value := range b.labels {
		if first {
			first = false
		} else {
			b.underlying.WriteString(",")
		}

		b.underlying.WriteString(label + `="` + value + `"`)
	}

	b.underlying.WriteString("}")

	return b.underlying.String()
}

const (
	commaLen       = 1 // commaLen is the length in bytes of a comma.
	curlyBraceslen = 2 // curlyBracesLen is the length in bytes of a pair of curly braces.
	equalQuotesLen = 3 // equalQuotesLen is the length in bytes of an equal symbol and a pair of double quotes.
)

// calculatePrealloc calculates the amount of bytes needed for this metric.
// The result can be used to preallocate the underlying builder's buffer.
func (b *Builder) calculatePrealloc() int {
	if n := len(b.labels); n > 0 {
		return b.size + curlyBraceslen + (n * equalQuotesLen) + n - commaLen
	}
	return b.size + curlyBraceslen
}

var _ fmt.Stringer = (*Builder)(nil)
