package monitoring

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/metrics"

	"github.com/kubevirt/cluster-network-addons-operator/pkg/render"
)

const (
	defaultMetricPort          = 8080
	defaultMonitoringNamespace = "monitoring"
	defaultServiceAccountName  = "prometheus-k8s"
)

func init() {
	metrics.Registry.MustRegister()
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
