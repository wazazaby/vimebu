package vimebu

import (
	"fmt"
	"strconv"
)

const (
	leftBracketByte  byte = '{'
	rightBracketByte byte = '}'
	commaByte        byte = ','
	equalByte        byte = '='
	doubleQuotesByte byte = '"'

	base10                 int  = 10
	floatFormattingVerb    byte = 'f'
	floatShortestPrecision int  = -1
	floatBitSize           int  = 64

	errorLabelName string = "error"
)

// Builder is used to efficiently build a VictoriaMetrics metric.
//
// It is forbidden copying [Builder] instances.
// [Builder] instances MUST not be used from concurrently running goroutines.
//
// The zero value is ready to use.
type Builder struct {
	_ noCopy

	buf []byte

	flName     bool
	flLabel    bool
	flAcquired bool
}

// BuilderOption
type BuilderOption func(*Builder)

// Reset zeroes out a [Builder] instance for reuse.
func (b *Builder) Reset() {
	b.buf = b.buf[:0]
	b.flName = false
	b.flLabel = false
	b.flAcquired = false
}

// Metric acquires and return a zeroed-out [Builder] instance from the
// [DefaultBuilderPool] and sets the metric's name.
func Metric(name string, options ...BuilderOption) *Builder {
	b := DefaultBuilderPool.Acquire()
	b.flAcquired = true
	return b.Metric(name, options...)
}

// Metric sets the metric's name of the [Builder].
//
// Panics if [Builder.Metric] was called previously on the same Builder instance
// without it being reset, or if the provided name is empty.
func (b *Builder) Metric(name string, options ...BuilderOption) *Builder {
	if len(name) == 0 {
		panic("vimebu: Builder.Metric has been passed an empty metric name")
	}
	if b.flName {
		panic("vimebu: Builder.Metric has already been called on this instance")
	}

	for _, applyOption := range options {
		applyOption(b)
	}

	b.buf = append(b.buf, name...)
	b.flName = true
	return b
}

// LabelString adds a label with a value of type string to the [Builder].
//
// NoOp if the label name or value are empty.
//
// Panics if [Builder.Metric] hasn't been called on this instance of the [Builder].
func (b *Builder) LabelString(name, value string) *Builder {
	if !b.flName {
		panic("vimebu: can't add a label to a Builder with no metric name")
	}
	if len(name) == 0 || len(value) == 0 {
		return b
	}
	b.buf = b.appendCommaOrLeftBracket()
	b.buf = append(b.buf, name...)
	b.buf = append(b.buf, equalByte)
	b.buf = append(b.buf, doubleQuotesByte)
	b.buf = append(b.buf, value...)
	b.buf = append(b.buf, doubleQuotesByte)
	return b
}

// LabelStringQuote adds a label with a value of type string to the [Builder].
// Quotes inside label value will be escaped using [strconv.AppendQuote].
//
// NoOp if the label name or value are empty.
//
// Panics if [Builder.Metric] hasn't been called on this instance of the [Builder].
func (b *Builder) LabelStringQuote(name, value string) *Builder {
	if !b.flName {
		panic("vimebu: can't add a label to a Builder with no metric name")
	}
	if len(name) == 0 || len(value) == 0 {
		return b
	}
	b.buf = b.appendCommaOrLeftBracket()
	b.buf = append(b.buf, name...)
	b.buf = append(b.buf, equalByte)
	b.buf = strconv.AppendQuote(b.buf, value)
	return b
}

// LabelError adds a label with a value implementing the error interface to the [Builder].
//
// NoOp if the label name is empty, or if err is nil.
//
// Panics if [Builder.Metric] hasn't been called on this instance of the [Builder].
func (b *Builder) LabelError(err error) *Builder {
	if err == nil {
		return b
	}
	return b.LabelString(errorLabelName, err.Error())
}

// LabelNamedError adds a label with a value implementing the error interface to the [Builder].
//
// NoOp if the label name is empty, or if err is nil.
//
// Panics if [Builder.Metric] hasn't been called on this instance of the [Builder].
func (b *Builder) LabelNamedError(name string, err error) *Builder {
	if err == nil {
		return b
	}
	return b.LabelString(name, err.Error())
}

