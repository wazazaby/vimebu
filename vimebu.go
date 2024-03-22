package vimebu

import (
	"fmt"
	"log"
	"strings"
)

var (
	MetricNameMaxLen = 256  // MetricNameMaxLen is the maximum len in bytes allowed for the metric name.
	LabelNameMaxLen  = 128  // LabelNameMaxLen is the maximum len in bytes allowed for a label name.
	LabelValueLen    = 1024 // LabelValueLen is the maximum len in bytes allowed for a label value.
)

const (
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
	*buffer
	flName, flLabel bool
}

func (b *Builder) init() {
	if b.buffer == nil {
		b.buffer = getBuffer()
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

	b.buf = append(b.buf, name...)
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
	return b.label(name, true, &value, nil, nil, nil, nil)
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
	return b.label(name, false, &value, nil, nil, nil, nil)
}

// LabelErrQuote appends a pair of label name and error label value to the Builder. Quotes inside the error label value will be escaped.
//
// NoOp if :
//   - no metric name has been set using [vimebu.Builder.Metric].
//   - the label name is empty or contains more than [vimebu.LabelNameMaxLen].
//   - the label value is empty or contains more than [vimebu.LabelValueMaxLen].
func (b *Builder) LabelErrQuote(name string, err error) *Builder {
	value := err.Error()
	return b.label(name, true, &value, nil, nil, nil, nil)
}

// LabelErr appends a pair of label name and error label value to the Builder.
// Unlike [vimebu.Builder.LabelErrQuote], quotes inside the error label value will not be escaped.
// It's better suited for a label value where you control the input (either it is already sanitized, or it comes from a const or an enum for example).
//
// NoOp if :
//   - no metric name has been set using [vimebu.Builder.Metric].
//   - the label name is empty or contains more than [vimebu.LabelNameMaxLen].
//   - the label value is empty or contains more than [vimebu.LabelValueMaxLen].
func (b *Builder) LabelErr(name string, err error) *Builder {
	value := err.Error()
	return b.label(name, false, &value, nil, nil, nil, nil)
}

// LabelBool appends a pair of label name and boolean label value to the Builder.
//
// NoOp if :
//   - no metric name has been set using [vimebu.Builder.Metric].
//   - the label name is empty or contains more than [vimebu.LabelNameMaxLen].
func (b *Builder) LabelBool(name string, value bool) *Builder {
	return b.label(name, false, nil, &value, nil, nil, nil)
}

// LabelUint appends a pair of label name and uint label value to the Builder.
//
// NoOp if :
//   - no metric name has been set using [vimebu.Builder.Metric].
//   - the label name is empty or contains more than [vimebu.LabelNameMaxLen].
func (b *Builder) LabelUint(name string, value uint) *Builder {
	i := uint64(value)
	return b.label(name, false, nil, nil, &i, nil, nil)
}

// LabelUint8 appends a pair of label name and uint8 label value to the Builder.
//
// NoOp if :
//   - no metric name has been set using [vimebu.Builder.Metric].
//   - the label name is empty or contains more than [vimebu.LabelNameMaxLen].
func (b *Builder) LabelUint8(name string, value uint8) *Builder {
	i := uint64(value)
	return b.label(name, false, nil, nil, &i, nil, nil)
}

// LabelUint16 appends a pair of label name and uint16 label value to the Builder.
//
// NoOp if :
//   - no metric name has been set using [vimebu.Builder.Metric].
//   - the label name is empty or contains more than [vimebu.LabelNameMaxLen].
func (b *Builder) LabelUint16(name string, value uint16) *Builder {
	i := uint64(value)
	return b.label(name, false, nil, nil, &i, nil, nil)
}

// LabelUint32 appends a pair of label name and uint32 label value to the Builder.
//
// NoOp if :
//   - no metric name has been set using [vimebu.Builder.Metric].
//   - the label name is empty or contains more than [vimebu.LabelNameMaxLen].
func (b *Builder) LabelUint32(name string, value uint32) *Builder {
	i := uint64(value)
	return b.label(name, false, nil, nil, &i, nil, nil)
}

// LabelUint64 appends a pair of label name and uint64 label value to the Builder.
//
// NoOp if :
//   - no metric name has been set using [vimebu.Builder.Metric].
//   - the label name is empty or contains more than [vimebu.LabelNameMaxLen].
func (b *Builder) LabelUint64(name string, value uint64) *Builder {
	return b.label(name, false, nil, nil, &value, nil, nil)
}

// LabelInt appends a pair of label name and int label value to the Builder.
//
// NoOp if :
//   - no metric name has been set using [vimebu.Builder.Metric].
//   - the label name is empty or contains more than [vimebu.LabelNameMaxLen].
func (b *Builder) LabelInt(name string, value int) *Builder {
	i := int64(value)
	return b.label(name, false, nil, nil, nil, &i, nil)
}

// LabelInt8 appends a pair of label name and int8 label value to the Builder.
//
// NoOp if :
//   - no metric name has been set using [vimebu.Builder.Metric].
//   - the label name is empty or contains more than [vimebu.LabelNameMaxLen].
func (b *Builder) LabelInt8(name string, value int8) *Builder {
	i := int64(value)
	return b.label(name, false, nil, nil, nil, &i, nil)
}

// LabelInt16 appends a pair of label name and int16 label value to the Builder.
//
// NoOp if :
//   - no metric name has been set using [vimebu.Builder.Metric].
//   - the label name is empty or contains more than [vimebu.LabelNameMaxLen].
func (b *Builder) LabelInt16(name string, value int16) *Builder {
	i := int64(value)
	return b.label(name, false, nil, nil, nil, &i, nil)
}

// LabelInt32 appends a pair of label name and int32 label value to the Builder.
//
// NoOp if :
//   - no metric name has been set using [vimebu.Builder.Metric].
//   - the label name is empty or contains more than [vimebu.LabelNameMaxLen].
func (b *Builder) LabelInt32(name string, value int32) *Builder {
	i := int64(value)
	return b.label(name, false, nil, nil, nil, &i, nil)
}

// LabelInt64 appends a pair of label name and int64 label value to the Builder.
//
// NoOp if :
//   - no metric name has been set using [vimebu.Builder.Metric].
//   - the label name is empty or contains more than [vimebu.LabelNameMaxLen].
func (b *Builder) LabelInt64(name string, value int64) *Builder {
	i := int64(value)
	return b.label(name, false, nil, nil, nil, &i, nil)
}

// LabelFloat32 appends a pair of label name and float32 label value to the Builder.
//
// NoOp if :
//   - no metric name has been set using [vimebu.Builder.Metric].
//   - the label name is empty or contains more than [vimebu.LabelNameMaxLen].
func (b *Builder) LabelFloat32(name string, value float32) *Builder {
	f := float64(value)
	return b.label(name, false, nil, nil, nil, nil, &f)
}

// LabelFloat64 appends a pair of label name and float64 label value to the Builder.
//
// NoOp if :
//   - no metric name has been set using [vimebu.Builder.Metric].
//   - the label name is empty or contains more than [vimebu.LabelNameMaxLen].
func (b *Builder) LabelFloat64(name string, value float64) *Builder {
	return b.label(name, false, nil, nil, nil, nil, &value)
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
	return b.label(name, false, &s, nil, nil, nil, nil)
}

// LabelStringerQuote appends a pair of label name and label value (implementing [fmt.Stringer]) to the Builder.
// Quotes inside the label value will be escaped.
//
// NoOp if :
//   - no metric name has been set using [vimebu.Builder.Metric].
//   - the label name is empty or contains more than [vimebu.LabelNameMaxLen].
func (b *Builder) LabelStringerQuote(name string, value fmt.Stringer) *Builder {
	s := value.String()
	return b.label(name, true, &s, nil, nil, nil, nil)
}

func (b *Builder) label(name string, escapeQuote bool, stringValue *string, boolValue *bool, uint64Value *uint64, int64Value *int64, float64Value *float64) *Builder {
	if !b.flName {
		log.Println("metric has not been called on this builder, skipping")
		return b
	}

	ln := len(name)
	if ln == 0 {
		log.Printf("metric: %q, label name must not be empty, skipping", b.buf)
		return b
	}
	if ln > LabelNameMaxLen {
		log.Printf("metric: %q, label name: %q, label name contains too many bytes, skipping", b.buf, name)
		return b
	}

	// String values need to be checked separately as they can be invalid (empty, or contain too many bytes).
	if stringValue != nil {
		lv := len(*stringValue)
		if lv == 0 {
			log.Printf("metric: %q, label name: %q, label value must not be empty, skipping", b.buf, name)
			return b
		}
		if lv > LabelValueLen {
			log.Printf("metric: %q, label name: %q, label value contains too many bytes, skipping", b.buf, name)
			return b
		}
	}

	if b.flLabel { // If we already wrote a label, start writing commas before label names.
		b.buf = append(b.buf, commaByte)
	} else { // Otherwise, mark flag as true for next pass.
		b.buf = append(b.buf, leftBracketByte)
		b.flLabel = true
	}

	b.buf = append(b.buf, name...)
	b.buf = append(b.buf, equalByte)
	switch {
	case stringValue != nil:
		b.buf = appendStringValue(b.buf, *stringValue, escapeQuote)
	case boolValue != nil:
		b.buf = appendBoolValue(b.buf, *boolValue)
	case uint64Value != nil:
		b.buf = appendUint64Value(b.buf, *uint64Value)
	case int64Value != nil:
		b.buf = appendInt64Value(b.buf, *int64Value)
	case float64Value != nil:
		b.buf = appendFloat64Value(b.buf, *float64Value)
	default: // Internal problem (wrong use of the label function), panic.
		panic("unsupported case - no label value set")
	}

	return b
}

// String builds the metric by returning the accumulated string.
func (b *Builder) String() string {
	defer putBuffer(b.buffer)
	if !b.flName {
		return ""
	}
	if b.flLabel {
		b.buf = append(b.buf, rightBracketByte)
	}

	return string(b.buf)
}
