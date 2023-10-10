package vimebu

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"strings"
)

const (
	MetricNameMaxLen = 256  // MetricNameMaxLen is the maximum len in bytes allowed for the metric name.
	LabelNameMaxLen  = 128  // LabelNameMaxLen is the maximum len in bytes allowed for a label name.
	LabelValueLen    = 1024 // LabelValueLen is the maximum len in bytes allowed for a label value.

	leftBracketByte  = byte('{')
	rightBracketByte = byte('}')
	commaByte        = byte(',')
	equalByte        = byte('=')
	doubleQuotesByte = byte('"')
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
//   - called more than once for the same builder instance.
//   - the name is empty or contains more than [vimebu.MetricNameMaxLen] bytes.
//   - the name contains a double quote.
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
//   - no metric name has been set using [vimebu.Builder.Metric].
//   - the label name is empty or contains more than [vimebu.LabelNameMaxLen].
//   - the label value is empty or contains more than [vimebu.LabelValueMaxLen].
func (b *Builder) LabelQuote(name, value string) *Builder {
	return b.label(name, true, &value, nil, nil, nil)
}

// Label appends a pair of label name and label value to the Builder.
// Unlike [vimebu.Builder.LabelQuote], quotes inside the label value will not be escaped.
// It's better suited for a label value where you control the input (either it is already sanitized, or it comes from a const or an enum for example).
//
// NoOp if :
//   - no metric name has been set using [vimebu.Builder.Metric].
//   - the label name is empty or contains more than [vimebu.LabelNameMaxLen].
//   - the label value is empty or contains more than [vimebu.LabelValueMaxLen].
func (b *Builder) Label(name, value string) *Builder {
	return b.label(name, false, &value, nil, nil, nil)
}

// LabelBool appends a pair of label name and boolean label value to the Builder.
//
// NoOp if :
//   - no metric name has been set using [vimebu.Builder.Metric].
//   - the label name is empty or contains more than [vimebu.LabelNameMaxLen].
func (b *Builder) LabelBool(name string, value bool) *Builder {
	return b.label(name, false, nil, &value, nil, nil)
}

// LabelInt appends a pair of label name and int64 label value to the Builder.
//
// NoOp if :
//   - no metric name has been set using [vimebu.Builder.Metric].
//   - the label name is empty or contains more than [vimebu.LabelNameMaxLen].
func (b *Builder) LabelInt(name string, value int64) *Builder {
	return b.label(name, false, nil, nil, &value, nil)
}

// LabelFloat appends a pair of label name and float64 label value to the Builder.
//
// NoOp if :
//   - no metric name has been set using [vimebu.Builder.Metric].
//   - the label name is empty or contains more than [vimebu.LabelNameMaxLen].
func (b *Builder) LabelFloat(name string, value float64) *Builder {
	return b.label(name, false, nil, nil, nil, &value)
}

// LabelStringer appends a pair of label name and label value (implementing [fmt.Stringer]) to the Builder.
// Unlike [vimebu.Builder.LabelStringerQuote], quotes inside the label value will not be escaped.
// It's better suited for a label value where you control the input (either it is already sanitized, or it comes from a const or an enum for example).
//
// NoOp if :
//   - no metric name has been set using [vimebu.Builder.Metric].
//   - the label name is empty or contains more than [vimebu.LabelNameMaxLen].
func (b *Builder) LabelStringer(name string, value fmt.Stringer) *Builder {
	s := value.String()
	return b.label(name, false, &s, nil, nil, nil)
}

// LabelStringerQuote appends a pair of label name and label value (implementing [fmt.Stringer]) to the Builder.
// Quotes inside the label value will be escaped.
//
// NoOp if :
//   - no metric name has been set using [vimebu.Builder.Metric].
//   - the label name is empty or contains more than [vimebu.LabelNameMaxLen].
func (b *Builder) LabelStringerQuote(name string, value fmt.Stringer) *Builder {
	s := value.String()
	return b.label(name, true, &s, nil, nil, nil)
}

func (b *Builder) label(name string, escapeQuote bool, stringValue *string, boolValue *bool, int64Value *int64, float64Value *float64) *Builder {
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

	// String values need to be checked separately as they can be invalid (empty, or contain too many bytes).
	if stringValue != nil {
		lv := len(*stringValue)
		if lv == 0 {
			log.Println("label value must not be empty, skipping")
			return b
		}
		if lv > LabelValueLen {
			log.Println("label value contains too many bytes, skipping")
			return b
		}
	}

	if b.flLabel { // If we already wrote a label, start writing commas before label names.
		b.underlying.WriteByte(commaByte)
	} else { // Otherwise, mark flag as true for next pass.
		b.flLabel = true
	}

	b.underlying.WriteString(name)
	b.underlying.WriteByte(equalByte)
	switch {
	case stringValue != nil:
		b.appendString(*stringValue, escapeQuote)
	case boolValue != nil:
		b.appendBool(*boolValue)
	case int64Value != nil:
		b.appendInt64(*int64Value)
	case float64Value != nil:
		b.appendFloat64(*float64Value)
	default: // Internal problem (wrong use of the label function), panic.
		panic("unsupported case - no label value set")
	}

	return b
}

// appendString quotes (if needed) and appends s to b's underlying buffer.
func (b *Builder) appendString(s string, escapeQuote bool) {
	buf := b.underlying.AvailableBuffer()
	if escapeQuote && strings.Contains(s, `"`) {
		buf = strconv.AppendQuote(buf, s)
	} else {
		buf = append(buf, doubleQuotesByte)
		buf = append(buf, s...)
		buf = append(buf, doubleQuotesByte)
	}
	b.underlying.Write(buf)
}

// appendBool appends bl to b's underlying buffer.
func (b *Builder) appendBool(bl bool) {
	buf := b.underlying.AvailableBuffer()
	buf = append(buf, doubleQuotesByte)
	buf = strconv.AppendBool(buf, bl)
	buf = append(buf, doubleQuotesByte)
	b.underlying.Write(buf)
}

// appendInt64 appends i to b's underlying buffer.
func (b *Builder) appendInt64(i int64) {
	buf := b.underlying.AvailableBuffer()
	buf = append(buf, doubleQuotesByte)
	buf = strconv.AppendInt(buf, i, 10)
	buf = append(buf, doubleQuotesByte)
	b.underlying.Write(buf)
}

// appendFloat64 appends f to b's underlying buffer.
func (b *Builder) appendFloat64(f float64) {
	buf := b.underlying.AvailableBuffer()
	buf = append(buf, doubleQuotesByte)
	buf = strconv.AppendFloat(buf, f, 'f', -1, 64)
	buf = append(buf, doubleQuotesByte)
	b.underlying.Write(buf)
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
