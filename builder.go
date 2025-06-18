package vimebu

import (
	"fmt"
	"log"
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

// BuilderOption represents a modifier function that will apply a specific
// configuration to a [Builder] instance.
type BuilderOption func(*Builder)

// WithLabelNameMaxLen sets the max authorized length for a label name.
//
// Zero means no length limit.
//
// If the max len is exceeded for a label name, a log line containing the
// reason will be written to [os.Stderr], and the label will be skipped.
func WithLabelNameMaxLen(maxLen int) BuilderOption {
	return func(b *Builder) {
		b.labelNameMaxLen = maxLen
	}
}

// WithLabelValueMaxLen sets the max authorized length for a label value.
//
// Only applies to label values added using the following methods :
//
//   - [Builder.LabelString]
//   - [Builder.LabelTrustedString]
//   - [Builder.LabelError]
//   - [Builder.LabelTrustedError]
//   - [Builder.LabelNamedError]
//   - [Builder.LabelNamedTrustedError]
//
// Zero means no length limit.
//
// If the max len is exceeded for a label value, a log line containing the
// reason will be written to [os.Stderr], and the label will be skipped.
func WithLabelValueMaxLen(maxLen int) BuilderOption {
	return func(b *Builder) {
		b.labelValueMaxLen = maxLen
	}
}

// Builder is used to efficiently build a VictoriaMetrics metric.
//
// It is forbidden copying [Builder] instances.
// [Builder] instances MUST not be used from concurrently running goroutines.
//
// The zero value is ready to use.
//
// When validating label names and values, [Builder] instances will write log lines
// to [os.Stderr] using the [log.Printf] function (standard logger).
//
// If you wish to redirect these log lines to your own logger, you can do this :
//   - Logrus : [log.SetOutput]([logrus.Logger.Writer])
//   - Zap : [zap.RedirectStdLog]([zap.Logger])
type Builder struct {
	_ noCopy

	pool *BuilderPool

	buf []byte

	hasMetricName bool
	hasLabel      bool

	labelNameMaxLen  int
	labelValueMaxLen int
}

// Reset zeroes out a [Builder] instance for reuse.
func (b *Builder) Reset() {
	b.pool = nil
	b.buf = b.buf[:0]
	b.hasMetricName = false
	b.hasLabel = false
	b.labelNameMaxLen = 0
	b.labelValueMaxLen = 0
}

// Metric acquires and returns a zeroed-out [Builder] instance from the
// default builder pool and sets the metric's name.
func Metric(name string, options ...BuilderOption) *Builder {
	return defaultBuilderPool.Metric(name, options...)
}

// Metric sets the metric's name of the [Builder].
//
// Panics if [Builder.Metric] was called previously on the same Builder instance
// without it being reset, or if the provided name is empty.
func (b *Builder) Metric(name string, options ...BuilderOption) *Builder {
	if len(name) == 0 {
		panic("vimebu: Builder.Metric has been passed an empty metric name")
	}
	if b.hasMetricName {
		panic("vimebu: Builder.Metric has already been called on this instance")
	}

	for _, applyOption := range options {
		applyOption(b)
	}

	b.buf = append(b.buf, name...)
	b.hasMetricName = true
	return b
}

// LabelString adds a label with a value of type string to the [Builder].
// Quotes inside label value will be escaped using [strconv.AppendQuote].
//
// NoOp if the label name or value are empty.
//
// Panics if [Builder.Metric] hasn't been called on this instance of the [Builder].
func (b *Builder) LabelString(name, value string) *Builder {
	return b.labelString(name, value, true)
}

// LabelTrustedString adds a label with a value of type string to the [Builder].
//
// Using this method over the standard one is more efficient as it avoid unecessary compute & allocations,
// the drawback being that your label needs to be safe / escaped already (otherwise, VM will panic).
//
// NoOp if the label name or value are empty.
//
// Panics if [Builder.Metric] hasn't been called on this instance of the [Builder].
func (b *Builder) LabelTrustedString(name, value string) *Builder {
	return b.labelString(name, value, false)
}

func (b *Builder) labelString(name, value string, escapeQuotes bool) *Builder {
	if !b.hasMetricName {
		panic("vimebu: can't add a label to a Builder with no metric name")
	}
	if !b.isValidLabelName(name) {
		return b
	}
	if !b.isValidLabelValue(name, value) {
		return b
	}
	b.buf = b.appendCommaOrLeftBracket()
	b.buf = append(b.buf, name...)
	b.buf = append(b.buf, equalByte)
	if escapeQuotes {
		b.buf = strconv.AppendQuote(b.buf, value)
	} else { // Fast path for when explicit quote escaping is not required.
		b.buf = append(b.buf, doubleQuotesByte)
		b.buf = append(b.buf, value...)
		b.buf = append(b.buf, doubleQuotesByte)
	}
	return b
}

// LabelError adds a label with a value implementing the error interface to the [Builder].
// Quotes inside label value will be escaped using [strconv.AppendQuote].
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
// Quotes inside label value will be escaped using [strconv.AppendQuote].
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

// LabelTrustedError adds a label with a value implementing the error interface to the [Builder].
//
// Using this method over the standard one is more efficient as it avoid unecessary compute & allocations,
// the drawback being that your label needs to be safe / escaped already (otherwise, VM will panic).
//
// NoOp if the label name is empty, or if err is nil.
//
// Panics if [Builder.Metric] hasn't been called on this instance of the [Builder].
func (b *Builder) LabelTrustedError(err error) *Builder {
	if err == nil {
		return b
	}
	return b.LabelTrustedString(errorLabelName, err.Error())
}

// LabelNamedTrustedError adds a label with a value implementing the error interface to the [Builder].
//
// Using this method over the standard one is more efficient as it avoid unecessary compute & allocations,
// the drawback being that your label needs to be safe / escaped already (otherwise, VM will panic).
//
// NoOp if the label name is empty, or if err is nil.
//
// Panics if [Builder.Metric] hasn't been called on this instance of the [Builder].
func (b *Builder) LabelNamedTrustedError(name string, err error) *Builder {
	if err == nil {
		return b
	}
	return b.LabelTrustedString(name, err.Error())
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
	if !b.hasMetricName {
		panic("vimebu: can't add a label to a Builder with no metric name")
	}
	if !b.isValidLabelName(name) {
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
	if !b.hasMetricName {
		panic("vimebu: can't add a label to a Builder with no metric name")
	}
	if !b.isValidLabelName(name) {
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
	if !b.hasMetricName {
		panic("vimebu: can't add a label to a Builder with no metric name")
	}
	if !b.isValidLabelName(name) {
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
// Quotes inside label value will be escaped using [strconv.AppendQuote].
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

// LabelTrustedStringer adds a label with a value implementing the [fmt.Stringer] interface to the [Builder].
//
// Using this method over the standard one is more efficient as it avoid unecessary compute & allocations,
// the drawback being that your label needs to be safe / escaped already (otherwise, VM will panic).
//
// NoOp if the label name is empty, if value is nil, or if the value.String() method call returns an empty string.
//
// Panics if [Builder.Metric] hasn't been called on this instance of the [Builder].
func (b *Builder) LabelTrustedStringer(name string, value fmt.Stringer) *Builder {
	if value == nil {
		return b
	}
	return b.LabelTrustedString(name, value.String())
}

// String builds the complete metric by returning the accumulated string.
func (b *Builder) String() string {
	if b.pool != nil {
		defer b.pool.Release(b)
	}
	if !b.hasMetricName {
		return ""
	}
	if b.hasLabel {
		b.buf = append(b.buf, rightBracketByte)
	}
	return string(b.buf)
}

// isValidLabelName checks if the provided label name is valid.
//
// For it to be valid, it's len must be greater than 0.
//
// If the [Builder] was passed the [WithLabelNameMaxLen] option, the
// label name len must also be less than the provided max len value.
//
// In case of an invalid label name, a log line containing the reasons will be written to [os.Stderr].
func (b *Builder) isValidLabelName(name string) bool {
	ln := len(name)
	if ln == 0 {
		log.Printf("vimebu: metric %q, empty label name - skipping", b.buf)
		return false
	}
	if b.labelNameMaxLen > 0 && ln > b.labelNameMaxLen {
		log.Printf("vimebu: metric %q, label name %q len exceeds set limit of %d - skipping", b.buf, name, b.labelNameMaxLen)
		return false
	}
	return true
}

// isValidLabelValue checks if the provided label value is valid.
//
// For it to be valid, it's len must be greater than 0.
//
// If the [Builder] was passed the [WithLabelValueMaxLen] option, the
// label value len must also be less than the provided max len value.
//
// In case of an invalid label value, a log line containing the reasons will be written to [os.Stderr].
func (b *Builder) isValidLabelValue(name, value string) bool {
	lv := len(value)
	if lv == 0 {
		log.Printf("vimebu: metric %q, label name: %q, received empty label value - skipping", b.buf, name)
		return false
	}
	if b.labelValueMaxLen > 0 && lv > b.labelValueMaxLen {
		log.Printf("vimebu: metric %q, label name %q, label value %q len exceeds set limit of %d - skipping", b.buf, name, value, b.labelNameMaxLen)
		return false
	}
	return true
}

func (b *Builder) appendCommaOrLeftBracket() []byte {
	if b.hasLabel {
		return append(b.buf, commaByte)
	}
	b.hasLabel = true
	return append(b.buf, leftBracketByte)
}
