package vimebu

import (
	"time"

	"github.com/VictoriaMetrics/metrics"
)

// GetOrCreateCounter calls [metrics.GetOrCreateCounter] using the Builder's accumulated string as argument.
func (b *Builder) GetOrCreateCounter() *metrics.Counter {
	return metrics.GetOrCreateCounter(b.String())
}

// NewCounter calls [metrics.NewCounter] using the Builder's accumulated string as argument.
func (b *Builder) NewCounter() *metrics.Counter {
	return metrics.NewCounter(b.String())
}

// GetOrCreateFloatCounter calls [metrics.GetOrCreateFloatCounter] using the Builder's accumulated string as argument.
func (b *Builder) GetOrCreateFloatCounter() *metrics.FloatCounter {
	return metrics.GetOrCreateFloatCounter(b.String())
}

// NewFloatCounter calls [metrics.NewFloatCounter] using the Builder's accumulated string as argument.
func (b *Builder) NewFloatCounter() *metrics.FloatCounter {
	return metrics.NewFloatCounter(b.String())
}

// GetOrCreateHistogram calls [metrics.GetOrCreateHistogram] using the Builder's accumulated string as argument.
func (b *Builder) GetOrCreateHistogram() *metrics.Histogram {
	return metrics.GetOrCreateHistogram(b.String())
}

// NewHistogram calls [metrics.NewHistogram] using the Builder's accumulated string as argument.
func (b *Builder) NewHistogram() *metrics.Histogram {
	return metrics.NewHistogram(b.String())
}

// GetOrCreateGauge calls [metrics.GetOrCreateGauge] using the Builder's accumulated string as argument.
func (b *Builder) GetOrCreateGauge(f func() float64) *metrics.Gauge {
	return metrics.GetOrCreateGauge(b.String(), f)
}

// NewGauge calls [metrics.NewGauge] using the Builder's accumulated string as argument.
func (b *Builder) NewGauge(f func() float64) *metrics.Gauge {
	return metrics.NewGauge(b.String(), f)
}

// GetOrCreateSummary calls [metrics.GetOrCreateSummary] using the Builder's accumulated string as argument.
func (b *Builder) GetOrCreateSummary() *metrics.Summary {
	return metrics.GetOrCreateSummary(b.String())
}

// NewSummary calls [metrics.NewSummary] using the Builder's accumulated string as argument.
func (b *Builder) NewSummary() *metrics.Summary {
	return metrics.NewSummary(b.String())
}

// GetOrCreateSummaryExt calls [metrics.GetOrCreateSummaryExt] using the Builder's accumulated string as argument.
func (b *Builder) GetOrCreateSummaryExt(window time.Duration, quantiles []float64) *metrics.Summary {
	return metrics.GetOrCreateSummaryExt(b.String(), window, quantiles)
}

// GetOrCreateHistogram calls [metrics.GetOrCreateHistogram] using the Builder's accumulated string as argument.
func (b *Builder) NewSummaryExt(window time.Duration, quantiles []float64) *metrics.Summary {
	return metrics.NewSummaryExt(b.String(), window, quantiles)
}
