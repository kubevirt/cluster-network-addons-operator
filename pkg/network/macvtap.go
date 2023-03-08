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

// renderMacvtapCni generates the manifests of macvtap-cni handler
func renderMacvtapCni(conf *cnao.NetworkAddonsConfigSpec, manifestDir string, clusterInfo *ClusterInfo) ([]*unstructured.Unstructured, error) {
	if conf.MacvtapCni == nil {
		return nil, nil
	}

	// render the manifests on disk
	data := render.MakeRenderData()
	data.Data["Namespace"] = os.Getenv("OPERAND_NAMESPACE")
	data.Data["ImagePullPolicy"] = conf.ImagePullPolicy
	data.Data["EnableSCC"] = clusterInfo.SCCAvailable
	data.Data["MacvtapImage"] = os.Getenv("MACVTAP_CNI_IMAGE")
	data.Data["DevicePluginConfigName"] = conf.MacvtapCni.DevicePluginConfig
	if clusterInfo.OpenShift4 {
		data.Data["CniMountPath"] = cni.BinDirOpenShift4
	} else {
		data.Data["CniMountPath"] = cni.BinDir
	}
	data.Data["Placement"] = conf.PlacementConfiguration.Workloads
	objs, err := render.RenderDir(filepath.Join(manifestDir, "macvtap"), &data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to render macvtap-cni state handler manifests")
	}

	return objs, nil
}

func fillMacvtapDefaults(conf *cnao.NetworkAddonsConfigSpec) {
	// https://github.com/kubevirt/macvtap-cni/blob/be1528fb09e9ac3c490a5df31330851d7e1f8b0a/manifests/macvtap.yaml#L23
	const defaultMacvtapDevicePluginConfigMapName = "macvtap-deviceplugin-config"
	if conf.MacvtapCni != nil && conf.MacvtapCni.DevicePluginConfig == "" {
		conf.MacvtapCni.DevicePluginConfig = defaultMacvtapDevicePluginConfigMapName
	}
}
