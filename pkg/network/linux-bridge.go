package network

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/kubevirt/cluster-network-addons-operator/pkg/render"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/network/cni"
)

func validateLinuxBridge(conf *cnao.NetworkAddonsConfigSpec) []error {
	if conf.LinuxBridge == nil || conf.LinuxBridge.BridgeMarkerHealthPort == nil {
		return nil
	}
	port := *conf.LinuxBridge.BridgeMarkerHealthPort
	if port < 1 || port > 65535 {
		return []error{fmt.Errorf("linuxBridge.bridgeMarkerHealthPort %d is out of range [1, 65535]", port)}
	}
	return nil
}

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
	if conf.LinuxBridge.BridgeMarkerHealthPort != nil {
		data.Data["HealthPort"] = strconv.Itoa(int(*conf.LinuxBridge.BridgeMarkerHealthPort))
	} else {
		data.Data["HealthPort"] = cni.DefaultBridgeMarkerHealthPort
	}

	objs, err := render.RenderDir(filepath.Join(manifestDir, "linux-bridge"), &data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to render linux-bridge manifests")
	}

	return objs, nil
}
