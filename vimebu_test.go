package vimebu

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type label struct {
	name, value string
}
type input struct {
	name   string
	labels []label
}
type testCase struct {
	name     string
	expected string
	input    input
}

var testCases = []testCase{
	{
		name: "metric with labels",
		input: input{
			labels: []label{{"cluster", "guava"}, {"host", "1.2.3.4"}},
			name:   "cassandra_query_count",
		},
		expected: `cassandra_query_count{cluster="guava",host="1.2.3.4"}`,
	},
	{
		name: "metric with single label",
		input: input{
			labels: []label{{"type", "std"}},
			name:   "produce_one_total",
		},
		expected: `produce_one_total{type="std"}`,
	},
	{
		name: "metric without label",
		input: input{
			name: "http_request_duration_seconds",
		},
		expected: `http_request_duration_seconds{}`,
	},
	{
		name:     "no name",
		expected: ``,
	},
	{
		name: "some empty labels and values",
		input: input{
			labels: []label{{"operation", ""}, {"", "1.2.3.4"}, {"status", "OK"}, {"node", ""}},
			name:   "api_http_requests_total",
		},
		expected: `api_http_requests_total{status="OK"}`,
	},
	{
		name: "values contain double quotes",
		input: input{
			labels: []label{{"error", `something went "horribly" wrong`}, {"path", `some/path/"with"/quo"tes`}},
			name:   "api_http_requests_total",
		},
		expected: `api_http_requests_total{error="something went \"horribly\" wrong",path="some/path/\"with\"/quo\"tes"}`,
	},
}

func TestBuilder(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var b Builder
			b.Grow(128)
			require.Equal(t, 128, b.underlying.Cap())
			b.Metric(tc.input.name)
			for _, label := range tc.input.labels {
				if strings.Contains(label.value, `"`) {
					b.LabelAppendQuote(label.name, label.value)
				} else {
					b.Label(label.name, label.value)
				}
			}
			result := b.String()
			require.Equal(t, tc.expected, result)

			t.Run("reset", func(t *testing.T) {
				b.Reset()
				require.False(t, b.flName)
				require.False(t, b.flLabel)
			})
		})
	}
}

var (
	status  = "Bad Request"
	path    = `some/path/"with"/quo"tes`
	host    = "1.2.3.4"
	cluster = "guava"
)

func BenchmarkBuilder(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_ = Metric("http_request_duration_seconds").
			Label("status", status).
			LabelAppendQuote("path", path).
			Label("host", host).
			Label("cluster", cluster).
			String()
	}
}

func BenchmarkBuilderNoAppendQuote(b *testing.B) {
	pathSafe := strconv.Quote(path)
	for n := 0; n < b.N; n++ {
		_ = Metric("http_request_duration_seconds").
			Label("status", status).
			Label("path", pathSafe).
			Label("host", host).
			Label("cluster", cluster).
			String()
	}
}

func BenchmarkBuilderAppendQuote(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_ = Metric("http_request_duration_seconds").
			LabelAppendQuote("status", status).
			LabelAppendQuote("path", path).
			LabelAppendQuote("host", host).
			LabelAppendQuote("cluster", cluster).
			String()
	}
}

func BenchmarkStringsBuilder(b *testing.B) {
	for n := 0; n < b.N; n++ {
		var b strings.Builder
		b.WriteString("http_request_duration_seconds{status=" + strconv.Quote(status) + ",path=" + strconv.Quote(path) + ",host=" + strconv.Quote(host) + ",cluster=" + strconv.Quote(cluster) + "}")
		_ = b.String()
	}
}

func BenchmarkBytesBuffer(b *testing.B) {
	for n := 0; n < b.N; n++ {
		var b bytes.Buffer
		b.WriteString("http_request_duration_seconds{status=" + strconv.Quote(status) + ",path=" + strconv.Quote(path) + ",host=" + strconv.Quote(host) + ",cluster=" + strconv.Quote(cluster) + "}")
		_ = b.String()
	}
}

func BenchmarkBytesBufferFastQuote(b *testing.B) {
	for n := 0; n < b.N; n++ {
		var b bytes.Buffer
		b.WriteString("http_request_duration_seconds{status=")
		buf := b.AvailableBuffer()
		quoted := strconv.AppendQuote(buf, status)
		b.Write(quoted)

		b.WriteString(",path=")
		buf = b.AvailableBuffer()
		quoted = strconv.AppendQuote(buf, path)
		b.Write(quoted)

		b.WriteString(",host=")
		buf = b.AvailableBuffer()
		quoted = strconv.AppendQuote(buf, cluster)
		b.Write(quoted)

		b.WriteString(",cluster=")
		buf = b.AvailableBuffer()
		quoted = strconv.AppendQuote(buf, cluster)
		b.Write(quoted)

		b.WriteString("}")

		_ = b.String()
	}
}

func BenchmarkFmtSprintf(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_ = fmt.Sprintf("http_request_duration_seconds{status=%q,path=%q,host=%q,cluster=%q}", status, path, host, cluster)
	}
}
