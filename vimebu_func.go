package vimebu

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"unsafe"

	"golang.org/x/exp/constraints"
)

type bytesBuffer struct {
	b []byte
}

var ownBytesBufferPool = sync.Pool{
	New: func() any {
		return &bytesBuffer{
			b: make([]byte, 0, 64),
		}
	},
}

func getBytesBuffer() *bytesBuffer {
	return ownBytesBufferPool.Get().(*bytesBuffer)
}

func putBytesBuffer(b *bytesBuffer) {
	b.b = b.b[:0]
	ownBytesBufferPool.Put(b)
}

type Label func() (string, bool, *string, *bool, *uint64, *int64, *float64)

func LabelString(name, value string) Label {
	return func() (string, bool, *string, *bool, *uint64, *int64, *float64) {
		return name, false, &value, nil, nil, nil, nil
	}
}

func LabelStringCond(name, value string, predicate func() bool) Label {
	if predicate() {
		return LabelString(name, value)
	}
	return nil
}

func LabelStringQuote(name, value string) Label {
	return func() (string, bool, *string, *bool, *uint64, *int64, *float64) {
		return name, true, &value, nil, nil, nil, nil
	}
}

func LabelStringQuoteCond(name, value string, predicate func() bool) Label {
	if predicate() {
		return LabelStringQuote(name, value)
	}
	return nil
}

func LabelStringer(name string, value fmt.Stringer) Label {
	s := value.String()
	return func() (string, bool, *string, *bool, *uint64, *int64, *float64) {
		return name, false, &s, nil, nil, nil, nil
	}
}

func LabelStringerCond(name string, value fmt.Stringer, predicate func() bool) Label {
	if predicate() {
		return LabelStringer(name, value)
	}
	return nil
}

func LabelStringerQuote(name string, value fmt.Stringer) Label {
	s := value.String()
	return func() (string, bool, *string, *bool, *uint64, *int64, *float64) {
		return name, true, &s, nil, nil, nil, nil
	}
}

func LabelStringerQuoteCond(name string, value fmt.Stringer, predicate func() bool) Label {
	if predicate() {
		return LabelStringerQuote(name, value)
	}
	return nil
}

func LabelBool(name string, value bool) Label {
	return func() (string, bool, *string, *bool, *uint64, *int64, *float64) {
		return name, true, nil, &value, nil, nil, nil
	}
}

func LabelBoolCond(name string, value bool, predicate func() bool) Label {
	if predicate() {
		return LabelBool(name, value)
	}
	return nil
}

func LabelUint[T constraints.Unsigned](name string, value T) Label {
	i := uint64(value)
	return func() (string, bool, *string, *bool, *uint64, *int64, *float64) {
		return name, false, nil, nil, &i, nil, nil
	}
}

func LabelUintCond[T constraints.Unsigned](name string, value T, predicate func() bool) Label {
	if predicate() {
		return LabelUint(name, value)
	}
	return nil
}

func LabelInt[T constraints.Signed](name string, value T) Label {
	i := int64(value)
	return func() (string, bool, *string, *bool, *uint64, *int64, *float64) {
		return name, false, nil, nil, nil, &i, nil
	}
}

func LabelIntCond[T constraints.Signed](name string, value T, predicate func() bool) Label {
	if predicate() {
		return LabelInt(name, value)
	}
	return nil
}

func LabelFloat[T constraints.Float](name string, value T) Label {
	f := float64(value)
	return func() (string, bool, *string, *bool, *uint64, *int64, *float64) {
		return name, false, nil, nil, nil, nil, &f
	}
}

func LabelFloatCond[T constraints.Float](name string, value T, predicate func() bool) Label {
	if predicate() {
		return LabelFloat(name, value)
	}
	return nil
}

func BuilderFunc(name string, labels ...Label) string {
	ln := len(name)
	if ln == 0 {
		log.Println("metric name must not be empty, skipping")
		return ""
	}
	if ln > MetricNameMaxLen {
		log.Println("metric name contains too many bytes, skipping")
		return ""
	}
	if strings.Contains(name, `"`) {
		log.Println("metric name contains double quotes, skipping")
		return ""
	}

	if len(labels) == 0 {
		return name
	}

	buf := getBytesBuffer()
	defer putBytesBuffer(buf)

	buf.b = append(buf.b, name...)
	var flLabel bool
	for _, labelFunc := range labels {
		if labelFunc == nil {
			continue
		}

		name, escapeQuote, sv, bv, ui64v, i64v, f64v := labelFunc()

		ln := len(name)
		if ln == 0 {
			log.Printf("metric: %q, label name must not be empty, skipping", buf)
			continue
		}
		if ln > LabelNameMaxLen {
			log.Printf("metric: %q, label name: %q, label name contains too many bytes, skipping", buf, name)
			continue
		}

		// String values need to be checked separately as they can be invalid (empty, or contain too many bytes).
		if sv != nil {
			lv := len(*sv)
			if lv == 0 {
				log.Printf("metric: %q, label name: %q, label value must not be empty, skipping", buf, name)
				continue
			}
			if lv > LabelValueLen {
				log.Printf("metric: %q, label name: %q, label value contains too many bytes, skipping", buf, name)
				continue
			}
		}

		if flLabel { // If we already wrote a label, start writing commas before label names.
			buf.b = append(buf.b, commaByte)
		} else { // Otherwise, mark flag as true for next pass.
			buf.b = append(buf.b, leftBracketByte)
			flLabel = true
		}

		buf.b = append(buf.b, name...)
		buf.b = append(buf.b, equalByte)
		switch {
		case sv != nil:
			buf.b = appendStringValue(buf.b, *sv, escapeQuote)
		case bv != nil:
			buf.b = appendBoolValue(buf.b, *bv)
		case ui64v != nil:
			buf.b = appendUint64Value(buf.b, *ui64v)
		case i64v != nil:
			buf.b = appendInt64Value(buf.b, *i64v)
		case f64v != nil:
			buf.b = appendFloat64Value(buf.b, *f64v)
		default: // Internal problem (wrong use of the label function), panic.
			panic("unsupported case - no label value set")
		}
	}

	if flLabel {
		buf.b = append(buf.b, rightBracketByte)
	}

	return unsafe.String(unsafe.SliceData(buf.b), len(buf.b))
}
