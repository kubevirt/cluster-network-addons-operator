package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/controller/statusmanager"
)

type MetricsOpts struct {
	Name string
	Help string
	Type string
}

type MetricsKey string

const (
	ReadyGauge     MetricsKey = "readyGauge"
	KMPDeployGauge MetricsKey = "kmpDeployedGauge"
)

var MetricsOptsList = map[MetricsKey]MetricsOpts{
	ReadyGauge: {
		Name: "kubevirt_cnao_cr_ready",
		Help: "CNAO CR Ready",
		Type: "Gauge",
	},
	KMPDeployGauge: {
		Name: "kubevirt_cnao_cr_kubemacpool_deployed",
		Help: "KubeMacpool is deployed by CNAO CR",
		Type: "Gauge",
	},
}

var (
	readyGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: MetricsOptsList[ReadyGauge].Name,
			Help: MetricsOptsList[ReadyGauge].Help,
		})
	kmpDeployedGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: MetricsOptsList[KMPDeployGauge].Name,
			Help: MetricsOptsList[KMPDeployGauge].Help,
		})
)

func init() {
	metrics.Registry.MustRegister(readyGauge, kmpDeployedGauge)
}

func setGaugeParam(setTrueFlag bool, gaugeParam *prometheus.Gauge) {
	if setTrueFlag {
		(*gaugeParam).Set(1)
	} else {
		(*gaugeParam).Set(0)
	}
}

func ResetMonitoredComponents() {
	setGaugeParam(false, &readyGauge)
	setGaugeParam(false, &kmpDeployedGauge)
}

func TrackMonitoredComponents(conf *cnao.NetworkAddonsConfigSpec, statusManager *statusmanager.StatusManager) {
	isKubemacpoolDeployed := conf.KubeMacPool != nil
	setGaugeParam(isKubemacpoolDeployed, &kmpDeployedGauge)
	setGaugeParam(statusManager.IsStatusAvailable(), &readyGauge)
}

func GetMetricsAddress() string {
	return metrics.DefaultBindAddress
}
