package network

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/network/cni"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/render"
)

// renderMultusDynamicNetworksController generates the manifests of multus-dynamic-networks-controller
func renderMultusDynamicNetworksController(conf *cnao.NetworkAddonsConfigSpec, manifestDir string, clusterInfo *ClusterInfo) ([]*unstructured.Unstructured, error) {
	if conf.MultusDynamicNetworksController == nil {
		return nil, nil
	}

	// render the manifests on disk
	data := render.MakeRenderData()
	data.Data["Namespace"] = os.Getenv("OPERAND_NAMESPACE")
	data.Data["ImagePullPolicy"] = conf.ImagePullPolicy
	data.Data["EnableSCC"] = clusterInfo.SCCAvailable
	data.Data["MultusDynamicNetworksControllerImage"] = os.Getenv("MULTUS_DYNAMIC_NETWORKS_CONTROLLER_IMAGE")
	if clusterInfo.OpenShift4 {
		data.Data["CniMountPath"] = cni.BinDirOpenShift4
	} else {
		data.Data["CniMountPath"] = cni.BinDir
	}
	data.Data["Placement"] = conf.PlacementConfiguration.Workloads
	objs, err := render.RenderDir(filepath.Join(manifestDir, "multus-dynamic-networks-controller"), &data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to render multus-dynamic-networks-controller state handler manifests")
	}

	return objs, nil
}
