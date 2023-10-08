package vimebu

const (
	MetricNameMaxLen = 256  // MetricNameMaxLen is the maximum len in bytes allowed for the metric name.
	LabelNameMaxLen  = 128  // LabelNameMaxLen is the maximum len in bytes allowed for a label name.
	LabelValueLen    = 1024 // LabelValueLen is the maximum len in bytes allowed for a label value.

	leftBracketByte  = byte('{')
	rightBracketByte = byte('}')
	commaByte        = byte(',')
	equalByte        = byte('=')
	doubleQuotesByte = byte('"')
)
