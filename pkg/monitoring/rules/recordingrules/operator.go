package recordingrules

import (
	"fmt"

	"github.com/machadovilaca/operator-observability/pkg/operatormetrics"
	"github.com/machadovilaca/operator-observability/pkg/operatorrules"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func operatorRecordingRules(namespace string) []operatorrules.RecordingRule {
	return []operatorrules.RecordingRule{
		{
			MetricsOpts: operatormetrics.MetricOpts{
				Name: "kubevirt_cnao_operator_up",
				Help: "Total count of running CNAO operators",
			},
			MetricType: operatormetrics.GaugeType,
			Expr:       intstr.FromString(fmt.Sprintf("sum(up{namespace='%s', pod=~'cluster-network-addons-operator-.*'} or vector(0))", namespace)),
		},
	}
}
