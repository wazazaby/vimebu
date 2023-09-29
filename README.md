# vimebu
Vimebu is a small library that provides a builder to create VictoriaMetrics compatible metrics.

### Usage
```go
import (
    "github.com/wazazaby/vimebu"
    vm "github.com/VictoriaMetrics/metrics"
)

var (
    requestsTotalCounter = vm.NewCounter(
        vimebu.Metric("request_total").Label("path", "/foo/bar").String(), // request_total{path="/foo/bar"}
    )

    responseSizeHistogram = vm.NewHistogram(
        vimebu.Metric("response_size").String(), // response_size{}
    )
)
```

It's even more useful when you want to build metrics with variable label values.
```go
import (
    "net"
    "github.com/wazazaby/vimebu"
    vm "github.com/VictoriaMetrics/metrics"
)

func getCassandraQueryCounter(name string, host net.IP) *vm.Counter {
    metric := vimebu.
        Metric("cassandra_query_total").
        Label("name", name).
        Label("host", host.String()).
        String() // cassandra_query_total{name="beep",host="1.2.3.4"}
    return vm.GetOrCreateCounter(metric)
}
```

Or if you want to conditionally add labels.
```go
import (
    "github.com/wazazaby/vimebu"
    vm "github.com/VictoriaMetrics/metrics"
)

func getHTTPRequestCounter(host string) *vm.Counter {
    var b vimebu.Builder
    b.Metric("api_http_requests_total")
    if host != "" {
        b.Label("host", host)
    }
    metric := b.String() // api_http_requests_total{} or api_http_requests_total{host="api.app.com"}
    return vm.GetOrCreateCounter(metric)
}
```

Vimebu also exposes a way to escape quotes on label values you don't control using `Builder.LabelQuote`.
```go
import (
    "github.com/wazazaby/vimebu"
    vm "github.com/VictoriaMetrics/metrics"
)

func getHTTPRequestCounter(path string) *vm.Counter {
    metric := vimebu.
        Metric("api_http_requests_total").
        LabelQuote("path", path).
        String() // api_http_requests_total{path="some/bro\"ken/path"}
    return vm.GetOrCreateCounter(metric)
}
```
