package alerts

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"

	promv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

func operatorAlerts(namespace string) []promv1.Rule {
	return []promv1.Rule{
		{
			Alert: "CnaoDown",
			Expr:  intstr.FromString("cluster:kubevirt_cnao_operator_up:sum == 0"),
			For:   ptr.To(promv1.Duration("5m")),
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
			For:   ptr.To(promv1.Duration("5m")),
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
