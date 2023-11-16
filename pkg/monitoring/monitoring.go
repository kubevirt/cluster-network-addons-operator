package monitoring

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/kubevirt/cluster-network-addons-operator/pkg/render"
)

const (
	defaultMonitoringNamespace = "monitoring"
	defaultServiceAccountName  = "prometheus-k8s"
	defaultRunbookURLTemplate  = "https://kubevirt.io/monitoring/runbooks/"
	runbookURLTemplateEnv      = "RUNBOOK_URL_TEMPLATE"
)

func RenderMonitoring(manifestDir string, monitoringAvailable bool) ([]*unstructured.Unstructured, error) {
	if !monitoringAvailable {
		return nil, nil
	}

	// render the manifests on disk
	data := render.MakeRenderData()
	data.Data["Namespace"] = os.Getenv("OPERAND_NAMESPACE")
	data.Data["MonitoringNamespace"] = getNamespace()
	data.Data["MonitoringServiceAccount"] = getServiceAccount()
	data.Data["RunbookURLTemplate"] = GetRunbookURLTemplate()

	objs, err := render.RenderDir(filepath.Join(manifestDir, "monitoring"), &data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to render monitoring manifests")
	}

	return objs, nil
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

func GetRunbookURLTemplate() string {
	runbookURLTemplate, exists := os.LookupEnv(runbookURLTemplateEnv)
	if !exists {
		runbookURLTemplate = defaultRunbookURLTemplate
	}

	return runbookURLTemplate
}
