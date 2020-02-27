package network

import (
	"os"
	"path/filepath"

	"github.com/kubevirt/cluster-network-addons-operator/pkg/render"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	opv1alpha1 "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/v1alpha1"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/network/cni"
)

// renderOvs generates the manifests of Ovs
func renderOvs(conf *opv1alpha1.NetworkAddonsConfigSpec, manifestDir string, clusterInfo *ClusterInfo) ([]*unstructured.Unstructured, error) {
	if conf.Ovs == nil {
		return nil, nil
	}

	// render the manifests on disk
	data := render.MakeRenderData()
	data.Data["Namespace"] = os.Getenv("OPERAND_NAMESPACE")
	data.Data["OvsCNIImage"] = os.Getenv("OVS_CNI_IMAGE")
	data.Data["OvsMarkerImage"] = os.Getenv("OVS_MARKER_IMAGE")
	data.Data["OvsImage"] = os.Getenv("OVS_IMAGE")
	data.Data["ImagePullPolicy"] = conf.ImagePullPolicy
	if clusterInfo.OpenShift4 {
		data.Data["CNIBinDir"] = cni.BinDirOpenShift4
	} else {
		data.Data["CNIBinDir"] = cni.BinDir
	}
	data.Data["EnableSCC"] = clusterInfo.SCCAvailable

	objs, err := render.RenderDir(filepath.Join(manifestDir, "ovs"), &data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to render ovs manifests")
	}

	return objs, nil
}
