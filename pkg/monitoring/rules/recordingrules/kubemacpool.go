package recordingrules

import (
	"fmt"

	"github.com/rhobs/operator-observability-toolkit/pkg/operatormetrics"
	"github.com/rhobs/operator-observability-toolkit/pkg/operatorrules"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func kubemacpoolRecordingRules(namespace string) []operatorrules.RecordingRule {
	return []operatorrules.RecordingRule{
		{
			MetricsOpts: operatormetrics.MetricOpts{
				Name: "cluster:kubevirt_cnao_kubemacpool_manager_up:sum",
				Help: "[Deprecated] The number of KubeMacPool manager pods that are up",
			},
			MetricType: operatormetrics.GaugeType,
			Expr:       intstr.FromString(fmt.Sprintf("sum(up{namespace='%s', pod=~'kubemacpool-mac-controller-manager-.*'} or vector(0))", namespace)),
		},
		{
			MetricsOpts: operatormetrics.MetricOpts{
				Name: "cluster:kubevirt_cnao_cr_kubemacpool_deployed:sum",
				Help: "[Deprecated] The number of KubeMacPool manager pods deployed by CNAO CR",
			},
			MetricType: operatormetrics.GaugeType,
			Expr:       intstr.FromString(fmt.Sprintf("sum(kubevirt_cnao_cr_kubemacpool_deployed{namespace='%s'} or vector(0))", namespace)),
		},
	}
}
