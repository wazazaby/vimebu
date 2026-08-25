# vimebu
[![CI](https://github.com/wazazaby/vimebu/actions/workflows/build-and-test.yml/badge.svg)](https://github.com/wazazaby/vimebu/actions/workflows/build-and-test.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/wazazaby/vimebu.svg)](https://pkg.go.dev/github.com/wazazaby/vimebu/v2)
[![Go Report Card](https://goreportcard.com/badge/github.com/wazazaby/vimebu)](https://goreportcard.com/report/github.com/wazazaby/vimebu)
[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/wazazaby/vimebu/blob/master/LICENSE)

vimebu provides a type-safe builder to create VictoriaMetrics compatible metrics. 

It aims to be as CPU & memory efficient as possible using strategies such as object pooling, buffer reuse etc.

## Installation
`go get -u github.com/wazazaby/vimebu/v2`

## Usage
```go
import (
    "github.com/VictoriaMetrics/metrics"
    "github.com/wazazaby/vimebu/v2"
)

// Only using the builder.
var requestsTotalCounter = metrics.NewCounter(
    vimebu.
        Metric("request_total").
        LabelString("path", "/foo/bar").
        String(), // request_total{path="/foo/bar"}
)

// Registering the metric using the provided helpers.
var updateTotalCounterV3 = vimebu.
    Metric("update_total").
    LabelInt("version", 3).
    NewCounter() // update_total{version="3"}
```

### Create metrics with variable label values
vimebu is even more useful when you want to build metrics with variable label values.
```go
import (
    "net"

    "github.com/VictoriaMetrics/metrics"
    "github.com/wazazaby/vimebu/v2"
)

func getCassandraQueryCounter(name string, host net.IP, err error) *metrics.Counter {
    return vimebu.Metric("cassandra_query_total").
        LabelString("name", name).
        LabelStringer("host", host).
        LabelErrorQuote(err). // The label "error" won't be added if err is nil.
        GetOrCreateCounter() // cassandra_query_total{name="beep",host="1.2.3.4",error="i/o timeout"}
}
```

### Create metrics with conditional labels
You can also have metrics with labels that are added under certain conditions.
```go
import (
    "github.com/VictoriaMetrics/metrics"
    "github.com/wazazaby/vimebu/v2"
)

func getHTTPRequestCounter(host string) *metrics.Counter {
    builder := vimebu.Metric("api_http_requests_total")
    if host != "" {
        builder.LabelString("host", host)
    }
    return builder.GetOrCreateCounter() // api_http_requests_total or api_http_requests_total{host="api.app.com"}
}
```

### Create metrics with label values that need to be escaped
For label values you don't control, use the `Quote` variants :
* `Builder.LabelStringQuote`
* `Builder.LabelStringerQuote`
* `Builder.LabelErrorQuote`
* `Builder.LabelNamedErrorQuote`

They escape the three characters the exposition format needs escaped inside a label value:
backslash, double quote and newline. Anything else - including non-ASCII runes and control
characters - is passed through as-is.

```go
import (
    "github.com/VictoriaMetrics/metrics"
    "github.com/wazazaby/vimebu/v2"
)

func getHTTPRequestCounter(path string) *metrics.Counter {
    return vimebu.Metric("api_http_requests_total").
      LabelStringQuote("path", path).
      GetOrCreateCounter() // api_http_requests_total{path="some/bro\"ken/path"}
}
```

### Create metrics with label values that aren't strings
You can use these methods to append specific value types to the builder :
* `Builder.LabelBool` for booleans
* `Builder.LabelInt` and variations for signed integers
* `Builder.LabelUint` and variations for unsigned integers
* `Builder.LabelFloat32` and `Builder.LabelFloat64` for floats
* `Builder.LabelStringer` for values implementing the `fmt.Stringer` interface
* `Builder.LabelError` for values implementing the `error` interface

### Benchmark comparison
Some simple benchmarks comparing building a metric with the `fmt` package vs vimebu.
Each loop builds 8 metrics, each with 4 labels (string, int, error and bool).

vimebu is about 2.3x faster sequentially and ~25% faster in parallel. Both allocate
once per built metric: that single allocation is the returned string, and it's the floor
for any API handing back a `string`.

Medians over 6 runs, `go test -bench="BenchmarkCompare" -benchmem -run=NONE -count=6 | benchstat -`
on an Apple M1 Max, Go 1.27, darwin/arm64:

| benchmark | sec/op | B/op | allocs/op |
| --- | --- | --- | --- |
| `CompareSequentialFmt` | 1.562µ ± 1% | 896 | 8 |
| `CompareSequentialVimebu` | **665.0n ± 1%** | 896 | 8 |
| `CompareParralelFmt` | 586.9n ± 11% | 896 | 8 |
| `CompareParralelVimebu` | **439.3n ± 28%** | 896 | 8 |

### Under the hood
Builders can be acquired and released using a BuilderPool, which is a wrapper around a `sync.Pool` instance.
A default BuilderPool instance is created and exposed by the package, it is accessible like this :

```go
import (
    "github.com/VictoriaMetrics/metrics"
    "github.com/wazazaby/vimebu/v2"
)

func getHTTPRequestCounter(path string) *metrics.Counter {
    builder := vimebu.AcquireBuilder()
    defer vimebu.ReleaseBuilder(builder)

    builder.Metric("api_http_requests_total").LabelStringQuote("path", path)
    return builder.GetOrCreateCounter() // api_http_requests_total{path="some/bro\"ken/path"}
}
```

Using a pool allows for reusing objects, thus relieving pressure on the garbage collector.

Understanding that this syntax can be quite verbose, vimebu also provides a simpler API that manages the lifecycle
of these objects internally by using the `vimebu.Metric` package level function.
Here, vimebu will automatically acquire a Builder, to finally reset and release it when the `Builder.String` method is called.

#### Concurrency notes
* A Builder instance is not safe to use from concurrently running goroutines
* A Builder instance must not be copied (it embeds a noOp `sync.Locker` implementation, to raise warnings with `go vet` when copied)
