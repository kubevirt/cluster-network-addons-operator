package monitoring

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/metrics"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/controller/statusmanager"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/render"
)

const (
	defaultMetricPort          = 8080
	defaultMonitoringNamespace = "monitoring"
	defaultServiceAccountName  = "prometheus-k8s"
)

var (
	readyGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "kubevirt_cnao_cr_ready",
			Help: "Cnao CR Ready",
		})
	kmpDeployedGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "kubevirt_cnao_cr_kubemacpool_deployed",
			Help: "Kubemacpool is deployed by Cnao CR",
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

func RenderMonitoring(manifestDir string, monitoringAvailable bool) ([]*unstructured.Unstructured, error) {
	if !monitoringAvailable {
		return nil, nil
	}

	// render the manifests on disk
	data := render.MakeRenderData()
	data.Data["Namespace"] = os.Getenv("OPERAND_NAMESPACE")
	data.Data["MonitoringNamespace"] = getNamespace()
	data.Data["MonitoringServiceAccount"] = getServiceAccount()

	objs, err := render.RenderDir(filepath.Join(manifestDir, "monitoring"), &data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to render monitoring manifests")
	}

	return objs, nil
}

func GetMetricsAddress() string {
	return metrics.DefaultBindAddress
}

func GetMetricsPort() int32 {
	re := regexp.MustCompile(`(?m).*:(\d+)`)

	portString := re.ReplaceAllString(metrics.DefaultBindAddress, "$1")
	portInt64, err := strconv.ParseUint(portString, 10, 32)
	if err != nil {
		return defaultMetricPort
	}
	return int32(portInt64)
}

func getNamespace() string {
	monitoringNamespaceFromEnv := os.Getenv("MONITORING_NAMESPACE")

	if monitoringNamespaceFromEnv != "" {
		return monitoringNamespaceFromEnv
	}
	return defaultMonitoringNamespace
}

func getServiceAccount() string {
	monitoringServiceAccountFromEnv := os.Getenv("MONITORING_SERVICE_ACCOUNT")

	if monitoringServiceAccountFromEnv != "" {
		return monitoringServiceAccountFromEnv
	}
	return defaultServiceAccountName
}
