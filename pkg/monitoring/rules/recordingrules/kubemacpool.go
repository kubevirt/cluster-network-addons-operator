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
				Name: "kubevirt_cnao_kubemacpool_manager_up",
				Help: "[Deprecated] Total count of running KubeMacPool manager pods",
			},
			MetricType: operatormetrics.GaugeType,
			Expr:       intstr.FromString(fmt.Sprintf("sum(up{namespace='%s', pod=~'kubemacpool-mac-controller-manager-.*'} or vector(0))", namespace)),
		},
		{
			MetricsOpts: operatormetrics.MetricOpts{
				Name: "cluster:kubevirt_cnao_kubemacpool_manager_up:sum",
				Help: "The number of KubeMacPool manager pods that are up",
			},
			MetricType: operatormetrics.GaugeType,
			Expr:       intstr.FromString(fmt.Sprintf("sum(up{namespace='%s', pod=~'kubemacpool-mac-controller-manager-.*'} or vector(0))", namespace)),
		},
		{
			MetricsOpts: operatormetrics.MetricOpts{
				Name: "kubevirt_cnao_cr_kubemacpool_aggregated",
				Help: "[Deprecated] Total count of KubeMacPool manager pods deployed by CNAO CR",
			},
			MetricType: operatormetrics.GaugeType,
			Expr:       intstr.FromString(fmt.Sprintf("sum(kubevirt_cnao_cr_kubemacpool_deployed{namespace='%s'} or vector(0))", namespace)),
		},
		{
			MetricsOpts: operatormetrics.MetricOpts{
				Name: "cluster:kubevirt_cnao_cr_kubemacpool_deployed:sum",
				Help: "The number of KubeMacPool manager pods deployed by CNAO CR",
			},
			MetricType: operatormetrics.GaugeType,
			Expr:       intstr.FromString(fmt.Sprintf("sum(kubevirt_cnao_cr_kubemacpool_deployed{namespace='%s'} or vector(0))", namespace)),
		},
		{
			MetricsOpts: operatormetrics.MetricOpts{
				Name: "kubevirt_cnao_kubemacpool_duplicate_macs",
				Help: "[DEPRECATED] Total count of duplicate KubeMacPool MAC addresses. This recording rule monitors VM MACs instead of running VMI MACs and will be removed in the next minor release. Use KubeMacPool's native VMI collision detection instead",
			},
			MetricType: operatormetrics.GaugeType,
			Expr:       intstr.FromString(fmt.Sprintf("sum(kubevirt_kmp_duplicate_macs{namespace='%s'} or vector(0))", namespace)),
		},
	}
}
