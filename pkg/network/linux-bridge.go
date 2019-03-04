package network

import (
	"github.com/phoracek/cluster-network-addons-operator/pkg/render"
	"github.com/pkg/errors"
	uns "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"os"
	"path/filepath"

	networkaddonsoperatorv1alpha1 "github.com/phoracek/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/v1alpha1"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func validateLinuxBridge(conf *networkaddonsoperatorv1alpha1.NetworkAddonsConfigSpec) []error {

	return []error{}
}

// renderLinuxBridge generates the manifests of Linux Bridge
func renderLinuxBridge(conf *networkaddonsoperatorv1alpha1.NetworkAddonsConfigSpec, manifestDir string) ([]*unstructured.Unstructured, error) {
	if conf.LinuxBridge == nil {
		return nil, nil
	}

	var err error
	out := []*unstructured.Unstructured{}
	objs := []*unstructured.Unstructured{}

	objs, err = renderLinuxBridgeConfig(conf.Multus.Delegates, manifestDir)
	if err != nil {
		return nil, err
	}
	out = append(out, objs...)

	return out, nil
}

// renderLinuxBridgeConfig returns the manifests of Multus
func renderLinuxBridgeConfig(delegates string, manifestDir string) ([]*uns.Unstructured, error) {
	objs := []*uns.Unstructured{}

	// render the manifests on disk
	data := render.MakeRenderData()
	data.Data["LinuxBridgeImage"] = os.Getenv("LINUX_BRIDGE_IMAGE")
	data.Data["ImagePullPolicy"] = os.Getenv("IMAGE_PULL_POLICY")

	manifests, err := render.RenderDir(filepath.Join(manifestDir, "linux-bridge"), &data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to render linux-bridge manifests")
	}
	objs = append(objs, manifests...)
	return objs, nil
}
