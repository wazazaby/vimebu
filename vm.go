package vimebu

import (
	"fmt"

	"github.com/VictoriaMetrics/metrics"
)

var defaultSet = metrics.GetDefaultSet()

// validateMetric just forwards the call to [metrics.ValidateMetric] and wraps
// the returned error with a nicer message.
func validateMetric(name string) error {
	if err := metrics.ValidateMetric(name); err != nil {
		return fmt.Errorf("invalid metric, err: %w", err)
	}
	return nil
}
