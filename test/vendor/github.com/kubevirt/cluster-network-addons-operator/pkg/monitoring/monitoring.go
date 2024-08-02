package monitoring

import (
	"os"
	"path/filepath"

	"github.com/kubevirt/cluster-network-addons-operator/pkg/monitoring/rules"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/kubevirt/cluster-network-addons-operator/pkg/render"
)

const (
	defaultMonitoringNamespace = "monitoring"
	defaultServiceAccountName  = "prometheus-k8s"
)

func RenderMonitoring(manifestDir string, monitoringAvailable bool) ([]*unstructured.Unstructured, error) {
	if !monitoringAvailable {
		return nil, nil
	}

	operandNamespace := os.Getenv("OPERAND_NAMESPACE")
	monitoringNamespace := getMonitoringNamespace()

	// render the manifests on disk
	data := render.MakeRenderData()
	data.Data["Namespace"] = operandNamespace
	data.Data["MonitoringNamespace"] = monitoringNamespace
	data.Data["MonitoringServiceAccount"] = getServiceAccount()

	objs, err := render.RenderDir(filepath.Join(manifestDir, "monitoring"), &data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to render monitoring manifests")
	}

	if err := rules.SetupRules(operandNamespace); err != nil {
		return nil, errors.Wrap(err, "failed to setup monitoring rules")
	}

	promRule, err := rules.BuildPrometheusRule(operandNamespace)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build PrometheusRule")
	}

	unstructuredPromRule, err := runtime.DefaultUnstructuredConverter.ToUnstructured(promRule)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert PrometheusRule to unstructured")
	}
	objs = append(objs, &unstructured.Unstructured{Object: unstructuredPromRule})

	return objs, nil
}

func getMonitoringNamespace() string {
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
