package metrics

import "github.com/machadovilaca/operator-observability/pkg/operatormetrics"

var (
	metrics = [][]operatormetrics.Metric{
		operatorMetrics,
	}
)

func SetupMetrics() error {
	return operatormetrics.RegisterMetrics(metrics...)
}

func ListMetrics() []operatormetrics.Metric {
	return operatormetrics.ListMetrics()
}
