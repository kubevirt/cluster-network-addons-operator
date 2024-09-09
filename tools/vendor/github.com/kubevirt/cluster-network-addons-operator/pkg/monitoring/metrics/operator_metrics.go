package metrics

import (
	"github.com/machadovilaca/operator-observability/pkg/operatormetrics"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/controller/statusmanager"
)

var (
	operatorMetrics = []operatormetrics.Metric{
		cnaoCrReady,
		kmpDeployed,
	}

	cnaoCrReady = operatormetrics.NewGauge(
		operatormetrics.MetricOpts{
			Name: "kubevirt_cnao_cr_ready",
			Help: "CNAO CR Ready",
		},
	)

	kmpDeployed = operatormetrics.NewGauge(
		operatormetrics.MetricOpts{
			Name: "kubevirt_cnao_cr_kubemacpool_deployed",
			Help: "KubeMacpool is deployed by CNAO CR",
		},
	)
)

func ResetMonitoredComponents() {
	setGaugeParam(false, cnaoCrReady)
	setGaugeParam(false, kmpDeployed)
}

func TrackMonitoredComponents(conf *cnao.NetworkAddonsConfigSpec, statusManager *statusmanager.StatusManager) {
	isKubemacpoolDeployed := conf.KubeMacPool != nil
	setGaugeParam(isKubemacpoolDeployed, kmpDeployed)
	setGaugeParam(statusManager.IsStatusAvailable(), cnaoCrReady)
}

func setGaugeParam(setTrueFlag bool, gaugeParam *operatormetrics.Gauge) {
	if setTrueFlag {
		gaugeParam.Set(1)
	} else {
		gaugeParam.Set(0)
	}
}
