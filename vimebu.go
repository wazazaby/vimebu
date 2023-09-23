package vimebu

import (
	"fmt"
	"strings"
)

// Builder is used to efficiently build a VictoriaMetrics or Prometheus metric.
// It's backed by a strings.Builder to minimize memory copying.
// The zero value is ready to use.
type Builder struct {
	name       string
	underlying strings.Builder
	labels     []pair
	size       int // size is used as a counter to know how many bytes we need to preallocate to the strings.Builder buffer
}

// NewBuilder creates a new Builder.
// It can be useful if you want to create a metric in a single line.
func NewBuilder() *Builder {
	return &Builder{}
}

// Metric TODO:
// NoOp
func (b *Builder) Metric(name string) *Builder {
	if b.name != "" || name == "" {
		return b
	}
	b.size += len(name)
	b.name = name
	return b
}

// pair TODO:
type pair struct {
	label, value string
}

// Label TODO:
func (b *Builder) Label(label, value string) *Builder {
	if label == "" || value == "" {
		return b
	}
	b.size += len(label + value)
	b.labels = append(b.labels, pair{label, value})
	return b
}

// String TODO:
func (b *Builder) String() string {
	if b.name == "" {
		return ""
	}

	b.underlying.Grow(b.calculatePrealloc())
	b.underlying.WriteString(b.name + `{`)

	first := true
	for _, pair := range b.labels {
		if first {
			first = false
		} else {
			b.underlying.WriteString(`,`)
		}

		b.underlying.WriteString(pair.label + `="` + pair.value + `"`)
	}

	b.underlying.WriteString(`}`)

	return b.underlying.String()
}

const (
	commaLen       = 1
	curlyBraceslen = 2
	equalQuotesLen = 3
)

func (b *Builder) calculatePrealloc() int {
	if n := len(b.labels); n > 0 {
		return b.size + curlyBraceslen + (n * equalQuotesLen) + n - commaLen
	}
	return b.size + curlyBraceslen
}

var _ fmt.Stringer = (*Builder)(nil)