// LabelErrorQuote adds a label with a value implementing the error interface to the [Builder].
// Quotes inside label value will be escaped using [strconv.AppendQuote].
//
// NoOp if the label name is empty, or if err is nil.
//
// Panics if [Builder.Metric] hasn't been called on this instance of the [Builder].
func (b *Builder) LabelErrorQuote(err error) *Builder {
	if err == nil {
		return b
	}
	return b.LabelStringQuote(errorLabelName, err.Error())
}

// LabelNamedErrorQuote adds a label with a value implementing the error interface to the [Builder].
// Quotes inside label value will be escaped using [strconv.AppendQuote].
//
// NoOp if the label name is empty, or if err is nil.
//
// Panics if [Builder.Metric] hasn't been called on this instance of the [Builder].
func (b *Builder) LabelNamedErrorQuote(name string, err error) *Builder {
	if err == nil {
		return b
	}
	return b.LabelStringQuote(name, err.Error())
}

// LabelBool adds a label with a value of type bool to the [Builder].
//
// NoOp if the label name is empty.
//
// Panics if [Builder.Metric] hasn't been called on this instance of the [Builder].
func (b *Builder) LabelBool(name string, value bool) *Builder {
	if value {
		return b.LabelString(name, "true")
	}
	return b.LabelString(name, "false")
}

// LabelUint adds a label with a value of type uint to the [Builder].
//
// NoOp if the label name is empty.
//
// Panics if [Builder.Metric] hasn't been called on this instance of the [Builder].
func (b *Builder) LabelUint(name string, value uint) *Builder {
	return b.LabelUint64(name, uint64(value))
}

// LabelUint8 adds a label with a value of type uint8 to the [Builder].
//
// NoOp if the label name is empty.
//
// Panics if [Builder.Metric] hasn't been called on this instance of the [Builder].
func (b *Builder) LabelUint8(name string, value uint8) *Builder {
	return b.LabelUint64(name, uint64(value))
}

// LabelUint16 adds a label with a value of type uint16 to the [Builder].
//
// NoOp if the label name is empty.
//
// Panics if [Builder.Metric] hasn't been called on this instance of the [Builder].
func (b *Builder) LabelUint16(name string, value uint16) *Builder {
	return b.LabelUint64(name, uint64(value))
}

// LabelUint32 adds a label with a value of type uint32 to the [Builder].
//
// NoOp if the label name is empty.
//
// Panics if [Builder.Metric] hasn't been called on this instance of the [Builder].
func (b *Builder) LabelUint32(name string, value uint32) *Builder {
	return b.LabelUint64(name, uint64(value))
}

// LabelUint64 adds a label with a value of type uint64 to the [Builder].
//
// NoOp if the label name is empty.
//
// Panics if [Builder.Metric] hasn't been called on this instance of the [Builder].
func (b *Builder) LabelUint64(name string, value uint64) *Builder {
	if !b.flName {
		panic("vimebu: can't add a label to a Builder with no metric name")
	}
	if len(name) == 0 {
		return b
	}
	b.buf = b.appendCommaOrLeftBracket()
	b.buf = append(b.buf, name...)
	b.buf = append(b.buf, equalByte)
	b.buf = append(b.buf, doubleQuotesByte)
	b.buf = strconv.AppendUint(b.buf, value, base10)
	b.buf = append(b.buf, doubleQuotesByte)
	return b
}

// LabelInt adds a label with a value of type int to the [Builder].
//
// NoOp if the label name is empty.
//
// Panics if [Builder.Metric] hasn't been called on this instance of the [Builder].
func (b *Builder) LabelInt(name string, value int) *Builder {
	return b.LabelInt64(name, int64(value))
}

// LabelInt8 adds a label with a value of type int8 to the [Builder].
//
// NoOp if the label name is empty.
//
// Panics if [Builder.Metric] hasn't been called on this instance of the [Builder].
func (b *Builder) LabelInt8(name string, value int8) *Builder {
	return b.LabelInt64(name, int64(value))
}

