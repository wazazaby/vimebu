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
        String() // cassandra_query_total{name="beep",host="1.2.3.4"}
    return vm.GetOrCreateCounter(metric)
}
```

You can also register a map of labels directly.
```go
import (
    "net"

    "github.com/wazazaby/vimebu"
    vm "github.com/VictoriaMetrics/metrics"
)

func getNodeMemoryUsageGauge(kvs map[string]string) *vm.Counter {
    metric := vimebu.
        Metric("node_memory_usage_bytes").
        Labels(kvs)
        String() // node_memory_usage_bytes{id="compute_1",store="whatever",cluster="guava"}
    return vm.GetOrCreateGauge(metric, func() float64 { return 0 })
}
```
