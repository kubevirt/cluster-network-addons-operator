package rules

import (
	"github.com/kubevirt/cluster-network-addons-operator/pkg/monitoring/rules/alerts"
	"github.com/machadovilaca/operator-observability/pkg/operatorrules"
	promv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"

	"github.com/kubevirt/cluster-network-addons-operator/pkg/monitoring/rules/recordingrules"
)

const prometheusRuleName = "prometheus-rules-cluster-network-addons-operator"

func SetupRules(namespace string) error {
	if err := recordingrules.Register(namespace); err != nil {
		return err
	}

	if err := alerts.Register(namespace); err != nil {
		return err
	}

	return nil
}

func BuildPrometheusRule(namespace string) (*promv1.PrometheusRule, error) {
	rules, err := operatorrules.BuildPrometheusRule(
		prometheusRuleName,
		namespace,
		map[string]string{
			"prometheus.cnao.io": "true",
		},
	)
	if err != nil {
		return nil, err
	}

	return rules, nil
}

func ListRecordingRules() []operatorrules.RecordingRule {
	return operatorrules.ListRecordingRules()
}

func ListAlerts() []promv1.Rule {
	return operatorrules.ListAlerts()
}
