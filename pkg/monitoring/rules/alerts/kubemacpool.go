package alerts

import (
	promv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"

	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
)

var kubemacpoolAlerts = []promv1.Rule{
	{
		Alert: "KubeMacPoolDuplicateMacsFound",
		Expr:  intstr.FromString("kubevirt_cnao_kubemacpool_duplicate_macs != 0"),
		For:   ptr.To(promv1.Duration("5m")),
		Annotations: map[string]string{
			"summary": "Duplicate macs found.",
		},
		Labels: map[string]string{
			severityAlertLabelKey:        "warning",
			operatorHealthImpactLabelKey: "warning",
		},
	},
	{
		Alert: "KubemacpoolDown",
		Expr:  intstr.FromString("kubevirt_cnao_cr_kubemacpool_aggregated == 1 and kubevirt_cnao_kubemacpool_manager_up == 0"),
		For:   ptr.To(promv1.Duration("5m")),
		Annotations: map[string]string{
			"summary": "KubeMacpool is deployed by CNAO CR but KubeMacpool pod is down.",
		},
		Labels: map[string]string{
			severityAlertLabelKey:        "critical",
			operatorHealthImpactLabelKey: "critical",
		},
	},
}
