package network

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/kubevirt/cluster-network-addons-operator/pkg/render"
)

func RenderNetworkPolicy(manifestDir string, info ClusterInfo) ([]*unstructured.Unstructured, error) {
	data := render.MakeRenderData()
	data.Data["Namespace"] = os.Getenv("OPERAND_NAMESPACE")
	if info.OpenShift4 {
		data.Data["ClusterDNSNamespace"] = "openshift-dns"
		data.Data["ClusterDNSPodsSelectorKey"] = "dns.operator.openshift.io/daemonset-dns"
		data.Data["ClusterDNSPodsSelectorValue"] = "default"
		data.Data["ClusterDNSPort"] = "5353"
	} else {
		data.Data["ClusterDNSNamespace"] = "kube-system"
		data.Data["ClusterDNSPodsSelectorKey"] = "k8s-app"
		data.Data["ClusterDNSPodsSelectorValue"] = "kube-dns"
		data.Data["ClusterDNSPort"] = "53"
	}

	objs, err := render.RenderDir(filepath.Join(manifestDir, "network-policy"), &data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to render network-policy manifests")
	}
	return objs, nil
}
