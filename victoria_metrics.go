package vimebu

import (
	"time"

	"github.com/VictoriaMetrics/metrics"
)

// GetOrCreateCounter calls [metrics.GetOrCreateCounter] using the Builder's accumulated string as argument.
func (b *Builder) GetOrCreateCounter() *metrics.Counter {
	return metrics.GetOrCreateCounter(b.String())
}

// GetOrCreateCounterInSet calls [set.GetOrCreateCounter] using the Builder's accumulated string as argument.
func (b *Builder) GetOrCreateCounterInSet(set *metrics.Set) *metrics.Counter {
	return set.GetOrCreateCounter(b.String())
}

// NewCounter calls [metrics.NewCounter] using the Builder's accumulated string as argument.
func (b *Builder) NewCounter() *metrics.Counter {
	return metrics.NewCounter(b.String())
}

// NewCounterInSet calls [set.NewCounter] using the Builder's accumulated string as argument.
func (b *Builder) NewCounterInSet(set *metrics.Set) *metrics.Counter {
	return set.NewCounter(b.String())
}

// GetOrCreateFloatCounter calls [metrics.GetOrCreateFloatCounter] using the Builder's accumulated string as argument.
func (b *Builder) GetOrCreateFloatCounter() *metrics.FloatCounter {
	return metrics.GetOrCreateFloatCounter(b.String())
}

// GetOrCreateFloatCounterInSet calls [set.GetOrCreateFloatCounter] using the Builder's accumulated string as argument.
func (b *Builder) GetOrCreateFloatCounterInSet(set *metrics.Set) *metrics.FloatCounter {
	return set.GetOrCreateFloatCounter(b.String())
}

// NewFloatCounter calls [metrics.NewFloatCounter] using the Builder's accumulated string as argument.
func (b *Builder) NewFloatCounter() *metrics.FloatCounter {
	return metrics.NewFloatCounter(b.String())
}

// NewFloatCounterInSet calls [set.NewFloatCounter] using the Builder's accumulated string as argument.
func (b *Builder) NewFloatCounterInSet(set *metrics.Set) *metrics.FloatCounter {
	return set.NewFloatCounter(b.String())
}

// GetOrCreateHistogram calls [metrics.GetOrCreateHistogram] using the Builder's accumulated string as argument.
func (b *Builder) GetOrCreateHistogram() *metrics.Histogram {
	return metrics.GetOrCreateHistogram(b.String())
}

// GetOrCreateHistogramInSet calls [set.GetOrCreateHistogram] using the Builder's accumulated string as argument.
func (b *Builder) GetOrCreateHistogramInSet(set *metrics.Set) *metrics.Histogram {
	return set.GetOrCreateHistogram(b.String())
}

// NewHistogram calls [metrics.NewHistogram] using the Builder's accumulated string as argument.
func (b *Builder) NewHistogram() *metrics.Histogram {
	return metrics.NewHistogram(b.String())
}

// NewHistogramInSet calls [set.NewHistogram] using the Builder's accumulated string as argument.
func (b *Builder) NewHistogramInSet(set *metrics.Set) *metrics.Histogram {
	return set.NewHistogram(b.String())
}

// GetOrCreateGauge calls [metrics.GetOrCreateGauge] using the Builder's accumulated string as argument.
func (b *Builder) GetOrCreateGauge(f func() float64) *metrics.Gauge {
	return metrics.GetOrCreateGauge(b.String(), f)
}

// GetOrCreateGaugeInSet calls [set.GetOrCreateGauge] using the Builder's accumulated string as argument.
func (b *Builder) GetOrCreateGaugeInSet(set *metrics.Set, f func() float64) *metrics.Gauge {
	return set.GetOrCreateGauge(b.String(), f)
}

// NewGauge calls [metrics.NewGauge] using the Builder's accumulated string as argument.
func (b *Builder) NewGauge(f func() float64) *metrics.Gauge {
	return metrics.NewGauge(b.String(), f)
}

// NewGaugeInSet calls [set.NewGauge] using the Builder's accumulated string as argument.
func (b *Builder) NewGaugeInSet(set *metrics.Set, f func() float64) *metrics.Gauge {
	return set.NewGauge(b.String(), f)
}

// GetOrCreateSummary calls [metrics.GetOrCreateSummary] using the Builder's accumulated string as argument.
func (b *Builder) GetOrCreateSummary() *metrics.Summary {
	return metrics.GetOrCreateSummary(b.String())
}

// GetOrCreateSummaryInSet calls [set.GetOrCreateSummary] using the Builder's accumulated string as argument.
func (b *Builder) GetOrCreateSummaryInSet(set *metrics.Set) *metrics.Summary {
	return set.GetOrCreateSummary(b.String())
}

// NewSummary calls [metrics.NewSummary] using the Builder's accumulated string as argument.
func (b *Builder) NewSummary() *metrics.Summary {
	return metrics.NewSummary(b.String())
}

// NewSummaryInSet calls [set.NewSummary] using the Builder's accumulated string as argument.
func (b *Builder) NewSummaryInSet(set *metrics.Set) *metrics.Summary {
	return set.NewSummary(b.String())
}

// GetOrCreateSummaryExt calls [metrics.GetOrCreateSummaryExt] using the Builder's accumulated string as argument.
func (b *Builder) GetOrCreateSummaryExt(window time.Duration, quantiles []float64) *metrics.Summary {
	return metrics.GetOrCreateSummaryExt(b.String(), window, quantiles)
}

// GetOrCreateSummaryExtInSet calls [set.GetOrCreateSummaryExt] using the Builder's accumulated string as argument.
func (b *Builder) GetOrCreateSummaryExtInSet(set *metrics.Set, window time.Duration, quantiles []float64) *metrics.Summary {
	return set.GetOrCreateSummaryExt(b.String(), window, quantiles)
}

// GetOrCreateHistogram calls [metrics.GetOrCreateHistogram] using the Builder's accumulated string as argument.
func (b *Builder) NewSummaryExt(window time.Duration, quantiles []float64) *metrics.Summary {
	return metrics.NewSummaryExt(b.String(), window, quantiles)
}

// GetOrCreateHistogramInSet calls [metrics.GetOrCreateHistogram] using the Builder's accumulated string as argument.
func (b *Builder) NewSummaryExtInSet(set *metrics.Set, window time.Duration, quantiles []float64) *metrics.Summary {
	return set.NewSummaryExt(b.String(), window, quantiles)
}
