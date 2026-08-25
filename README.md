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
* `Builder.LabelError` and `Builder.LabelNamedError` for values implementing the `error` interface

### Benchmark comparison
Building the same metric — 4 labels (string, int, error, bool) — with the `fmt` package vs vimebu.

vimebu is roughly 2.7x faster and allocates once instead of twice. The single allocation is the
returned string, which is the floor for any API handing back a `string`; `fmt.Sprintf` adds a
second for the `[]any` it has to build for its variadic arguments.

That gap widens with the label types. `fmt` boxes every argument into an `any`, and only small
integers and `false` come free from the runtime's static cache — pass a float or an int above
255 and it allocates for each one. vimebu stays at one allocation regardless, because
`strconv.Append*` writes into the buffer with no interface conversion.

Medians over 6 runs, `go test -bench=BenchmarkCompare -benchmem -run=NONE -count=6 | benchstat -`
on an Apple M1 Max, Go 1.27, darwin/arm64:

| benchmark | sec/op | B/op | allocs/op |
| --- | --- | --- | --- |
| `CompareFmt` | 186.9n ± 2% | 96 | 2 |
| `CompareVimebu` | **68.90n ± 2%** | 80 | **1** |
| `CompareFmtBoxing` (float + int >255) | 238.5n ± 5% | 128 | 4 |
| `CompareVimebuBoxing` | **110.1n ± 4%** | 96 | **1** |
| `CompareFmtParallel` | 100.0n ± 6% | 96 | 2 |
| `CompareVimebuParallel` | **70.88n ± 3%** | 80 | **1** |
| `CompareEndToEndFmt` | 196.5n ± 2% | 96 | 2 |
| `CompareEndToEndVimebu` | **81.23n ± 3%** | 80 | **1** |

The `EndToEnd` pair includes `GetOrCreateCounter().Inc()`, which is what you actually write on a
hot path with variable label values. VictoriaMetrics' own map lookup and mutex sit on both sides,
so they narrow the gap without closing it.

Note the benchmarks call each formatter **once per iteration**, with the result assigned to a
package-level sink. Calling `fmt.Sprintf` several times per iteration with the same arguments lets
one `[]any` allocation be shared between the calls, which reports 1 alloc/op and hides the real
per-call cost.

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

One consequence worth knowing: `Builder.String` **consumes** a pooled Builder. Since `Builder`
implements `fmt.Stringer`, that also happens if you hand one to a `fmt` verb — so
`fmt.Sprintf("%v", builder)` returns the metric *and* releases the builder, leaving it empty for
any later use. Don't format a Builder you still intend to build with.

#### Concurrency notes
* A Builder instance is not safe to use from concurrently running goroutines
* A Builder instance must not be copied (it embeds a noOp `sync.Locker` implementation, to raise warnings with `go vet` when copied)
