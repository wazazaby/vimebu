package vimebu

import (
	"fmt"
	"log"
	"strconv"
	"sync"
)

const errorLabelName = "error"

// noCopy should be embedded into a struct which mustn't be copied,
// so `go vet` gives a warning if this struct is copied.
//
// See https://github.com/golang/go/issues/8005#issuecomment-190753527 for details.
type noCopy [0]sync.Mutex

// BuilderOption represents a modifier function that will apply a specific
// configuration to a [Builder] instance.
type BuilderOption func(*Builder)

// WithLabelNameMaxLen sets the max authorized length for a label name.
// Zero means no limit. Labels with a longer name are skipped.
func WithLabelNameMaxLen(maxLen int) BuilderOption {
	return func(b *Builder) {
		b.labelNameMaxLen = maxLen
	}
}

// WithLabelValueMaxLen sets the max authorized length for a label value.
// Zero means no limit. Labels with a longer value are skipped.
//
// Only applies to string, error and [fmt.Stringer] values - numeric and bool
// labels are never skipped on length.
func WithLabelValueMaxLen(maxLen int) BuilderOption {
	return func(b *Builder) {
		b.labelValueMaxLen = maxLen
	}
}

// Builder is used to efficiently build a VictoriaMetrics metric.
//
// The zero value is ready to use. [Builder] instances must not be copied, nor
// used from concurrently running goroutines.
//
// Every Label* method is a NoOp when the label name is empty, and panics when
// [Builder.Metric] hasn't been called on the instance yet. Skipped labels are
// reported with a [log.Printf] line on the standard logger, which you can
// redirect using [log.SetOutput] :
//   - Logrus : [log.SetOutput]([logrus.Logger.Writer])
//   - Zap : [zap.RedirectStdLog]([zap.Logger])
type Builder struct {
	_ noCopy

	pool *BuilderPool

	buf []byte

	labelNameMaxLen  int
	labelValueMaxLen int

	hasLabel bool
}

// Reset zeroes out a [Builder] instance for reuse.
func (b *Builder) Reset() {
	b.pool = nil
	b.buf = b.buf[:0]
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
	if len(b.buf) > 0 {
		panic("vimebu: Builder.Metric has already been called on this instance")
	}

	for _, applyOption := range options {
		applyOption(b)
	}

	b.buf = append(b.buf, name...)
	return b
}

// LabelString adds a label with a value of type string to the [Builder].
//
// NoOp if the label value is empty.
func (b *Builder) LabelString(name, value string) *Builder {
	return b.labelString(name, value, false)
}

// LabelStringQuote adds a label with a value of type string to the [Builder],
// escaping backslashes, double quotes and newlines inside the value.
//
// NoOp if the label value is empty.
func (b *Builder) LabelStringQuote(name, value string) *Builder {
	return b.labelString(name, value, true)
}

func (b *Builder) labelString(name, value string, escape bool) *Builder {
	if !b.validLabelName(name) {
		return b
	}
	if lv := len(value); lv == 0 || (b.labelValueMaxLen > 0 && lv > b.labelValueMaxLen) {
		b.logSkippedLabelValue(name, value)
		return b
	}
	b.openLabel(name)
	if escape {
		b.buf = appendEscaped(b.buf, value)
	} else {
		b.buf = append(b.buf, value...)
	}
	b.buf = append(b.buf, '"')
	return b
}

// LabelError adds an "error" label holding err's message to the [Builder].
//
// NoOp if err is nil.
func (b *Builder) LabelError(err error) *Builder {
	if err == nil {
		return b
	}
	return b.LabelString(errorLabelName, err.Error())
}

// LabelNamedError adds a label holding err's message to the [Builder].
//
// NoOp if err is nil.
func (b *Builder) LabelNamedError(name string, err error) *Builder {
	if err == nil {
		return b
	}
	return b.LabelString(name, err.Error())
}

// LabelErrorQuote adds an "error" label holding err's message to the [Builder],
// escaping backslashes, double quotes and newlines inside the message.
//
// NoOp if err is nil.
func (b *Builder) LabelErrorQuote(err error) *Builder {
	if err == nil {
		return b
	}
	return b.LabelStringQuote(errorLabelName, err.Error())
}

// LabelNamedErrorQuote adds a label holding err's message to the [Builder],
// escaping backslashes, double quotes and newlines inside the message.
//
// NoOp if err is nil.
func (b *Builder) LabelNamedErrorQuote(name string, err error) *Builder {
	if err == nil {
		return b
	}
	return b.LabelStringQuote(name, err.Error())
}

// LabelStringer adds a label with a value implementing [fmt.Stringer] to the [Builder].
//
// NoOp if value is nil or if value.String() returns an empty string.
func (b *Builder) LabelStringer(name string, value fmt.Stringer) *Builder {
	if value == nil {
		return b
	}
	return b.LabelString(name, value.String())
}

// LabelStringerQuote adds a label with a value implementing [fmt.Stringer] to the
// [Builder], escaping backslashes, double quotes and newlines inside the value.
//
// NoOp if value is nil or if value.String() returns an empty string.
func (b *Builder) LabelStringerQuote(name string, value fmt.Stringer) *Builder {
	if value == nil {
		return b
	}
	return b.LabelStringQuote(name, value.String())
}

// LabelBool adds a label with a value of type bool to the [Builder].
func (b *Builder) LabelBool(name string, value bool) *Builder {
	if !b.validLabelName(name) {
		return b
	}
	b.openLabel(name)
	b.buf = strconv.AppendBool(b.buf, value)
	b.buf = append(b.buf, '"')
	return b
}

