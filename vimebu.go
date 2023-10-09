package vimebu

import (
	"bytes"
	"log"
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
// NoOp if :
// * called more than once for the same builder instance.
// * the name is empty or contains more than [vimebu.MetricNameMaxLen] bytes.
// * the name contains a double quote.
func (b *Builder) Metric(name string) *Builder {
	if b.flName {
		log.Println("metric has already been called for this builder, skipping")
		return b
	}

	ln := len(name)
	if ln == 0 {
		log.Println("metric name must not be empty, skipping")
		return b
	}
	if ln > MetricNameMaxLen {
		log.Println("metric name contains too many bytes, skipping ")
		return b
	}

	if strings.Contains(name, `"`) {
		log.Println("metric name contains double quotes, skipping")
		return b
	}

	b.init()

	b.underlying.WriteString(name)
	b.underlying.WriteByte(leftBracketByte)
	b.flName = true

	return b
}

// LabelQuote appends a pair of label name and label value to the Builder. Quotes inside the label value will be escaped.
//
// NoOp if :
// * no metric name has been set using [vimebu.Builder.Metric].
// * the label name is empty or contains more than [vimebu.LabelNameMaxLen].
// * the label value is empty or contains more than [vimebu.LabelValueMaxLen].
func (b *Builder) LabelQuote(name, value string) *Builder {
	return b.label(name, value, true)
}

// Label appends a pair of label name and label value to the Builder.
// Unlike [vimebu.Builder.LabelQuote], quotes inside the label value will not be escaped.
// It's better suited for a label value where you control the input (either it is already sanitized, or it comes from a const or an enum for example).
//
// NoOp if :
// * no metric name has been set using [vimebu.Builder.Metric].
// * the label name is empty or contains more than [vimebu.LabelNameMaxLen].
// * the label value is empty or contains more than [vimebu.LabelValueMaxLen].
func (b *Builder) Label(name, value string) *Builder {
	return b.label(name, value, false)
}

func (b *Builder) label(name, value string, escapeQuote bool) *Builder {
	if !b.flName {
		log.Println("metric has not been called on this builder, skipping")
		return b
	}

	ln := len(name)
	if ln == 0 {
		log.Println("label name must not be empty, skipping")
		return b
	}
	if ln > LabelNameMaxLen {
		log.Println("label name contains too many bytes, skipping")
		return b
	}

	lv := len(value)
	if lv == 0 {
		log.Println("label value must not be empty, skipping")
		return b
	}
	if lv > LabelValueLen {
		log.Println("label value contains too many bytes, skipping")
		return b
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
