package network

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/kubevirt/cluster-network-addons-operator/pkg/render"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	opv1alpha1 "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/v1alpha1"
)

func changeSafeSriov(prev, next *opv1alpha1.NetworkAddonsConfigSpec) []error {
	if prev.Sriov != nil && !reflect.DeepEqual(prev.Sriov, next.Sriov) {
		return []error{errors.Errorf("cannot modify Sriov configuration once it is deployed")}
	}
	return nil
}

func getRootDevicesConfigString(rootDevices string) string {
	devices := make([]string, 0)
	for _, id := range strings.Split(rootDevices, ",") {
		if id != "" {
			devices = append(devices, fmt.Sprintf("\"%s\"", id))
		}
	}
	return strings.Join(devices, ",")
}

// renderSriov generates the manifests of SR-IOV plugins
func renderSriov(conf *opv1alpha1.NetworkAddonsConfigSpec, manifestDir string, enableSCC bool) ([]*unstructured.Unstructured, error) {
	if conf.Sriov == nil {
		return nil, nil
	}

	// render the manifests on disk
	data := render.MakeRenderData()
	data.Data["SriovRootDevices"] = getRootDevicesConfigString(os.Getenv("SRIOV_ROOT_DEVICES"))
	data.Data["SriovDpImage"] = os.Getenv("SRIOV_DP_IMAGE")
	data.Data["SriovCniImage"] = os.Getenv("SRIOV_CNI_IMAGE")
	data.Data["ImagePullPolicy"] = os.Getenv("IMAGE_PULL_POLICY")
	data.Data["EnableSCC"] = enableSCC

	objs, err := render.RenderDir(filepath.Join(manifestDir, "sriov"), &data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to render sriov manifests")
	}

	return objs, nil
}
