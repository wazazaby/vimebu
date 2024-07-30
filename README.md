# vimebu
[![CI](https://github.com/wazazaby/vimebu/actions/workflows/build-and-test.yml/badge.svg)](https://github.com/wazazaby/vimebu/actions/workflows/build-and-test.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/wazazaby/vimebu.svg)](https://pkg.go.dev/github.com/wazazaby/vimebu)
[![Go Report Card](https://goreportcard.com/badge/github.com/wazazaby/vimebu)](https://goreportcard.com/report/github.com/wazazaby/vimebu)
[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/wazazaby/vimebu/blob/master/LICENSE)

vimebu is a small library that provides a builder to create VictoriaMetrics compatible metrics.

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

var responseSizeHistogram = metrics.NewHistogram(
    vimebu.Metric("response_size").String(), // response_size
)

// Registering the metric using the provided helpers.
var updateTotalCounterV3 = vimebu.
    Metric("update_total").
    LabelInt("version", 3).
    NewCounter() // update_total{version="3"}
```

### Create metrics with variable label values
Vimebu is even more useful when you want to build metrics with variable label values.
```go
import (
    "net"

    "github.com/wazazaby/vimebu/v2"
)

func getCassandraQueryCounter(name string, host net.IP) *metrics.Counter {
    builder := vimebu.Metric("cassandra_query_total")
    builder.LabelString("name", name)
    builder.LabelStringer("host", host)
    return builder.GetOrCreateCounter() // cassandra_query_total{name="beep",host="1.2.3.4"}
}
```

### Create metrics with conditional labels
```go
import (
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
Vimebu also exposes a way to escape quotes on label values you don't control using the following methods :
* `Builder.LabelStringQuote`
* `Builder.LabelStringerQuote`
* `Builder.LabelErrorQuote`

```go
import (
    "github.com/wazazaby/vimebu"
)

func getHTTPRequestCounter(path string) *metrics.Counter {
    builder := vimebu.Metric("api_http_requests_total")
    builder.LabelQuote("path", path)
    return builder.GetOrCreateCounter() // api_http_requests_total{path="some/bro\"ken/path"}
}
```

### Create metrics with label values that aren't string
You can use these methods to append specific value types to the builder :
* `Builder.LabelBool` for booleans
* `Builder.LabelInt` and variations for signed integers
* `Builder.LabelUint` and variations for unsigned integers
* `Builder.LabelFloat` and variations for floats
* `Builder.LabelStringer` for values implementing the `fmt.Stringer` interface
* `Builder.LabelError` for values implementing the `error` interface
