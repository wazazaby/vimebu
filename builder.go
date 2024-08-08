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
)

// It is forbidden copying URI instances. Create new instance and use CopyTo
// instead.
//
// URI instance MUST NOT be used from concurrently running goroutines.
type Builder struct {
	noCopy noCopy

	buf []byte

	flName     bool
	flLabel    bool
	flAcquired bool
}

func (b *Builder) Reset() {
	b.buf = b.buf[:0]
	b.flName = false
	b.flLabel = false
	b.flAcquired = false
}

func Metric(name string) *Builder {
	builder := DefaultBuilderPool.Acquire()
	builder.flAcquired = true
	return builder.Metric(name)
}

func (b *Builder) Metric(name string) *Builder {
	if b.flName {
		return b
	}

	b.buf = append(b.buf, name...)
	b.flName = true
	return b
}

func (b *Builder) commaOrLeftBracket() byte {
	if b.flLabel {
		return commaByte
	}
	b.flLabel = true
	return leftBracketByte
}

func (b *Builder) LabelString(name, value string) *Builder {
	if len(name) <= 0 || len(value) <= 0 {
		return b
	}
	b.buf = append(b.buf, b.commaOrLeftBracket())
	b.buf = append(b.buf, name...)
	b.buf = append(b.buf, equalByte)
	b.buf = append(b.buf, doubleQuotesByte)
	b.buf = append(b.buf, value...)
	b.buf = append(b.buf, doubleQuotesByte)
	return b
}

func (b *Builder) LabelStringQuote(name, value string) *Builder {
	if len(name) <= 0 || len(value) <= 0 {
		return b
	}
	b.buf = append(b.buf, b.commaOrLeftBracket())
	b.buf = append(b.buf, name...)
	b.buf = append(b.buf, equalByte)
	b.buf = strconv.AppendQuote(b.buf, value)
	return b
}

func (b *Builder) LabelErrQuote(name string, err error) *Builder {
	return b.LabelStringQuote(name, err.Error())
}

func (b *Builder) LabelErr(name string, err error) *Builder {
	return b.LabelString(name, err.Error())
}

func (b *Builder) LabelBool(name string, value bool) *Builder {
	if value {
		return b.LabelString(name, "true")
	}
	return b.LabelString(name, "false")
}

func (b *Builder) LabelUint(name string, value uint) *Builder {
	return b.LabelUint64(name, uint64(value))
}

func (b *Builder) LabelUint8(name string, value uint8) *Builder {
	return b.LabelUint64(name, uint64(value))
}

func (b *Builder) LabelUint16(name string, value uint16) *Builder {
	return b.LabelUint64(name, uint64(value))
}

func (b *Builder) LabelUint32(name string, value uint32) *Builder {
	return b.LabelUint64(name, uint64(value))
}

func (b *Builder) LabelUint64(name string, value uint64) *Builder {
	if len(name) <= 0 {
		return b
	}
	b.buf = append(b.buf, b.commaOrLeftBracket())
	b.buf = append(b.buf, name...)
	b.buf = append(b.buf, equalByte)
	b.buf = append(b.buf, doubleQuotesByte)
	b.buf = strconv.AppendUint(b.buf, value, base10)
	b.buf = append(b.buf, doubleQuotesByte)
	return b
}

func (b *Builder) LabelInt(name string, value int) *Builder {
	return b.LabelInt64(name, int64(value))
}

func (b *Builder) LabelInt8(name string, value int8) *Builder {
	return b.LabelInt64(name, int64(value))
}

func (b *Builder) LabelInt16(name string, value int16) *Builder {
	return b.LabelInt64(name, int64(value))
}

func (b *Builder) LabelInt32(name string, value int32) *Builder {
	return b.LabelInt64(name, int64(value))
}

func (b *Builder) LabelInt64(name string, value int64) *Builder {
	if len(name) <= 0 {
		return b
	}
	b.buf = append(b.buf, b.commaOrLeftBracket())
	b.buf = append(b.buf, name...)
	b.buf = append(b.buf, equalByte)
	b.buf = append(b.buf, doubleQuotesByte)
	b.buf = strconv.AppendInt(b.buf, value, base10)
	b.buf = append(b.buf, doubleQuotesByte)
	return b
}

func (b *Builder) LabelFloat32(name string, value float32) *Builder {
	return b.LabelFloat64(name, float64(value))
}

func (b *Builder) LabelFloat64(name string, value float64) *Builder {
	if len(name) <= 0 {
		return b
	}
	b.buf = append(b.buf, b.commaOrLeftBracket())
	b.buf = append(b.buf, name...)
	b.buf = append(b.buf, equalByte)
	b.buf = append(b.buf, doubleQuotesByte)
	b.buf = strconv.AppendFloat(b.buf, value, floatFormattingVerb, floatShortestPrecision, floatBitSize)
	b.buf = append(b.buf, doubleQuotesByte)
	return b
}

func (b *Builder) LabelStringer(name string, value fmt.Stringer) *Builder {
	return b.LabelString(name, value.String())
}

func (b *Builder) LabelStringerQuote(name string, value fmt.Stringer) *Builder {
	return b.LabelStringQuote(name, value.String())
}

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
