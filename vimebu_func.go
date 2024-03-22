package vimebu

import (
	"fmt"
	"log"
	"strings"
	"unsafe"

	"golang.org/x/exp/constraints"
)

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

	b := getBuffer()
	defer putBuffer(b)

	b.WriteString(name)
	var flLabel bool
	for _, labelFunc := range labels {
		if labelFunc == nil {
			continue
		}

		name, escapeQuote, sv, bv, ui64v, i64v, f64v := labelFunc()

		ln := len(name)
		if ln == 0 {
			log.Printf("metric: %q, label name must not be empty, skipping", b)
			continue
		}
		if ln > LabelNameMaxLen {
			log.Printf("metric: %q, label name: %q, label name contains too many bytes, skipping", b, name)
			continue
		}

		// String values need to be checked separately as they can be invalid (empty, or contain too many bytes).
		if sv != nil {
			lv := len(*sv)
			if lv == 0 {
				log.Printf("metric: %q, label name: %q, label value must not be empty, skipping", b, name)
				continue
			}
			if lv > LabelValueLen {
				log.Printf("metric: %q, label name: %q, label value contains too many bytes, skipping", b, name)
				continue
			}
		}

		if flLabel { // If we already wrote a label, start writing commas before label names.
			b.WriteByte(commaByte)
		} else { // Otherwise, mark flag as true for next pass.
			b.WriteByte(leftBracketByte)
			flLabel = true
		}

		b.WriteString(name)
		b.WriteByte(equalByte)
		buf := b.AvailableBuffer()
		switch {
		case sv != nil:
			buf = appendStringValue(buf, *sv, escapeQuote)
		case bv != nil:
			buf = appendBoolValue(buf, *bv)
		case ui64v != nil:
			buf = appendUint64Value(buf, *ui64v)
		case i64v != nil:
			buf = appendInt64Value(buf, *i64v)
		case f64v != nil:
			buf = appendFloat64Value(buf, *f64v)
		default: // Internal problem (wrong use of the label function), panic.
			panic("unsupported case - no label value set")
		}
		b.Write(buf)
	}

	if flLabel {
		b.WriteByte(rightBracketByte)
	}

	return unsafe.String(unsafe.SliceData(b.Bytes()), b.Len())
}
