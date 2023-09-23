# vimebu
Vimebu is a small library that provides a builder to create Prometheus & VictoriaMetrics compatible metrics.

### Usage
```go
import (
    "github.com/wazazaby/vimebu"
    vm "github.com/VictoriaMetrics/metrics"
)

var (
    requestsTotalCounter = metrics.NewCounter(
        vimebu.Metric("request_total").Label("path", "/foo/bar").String(), // request_total{path="/foo/bar"}
    )

    responseSizeHistogram = metrics.NewHistogram(
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
        String()
    return vm.GetOrCreateCounter(metric)
}

```