// LabelUint adds a label with a value of type uint to the [Builder].
func (b *Builder) LabelUint(name string, value uint) *Builder {
	return b.LabelUint64(name, uint64(value))
}

// LabelUint8 adds a label with a value of type uint8 to the [Builder].
func (b *Builder) LabelUint8(name string, value uint8) *Builder {
	return b.LabelUint64(name, uint64(value))
}

// LabelUint16 adds a label with a value of type uint16 to the [Builder].
func (b *Builder) LabelUint16(name string, value uint16) *Builder {
	return b.LabelUint64(name, uint64(value))
}

// LabelUint32 adds a label with a value of type uint32 to the [Builder].
func (b *Builder) LabelUint32(name string, value uint32) *Builder {
	return b.LabelUint64(name, uint64(value))
}

// LabelUint64 adds a label with a value of type uint64 to the [Builder].
func (b *Builder) LabelUint64(name string, value uint64) *Builder {
	if !b.validLabelName(name) {
		return b
	}
	b.openLabel(name)
	b.buf = strconv.AppendUint(b.buf, value, 10)
	b.buf = append(b.buf, '"')
	return b
}

// LabelInt adds a label with a value of type int to the [Builder].
func (b *Builder) LabelInt(name string, value int) *Builder {
	return b.LabelInt64(name, int64(value))
}

// LabelInt8 adds a label with a value of type int8 to the [Builder].
func (b *Builder) LabelInt8(name string, value int8) *Builder {
	return b.LabelInt64(name, int64(value))
}

// LabelInt16 adds a label with a value of type int16 to the [Builder].
func (b *Builder) LabelInt16(name string, value int16) *Builder {
	return b.LabelInt64(name, int64(value))
}

// LabelInt32 adds a label with a value of type int32 to the [Builder].
func (b *Builder) LabelInt32(name string, value int32) *Builder {
	return b.LabelInt64(name, int64(value))
}

// LabelInt64 adds a label with a value of type int64 to the [Builder].
func (b *Builder) LabelInt64(name string, value int64) *Builder {
	if !b.validLabelName(name) {
		return b
	}
	b.openLabel(name)
	b.buf = strconv.AppendInt(b.buf, value, 10)
	b.buf = append(b.buf, '"')
	return b
}

// LabelFloat32 adds a label with a value of type float32 to the [Builder].
func (b *Builder) LabelFloat32(name string, value float32) *Builder {
	return b.LabelFloat64(name, float64(value))
}

// LabelFloat64 adds a label with a value of type float64 to the [Builder].
func (b *Builder) LabelFloat64(name string, value float64) *Builder {
	if !b.validLabelName(name) {
		return b
	}
	b.openLabel(name)
	b.buf = strconv.AppendFloat(b.buf, value, 'f', -1, 64)
	b.buf = append(b.buf, '"')
	return b
}

// String builds the complete metric by returning the accumulated string.
//
// [Builder] instances obtained from a pool - via [Metric] or [BuilderPool.Metric] -
// are reset and released back to it.
func (b *Builder) String() string {
	if b.hasLabel {
		b.buf = append(b.buf, '}')
	}
	s := string(b.buf)
	if b.pool != nil {
		b.pool.Release(b)
	}
	return s
}

// openLabel appends the label separator, the name, the equal sign and the
// opening double quote. Callers append the value, then the closing quote.
func (b *Builder) openLabel(name string) {
	if b.hasLabel {
		b.buf = append(b.buf, ',')
	} else {
		b.buf = append(b.buf, '{')
		b.hasLabel = true
	}
	b.buf = append(b.buf, name...)
	b.buf = append(b.buf, '=', '"')
}

// validLabelName reports whether name can be used as a label name.
//
// Panics if the [Builder] has no metric name yet.
func (b *Builder) validLabelName(name string) bool {
	if ln := len(name); len(b.buf) == 0 || ln == 0 || (b.labelNameMaxLen > 0 && ln > b.labelNameMaxLen) {
		return b.rejectLabelName(name)
	}
	return true
}

// rejectLabelName is the cold path of [Builder.validLabelName], kept out of line so
// the checks themselves stay cheap enough to inline. It always returns false.
func (b *Builder) rejectLabelName(name string) bool {
	switch {
	case len(b.buf) == 0:
		panic("vimebu: can't add a label to a Builder with no metric name")
	case len(name) == 0:
		log.Printf("vimebu: metric %q, empty label name - skipping", b.buf)
	default:
		log.Printf("vimebu: metric %q, label name %q len exceeds set limit of %d - skipping", b.buf, name, b.labelNameMaxLen)
	}
	return false
}

// logSkippedLabelValue is the cold path of the label value checks done by
// [Builder.labelString], kept out of line so those checks stay cheap.
func (b *Builder) logSkippedLabelValue(name, value string) {
	if len(value) == 0 {
		log.Printf("vimebu: metric %q, label name: %q, received empty label value - skipping", b.buf, name)
		return
	}
	log.Printf("vimebu: metric %q, label name %q, label value %q len exceeds set limit of %d - skipping", b.buf, name, value, b.labelValueMaxLen)
}

// appendEscaped appends value, escaping the only characters the exposition format
// needs escaped inside a label value: backslash, double quote and newline.
//
// Runs between two escapes are copied in bulk, so a value needing no escaping at
// all - the common case - costs a single append.
func appendEscaped(dst []byte, value string) []byte {
	last := 0
	for i := range len(value) {
		var escaped byte
		switch value[i] {
		case '\\', '"':
			escaped = value[i]
		case '\n':
			escaped = 'n'
		default:
			continue
		}
		dst = append(dst, value[last:i]...)
		dst = append(dst, '\\', escaped)
		last = i + 1
	}
	return append(dst, value[last:]...)
}
