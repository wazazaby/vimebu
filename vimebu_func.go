package vimebu

import (
	"log"
	"strconv"
	"strings"
)

type LabelCallback func() (name, value string, escapeQuote bool)

// WithLabel appends a pair of label name and label value to the builder.
// Unlike [vimebu.WithLabelQuote], quotes inside the label value will not be escaped.
// It's better suited for a label value where you control the input (either it is already sanitized, or it comes from a const or an enum for example).
//
// Panics if the label name or label value contain more than [vimebu.LabelNameMaxLen] or [vimebu.LabelValueMaxLen] bytes respectively.
func WithLabel(name, value string) LabelCallback {
	if len(name) > LabelNameMaxLen {
		log.Println("label name contains too many bytes, skipping")
		return nil
	}

	if len(value) > LabelValueLen {
		log.Println("label value contains too many bytes, skipping")
		return nil
	}

	return func() (string, string, bool) {
		return name, value, false
	}
}

// WithLabelCond is a wrapper around [vimebu.WithLabel].
// It allows you to conditionally add a label depending on the output of a predicate function.
//
// Panics if the label name or label value contain more than [vimebu.LabelNameMaxLen] or [vimebu.LabelValueMaxLen] bytes respectively.
func WithLabelCond(fn func() bool, name, value string) LabelCallback {
	if fn() {
		return WithLabel(name, value)
	}
	return WithLabel("", "")
}

// WithLabelQuote appends a pair of label name and label value to the builder. Quotes inside the label value will be escaped.
//
// Panics if the label name or label value contain more than [vimebu.LabelNameMaxLen] or [vimebu.LabelValueMaxLen] bytes respectively.
func WithLabelQuote(name, value string) LabelCallback {
	if len(name) > LabelNameMaxLen {
		log.Println("label name contains too many bytes, skipping")
		return nil
	}

	if len(value) > LabelValueLen {
		log.Println("label value contains too many bytes, skipping")
		return nil
	}

	return func() (string, string, bool) {
		return name, value, true
	}
}

// WithLabelQuoteCond is a wrapper around [vimebu.WithLabelQuote].
// It allows you to conditionally add a label depending on the output of a predicate function.
//
// Panics if the label name or label value contain more than [vimebu.LabelNameMaxLen] or [vimebu.LabelValueMaxLen] bytes respectively.
func WithLabelQuoteCond(fn func() bool, name, value string) LabelCallback {
	if fn() {
		return WithLabelQuote(name, value)
	}
	return WithLabelQuote("", "")
}

// BuilderFunc is used to efficiently build a VictoriaMetrics metric.
// It's backed by a bytes.Buffer to minimize memory copying.
//
// NoOp if :
// * the name is empty or contains more than [vimebu.MetricNameMaxLen] bytes.
// * the name contains a double quote.
func BuilderFunc(name string, labels ...LabelCallback) string {
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

	// In case there are no labels, using a strings.Builder is actually
	// about ~9% faster than concatenating the name and brackets.
	if len(labels) == 0 {
		var b strings.Builder
		b.Grow(ln + 2)
		b.WriteString(name)
		b.WriteString("{}")
		return b.String()
	}

	b := getBuffer()
	defer putBuffer(b)

	b.WriteString(name)
	b.WriteByte(leftBracketByte)
	var flLabel bool
	for _, callback := range labels {
		if callback == nil {
			continue
		}

		name, value, escapeQuote := callback()
		if name == "" {
			log.Println("label name must not be empty, skipping")
			continue
		}
		if value == "" {
			log.Println("label value must not be empty, skipping")
			continue
		}

		if flLabel {
			b.WriteByte(commaByte)
		} else {
			flLabel = true
		}

		b.WriteString(name)
		b.WriteByte(equalByte)
		if escapeQuote && strings.Contains(value, `"`) {
			buf := b.AvailableBuffer()
			quoted := strconv.AppendQuote(buf, value)
			b.Write(quoted)
		} else {
			b.WriteByte(doubleQuotesByte)
			b.WriteString(value)
			b.WriteByte(doubleQuotesByte)
		}
	}

	b.WriteByte(rightBracketByte)
	return b.String()
}
