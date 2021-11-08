package network

import (
	"os"
	"path/filepath"

	"github.com/kubevirt/cluster-network-addons-operator/pkg/render"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/network/cni"
)

// renderLinuxBridge generates the manifests of Linux Bridge
func renderLinuxBridge(conf *cnao.NetworkAddonsConfigSpec, manifestDir string, clusterInfo *ClusterInfo) ([]*unstructured.Unstructured, error) {
	if conf.LinuxBridge == nil {
		return nil, nil
	}

	// render the manifests on disk
	data := render.MakeRenderData()
	data.Data["Namespace"] = os.Getenv("OPERAND_NAMESPACE")
	data.Data["LinuxBridgeMarkerImage"] = os.Getenv("LINUX_BRIDGE_MARKER_IMAGE")
	data.Data["LinuxBridgeImage"] = os.Getenv("LINUX_BRIDGE_IMAGE")
	data.Data["ImagePullPolicy"] = conf.ImagePullPolicy
	if clusterInfo.OpenShift4 {
		data.Data["CNIBinDir"] = cni.BinDirOpenShift4
	} else {
		data.Data["CNIBinDir"] = cni.BinDir
	}
	data.Data["EnableSCC"] = clusterInfo.SCCAvailable
	data.Data["Placement"] = conf.PlacementConfiguration.Workloads
	data.Data["PodSecurityEnabled"] = clusterInfo.PodSecurityEnabled

	objs, err := render.RenderDir(filepath.Join(manifestDir, "linux-bridge"), &data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to render linux-bridge manifests")
	}

	return objs, nil
}
