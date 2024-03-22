package vimebu

import "strconv"

// appendStringValue quotes (if needed) and appends s to dst and returns the extended buffer.
func appendStringValue(dst []byte, s string, escapeQuote bool) []byte {
	if escapeQuote {
		return strconv.AppendQuote(dst, s)
	}
	dst = append(dst, doubleQuotesByte)
	dst = append(dst, s...)
	return append(dst, doubleQuotesByte)
}

// appendBoolValue appends b to dst and returns the extended buffer.
func appendBoolValue(dst []byte, b bool) []byte {
	dst = append(dst, doubleQuotesByte)
	dst = strconv.AppendBool(dst, b)
	return append(dst, doubleQuotesByte)
}

// appendInt64Value appends i to dst and returns the extended buffer.
func appendInt64Value(dst []byte, i int64) []byte {
	dst = append(dst, doubleQuotesByte)
	dst = strconv.AppendInt(dst, i, 10)
	return append(dst, doubleQuotesByte)
}

// appendUint64Value appends i to dst and returns the extended buffer.
func appendUint64Value(dst []byte, i uint64) []byte {
	dst = append(dst, doubleQuotesByte)
	dst = strconv.AppendUint(dst, i, 10)
	return append(dst, doubleQuotesByte)
}

// appendFloat64Value appends f to dst and returns the extended buffer.
func appendFloat64Value(dst []byte, f float64) []byte {
	dst = append(dst, doubleQuotesByte)
	dst = strconv.AppendFloat(dst, f, 'f', -1, 64)
	return append(dst, doubleQuotesByte)
}
