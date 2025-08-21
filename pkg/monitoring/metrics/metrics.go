package metrics

import "github.com/rhobs/operator-observability-toolkit/pkg/operatormetrics"

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