// LabelInt16 adds a label with a value of type int16 to the [Builder].
//
// NoOp if the label name is empty.
//
// Panics if [Builder.Metric] hasn't been called on this instance of the [Builder].
func (b *Builder) LabelInt16(name string, value int16) *Builder {
	return b.LabelInt64(name, int64(value))
}

// LabelInt32 adds a label with a value of type int32 to the [Builder].
//
// NoOp if the label name is empty.
//
// Panics if [Builder.Metric] hasn't been called on this instance of the [Builder].
func (b *Builder) LabelInt32(name string, value int32) *Builder {
	return b.LabelInt64(name, int64(value))
}

// LabelInt64 adds a label with a value of type int64 to the [Builder].
//
// NoOp if the label name is empty.
//
// Panics if [Builder.Metric] hasn't been called on this instance of the [Builder].
func (b *Builder) LabelInt64(name string, value int64) *Builder {
	if !b.flName {
		panic("vimebu: can't add a label to a Builder with no metric name")
	}
	if len(name) == 0 {
		return b
	}
	b.buf = b.appendCommaOrLeftBracket()
	b.buf = append(b.buf, name...)
	b.buf = append(b.buf, equalByte)
	b.buf = append(b.buf, doubleQuotesByte)
	b.buf = strconv.AppendInt(b.buf, value, base10)
	b.buf = append(b.buf, doubleQuotesByte)
	return b
}

// LabelFloat32 adds a label with a value of type float32 to the [Builder].
//
// NoOp if the label name is empty.
//
// Panics if [Builder.Metric] hasn't been called on this instance of the [Builder].
func (b *Builder) LabelFloat32(name string, value float32) *Builder {
	return b.LabelFloat64(name, float64(value))
}

// LabelFloat64 adds a label with a value of type float64 to the [Builder].
//
// NoOp if the label name is empty.
//
// Panics if [Builder.Metric] hasn't been called on this instance of the [Builder].
func (b *Builder) LabelFloat64(name string, value float64) *Builder {
	if !b.flName {
		panic("vimebu: can't add a label to a Builder with no metric name")
	}
	if len(name) == 0 {
		return b
	}
	b.buf = b.appendCommaOrLeftBracket()
	b.buf = append(b.buf, name...)
	b.buf = append(b.buf, equalByte)
	b.buf = append(b.buf, doubleQuotesByte)
	b.buf = strconv.AppendFloat(b.buf, value, floatFormattingVerb, floatShortestPrecision, floatBitSize)
	b.buf = append(b.buf, doubleQuotesByte)
	return b
}

// LabelStringer adds a label with a value implementing the [fmt.Stringer] interface to the [Builder].
//
// NoOp if the label name is empty, if value is nil, or if the value.String() method call returns an empty string.
//
// Panics if [Builder.Metric] hasn't been called on this instance of the [Builder].
func (b *Builder) LabelStringer(name string, value fmt.Stringer) *Builder {
	if value == nil {
		return b
	}
	return b.LabelString(name, value.String())
}

// LabelStringerQuote adds a label with a value implementing the [fmt.Stringer] interface to the [Builder].
// Quotes inside label value will be escaped using [strconv.AppendQuote].
//
// NoOp if the label name is empty, if value is nil, or if the value.String() method call returns an empty string.
//
// Panics if [Builder.Metric] hasn't been called on this instance of the [Builder].
func (b *Builder) LabelStringerQuote(name string, value fmt.Stringer) *Builder {
	if value == nil {
		return b
	}
	return b.LabelStringQuote(name, value.String())
}

// String builds the complete metric by returning the accumulated string.
func (b *Builder) String() string {
	if b.flAcquired {
		defer DefaultBuilderPool.Release(b)
	}
	if !b.flName {
		return ""
	}
	if b.flLabel {
		b.buf = append(b.buf, rightBracketByte)
	}
	return string(b.buf)
}

func (b *Builder) appendCommaOrLeftBracket() []byte {
	if b.flLabel {
		return append(b.buf, commaByte)
	}
	b.flLabel = true
	return append(b.buf, leftBracketByte)
}
