# vimebu
[![CI](https://github.com/wazazaby/vimebu/actions/workflows/build-and-test.yml/badge.svg)](https://github.com/wazazaby/vimebu/actions/workflows/build-and-test.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/wazazaby/vimebu.svg)](https://pkg.go.dev/github.com/wazazaby/vimebu)
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

    "github.com/wazazaby/vimebu/v2"
)

func getCassandraQueryCounter(name string, host net.IP, err error) *metrics.Counter {
    return vimebu.Metric("cassandra_query_total").
        LabelString("name", name).
        LabelStringer("host", host).
        LabelErrorQuote("error", err). // The label "error" won't be added if err is nil.
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
vimebu also exposes a way to escape quotes on label values you don't control using the following methods :
* `Builder.LabelStringQuote`
* `Builder.LabelStringerQuote`
* `Builder.LabelErrorQuote`

```go
import (
    "github.com/VictoriaMetrics/metrics"
    "github.com/wazazaby/vimebu/v2"
)

func getHTTPRequestCounter(path string) *metrics.Counter {
    return vimebu.Metric("api_http_requests_total").
      LabelQuote("path", path).
      GetOrCreateCounter() // api_http_requests_total{path="some/bro\"ken/path"}
}
```

### Create metrics with label values that aren't strings
You can use these methods to append specific value types to the builder :
* `Builder.LabelBool` for booleans
* `Builder.LabelInt` and variations for signed integers
* `Builder.LabelUint` and variations for unsigned integers
* `Builder.LabelFloat` and variations for floats
* `Builder.LabelStringer` for values implementing the `fmt.Stringer` interface
* `Builder.LabelError` for values implementing the `error` interface

### Benchmark comparison
Here are some simple benchmarks comparing building a metric using the `fmt` package vs vimebu.
Each metric is built with 4 labels (string, int, error and bool).

As you can see, in the sequential benchmarks, vimebu is about twice as fast.
For the parralel benchmarks, vimebu is about ~30% faster.

In each case, it allocates half as much per operation. Yay!
```
❯ go test -bench="BenchmarkCompare" -benchmem -run=NONE
goos: darwin
goarch: arm64
pkg: github.com/wazazaby/vimebu/v2
cpu: Apple M1 Max
BenchmarkCompareSequentialFmt-10          717055              1660 ns/op            1024 B/op         16 allocs/op
BenchmarkCompareParralelFmt-10           2279166               528.0 ns/op          1024 B/op         16 allocs/op
BenchmarkCompareSequentialVimebu-10      1557618               765.0 ns/op           896 B/op          8 allocs/op
BenchmarkCompareParralelVimebu-10        3098870               398.5 ns/op           896 B/op          8 allocs/op
PASS
ok      github.com/wazazaby/vimebu/v2   7.445s
```

### Under the hood
Builders can be acquired and released using a BuilderPool, which is a wrapper around a `sync.Pool` instance.
A default BuilderPool instance is created and exposed by the package, it is accessible like this :

```go
import (
    "github.com/VictoriaMetrics/metrics"
    "github.com/wazazaby/vimebu/v2"
)

func getHTTPRequestCounter(path string) *metrics.Counter {
    builder := vimebu.DefaultBuilderPool.Acquire()
    defer vimebu.DefaultBuilderPool.Release(builder)

    builder.LabelQuote("path", path)
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
