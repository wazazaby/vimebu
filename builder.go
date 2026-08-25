package vimebu

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
)

const errorLabelName = "error"

// escapeScanMinLen is the value length at which the IndexByte scans in [appendEscaped]
// start beating a plain byte loop. Empirically tuned on arm64: below it, values tend to
// carry several escapes and the three scans don't pay for themselves, so lowering this
// regresses short quoted values.
const escapeScanMinLen = 48

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
// Only applies to string, error and [fmt.Stringer] values - numeric and bool labels are
// never skipped on length. Note that this leaves floats uncapped: [Builder.LabelFloat64]
// formats without an exponent, so a value like 1e300 emits over 300 bytes regardless of
// this option.
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
// Label* methods panic if [Builder.Metric] hasn't been called yet, except the error and
// [fmt.Stringer] variants, which return early on a nil value. Labels are skipped when the
// name is empty or longer than [WithLabelNameMaxLen], and when a string, error or
// [fmt.Stringer] value is empty or longer than [WithLabelValueMaxLen]. Every skip is
// reported with a [log.Printf] line on the standard logger, which you can redirect using
// [log.SetOutput] :
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
	if b.hasMetricName() {
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

// LabelStringQuote adds a label with a value of type string to the [Builder], escaping
// backslashes, double quotes and newlines inside the value.
//
// NoOp if the label value is empty.
func (b *Builder) LabelStringQuote(name, value string) *Builder {
	return b.labelString(name, value, true)
}

func (b *Builder) labelString(name, value string, escape bool) *Builder {
	if !b.validLabelName(name) {
		return b.skipLabelName(name)
	}
	if lv := len(value); lv == 0 || (b.labelValueMaxLen > 0 && lv > b.labelValueMaxLen) {
		return b.skipLabelValue(name, value)
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

// LabelErrorQuote adds an "error" label holding err's message to the [Builder], escaped
// as [Builder.LabelStringQuote] does.
//
// NoOp if err is nil.
func (b *Builder) LabelErrorQuote(err error) *Builder {
	if err == nil {
		return b
	}
	return b.LabelStringQuote(errorLabelName, err.Error())
}

// LabelNamedErrorQuote adds a label holding err's message to the [Builder], escaped as
// [Builder.LabelStringQuote] does.
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
// [Builder], escaped as [Builder.LabelStringQuote] does.
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
		return b.skipLabelName(name)
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
		return b.skipLabelName(name)
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
		return b.skipLabelName(name)
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
		return b.skipLabelName(name)
	}
	b.openLabel(name)
	b.buf = strconv.AppendFloat(b.buf, value, 'f', -1, 64)
	b.buf = append(b.buf, '"')
	return b
}

// String builds the complete metric by returning the accumulated string.
//
// It CONSUMES [Builder] instances that came from a pool - via [Metric] or
// [BuilderPool.Metric] - resetting them and releasing them back to it. Since [Builder]
// implements [fmt.Stringer], that also happens when a Builder is handed to any fmt verb,
// so don't format a Builder you still intend to use.
func (b *Builder) String() string {
	if b.hasLabel {
		b.buf = append(b.buf, '}')
		b.hasLabel = false
	}
	s := string(b.buf)
	if b.pool != nil {
		b.pool.Release(b)
	}
	return s
}

// hasMetricName reports whether [Builder.Metric] has been called on this instance.
//
// It holds because Metric panics on an empty name and is the only thing that writes to
// the buffer before a label - anything that writes earlier must maintain this.
func (b *Builder) hasMetricName() bool {
	return len(b.buf) > 0
}

// openLabel starts a label: callers must append the value, then the closing double quote.
func (b *Builder) openLabel(name string) {
	sep := byte('{')
	if b.hasLabel {
		sep = ','
	}
	b.hasLabel = true
	// Chained so the slice header is stored back once instead of after every append.
	b.buf = append(append(append(b.buf, sep), name...), '=', '"')
}

// validLabelName reports whether name can be used as a label name.
//
// Deliberately free of any call - one would cost 57 of the inliner's 80-point budget and
// stop this from being inlined into every Label* method. Rejections go to
// [Builder.skipLabelName].
func (b *Builder) validLabelName(name string) bool {
	ln := len(name)
	return ln > 0 && b.hasMetricName() && (b.labelNameMaxLen == 0 || ln <= b.labelNameMaxLen)
}

// skipLabelName is the cold path of [Builder.validLabelName]. Its branches mirror that
// predicate's conditions in the same order - keep them in sync, or a rejection gets
// reported with the wrong reason.
func (b *Builder) skipLabelName(name string) *Builder {
	switch {
	case !b.hasMetricName():
		panic("vimebu: can't add a label to a Builder with no metric name")
	case len(name) == 0:
		log.Printf("vimebu: metric %q, empty label name - skipping", b.buf)
	default:
		log.Printf("vimebu: metric %q, label name %q len exceeds set limit of %d - skipping", b.buf, name, b.labelNameMaxLen)
	}
	return b
}

// skipLabelValue is the cold path of the label value checks in [Builder.labelString].
func (b *Builder) skipLabelValue(name, value string) *Builder {
	if len(value) == 0 {
		log.Printf("vimebu: metric %q, label name %q, received empty label value - skipping", b.buf, name)
		return b
	}
	log.Printf("vimebu: metric %q, label name %q, label value %q len exceeds set limit of %d - skipping", b.buf, name, value, b.labelValueMaxLen)
	return b
}

// appendEscaped appends value, escaping the only three characters the exposition format
// needs escaped inside a label value: backslash, double quote and newline.
//
// Deliberately not [strconv.AppendQuote]: that writes its own surrounding quotes, and it
// expands control bytes and invalid UTF-8 into \x / \u escapes that VictoriaMetrics never
// decodes. Everything outside those three characters is passed through untouched.
//
// Runs between escapes are copied in bulk, so a value needing no escaping costs a single
// append. Above [escapeScanMinLen] the IndexByte scans locate the first escape - or prove
// there is none - faster than the byte loop can.
func appendEscaped(dst []byte, value string) []byte {
	if len(value) >= escapeScanMinLen {
		i := strings.IndexByte(value, '"')
		if j := strings.IndexByte(value, '\\'); j >= 0 && (i < 0 || j < i) {
			i = j
		}
		if j := strings.IndexByte(value, '\n'); j >= 0 && (i < 0 || j < i) {
			i = j
		}
		if i < 0 {
			return append(dst, value...)
		}
		dst = append(dst, value[:i]...)
		value = value[i:]
	}

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
