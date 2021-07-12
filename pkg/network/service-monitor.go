package network

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/kubevirt/cluster-network-addons-operator/pkg/render"
)

func renderMonitoring(manifestDir string, clusterInfo *ClusterInfo) ([]*unstructured.Unstructured, error) {
	if !clusterInfo.PrometheusDeployed {
		return nil, nil
	}

	// render the manifests on disk
	data := render.MakeRenderData()
	data.Data["Namespace"] = os.Getenv("OPERAND_NAMESPACE")
	data.Data["MonitoringNamespace"] = os.Getenv("MONITORING_NAMESPACE")

	objs, err := render.RenderDir(filepath.Join(manifestDir, "monitoring"), &data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to render monitoring manifests")
	}

	return objs, nil
}
