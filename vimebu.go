package vimebu

import (
	"fmt"
	"strings"
)

// Builder is used to efficiently build a VictoriaMetrics or Prometheus metric.
// It's backed by a strings.Builder to minimize memory copying.
// The zero value is ready to use.
type Builder struct {
	underlying strings.Builder
	flLabel    bool // flLabel is set to true when the first label is written, so we can know later on if a comma is needed or not
}

// Metric creates a new Builder.
// It can be useful if you want to create a metric in a single line.
func Metric(name string) *Builder {
	return (&Builder{}).Metric(name)
}

// Metric sets the metric name of the Builder.
// NoOp if the name is empty.
func (b *Builder) Metric(name string) *Builder {
	if name == "" {
		return b
	}

	b.underlying.WriteString(name + "{")

	return b
}

// Label appends a pair of label and label value to the Builder.
// NoOp if the label or label value are empty.
func (b *Builder) Label(label, value string) *Builder {
	if label == "" || value == "" {
		return b
	}

	if b.flLabel {
		b.underlying.WriteString("," + label + `="` + value + `"`)
	} else {
		b.underlying.WriteString(label + `="` + value + `"`)
		b.flLabel = true
	}

	return b
}

// Labels appends multiple labels and label values to the Builder.
// NoOp if the map is empty.
// Pairs containing an empty label or label value will be skipped.
func (b *Builder) Labels(labels map[string]string) *Builder {
	if len(labels) == 0 {
		return b
	}

	for label, value := range labels {
		b.Label(label, value)
	}

	return b
}

// String builds the metric by returning the accumulated string.
func (b *Builder) String() string {
	b.underlying.WriteString("}")
	return b.underlying.String()
}

// Reset resets the Builder to be empty.
func (b *Builder) Reset() {
	b.underlying.Reset()
}

var _ fmt.Stringer = (*Builder)(nil)
