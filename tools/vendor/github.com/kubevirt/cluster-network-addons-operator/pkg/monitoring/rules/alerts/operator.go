package alerts

import (
	"fmt"

	promv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func operatorAlerts(namespace string) []promv1.Rule {
	return []promv1.Rule{
		{
			Alert: "CnaoDown",
			Expr:  intstr.FromString("kubevirt_cnao_operator_up == 0"),
			For:   "5m",
			Annotations: map[string]string{
				"summary": "CNAO pod is down.",
			},
			Labels: map[string]string{
				severityAlertLabelKey:        "warning",
				operatorHealthImpactLabelKey: "warning",
			},
		},
		{
			Alert: "NetworkAddonsConfigNotReady",
			Expr:  intstr.FromString(fmt.Sprintf("sum(kubevirt_cnao_cr_ready{namespace='%s'} or vector(0)) == 0", namespace)),
			For:   "5m",
			Annotations: map[string]string{
				"summary": "CNAO CR NetworkAddonsConfig is not ready.",
			},
			Labels: map[string]string{
				severityAlertLabelKey:        "warning",
				operatorHealthImpactLabelKey: "warning",
			},
		},
	}
}
