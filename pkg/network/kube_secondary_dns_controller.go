package network

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/render"
)

// renderKubeSecondaryDNS generates the manifests of kube-secondary-dns
func renderKubeSecondaryDNS(conf *cnao.NetworkAddonsConfigSpec, manifestDir string, clusterInfo *ClusterInfo) ([]*unstructured.Unstructured, error) {
	if conf.KubeSecondaryDNS == nil {
		return nil, nil
	}

	// render the manifests on disk
	data := render.MakeRenderData()
	data.Data["Namespace"] = os.Getenv("OPERAND_NAMESPACE")
	data.Data["ImagePullPolicy"] = conf.ImagePullPolicy
	data.Data["Domain"] = conf.KubeSecondaryDNS.Domain
	data.Data["NameServerIp"] = conf.KubeSecondaryDNS.NameServerIP
	data.Data["KubeSecondaryDNSImage"] = os.Getenv("KUBE_SECONDARY_DNS_IMAGE")
	data.Data["CoreDNSImage"] = os.Getenv("CORE_DNS_IMAGE")

	objs, err := render.RenderDir(filepath.Join(manifestDir, "kube-secondary-dns"), &data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to render kube-secondary-dns state handler manifests")
	}

	return objs, nil
}
