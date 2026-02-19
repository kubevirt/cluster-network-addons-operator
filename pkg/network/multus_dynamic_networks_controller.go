package network

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	openshiftoperatorv1 "github.com/openshift/api/operator/v1"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/network/cni"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/render"
)

func fillDefaultsMultusDynamicNetworks(conf, previous *cnao.NetworkAddonsConfigSpec) []error {
	if conf.MultusDynamicNetworks == nil {
		return []error{}
	}

	// If user hasn't explicitly set host cri socket path, set it to default value
	if conf.MultusDynamicNetworks.HostCRISocketPath == "" {
		conf.MultusDynamicNetworks.HostCRISocketPath = "/run/crio/crio.sock"
	}

	return []error{}
}

func changeSafeMultusDynamicNetworks(prev, next *cnao.NetworkAddonsConfigSpec) []error {
	return []error{}
}

// renderMultusDynamicNetworks generates the manifests of multus-dynamic-networks-controller
func renderMultusDynamicNetworks(conf *cnao.NetworkAddonsConfigSpec, manifestDir string, clusterInfo *ClusterInfo) ([]*unstructured.Unstructured, error) {
	if conf.MultusDynamicNetworks == nil {
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
	data.Data["HostCRISocketPath"] = conf.MultusDynamicNetworks.HostCRISocketPath
	objs, err := render.RenderDir(filepath.Join(manifestDir, "multus-dynamic-networks-controller"), &data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to render multus-dynamic-networks-controller state handler manifests")
	}

	return objs, nil
}

func validateMultusDynamicNetworks(conf *cnao.NetworkAddonsConfigSpec, openshiftNetworkConfig *openshiftoperatorv1.Network) []error {
	if conf.MultusDynamicNetworks == nil {
		return nil
	}

	if openshiftNetworkConfig != nil {
		return []error{fmt.Errorf("`multusDynamicNetworks` configuration is not supported on Openshift yet")}
	}

	if conf.Multus == nil {
		return []error{fmt.Errorf("the `multus` configuration is required")}
	}
	return nil
}
