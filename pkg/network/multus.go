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

// TODO: render default network
// TODO: use manifests from multus-cni
// TODO: make sure that multus will always take the first place

// ValidateMultus validates the combination of DisableMultiNetwork and AddtionalNetworks
func validateMultus(conf *networkaddonsoperatorv1alpha1.NetworkAddonsConfigSpec) []error {
	// TODO config must be there (maybe in crd def)

	return []error{}
}

// RenderMultus generates the manifests of Multus
func renderMultus(conf *networkaddonsoperatorv1alpha1.NetworkAddonsConfigSpec, manifestDir string) ([]*unstructured.Unstructured, error) {
	if conf.Multus == nil {
		return nil, nil
	}

	var err error
	out := []*unstructured.Unstructured{}
	objs := []*unstructured.Unstructured{}

	objs, err = renderMultusConfig(conf.Multus.Delegates, manifestDir)
	if err != nil {
		return nil, err
	}
	out = append(out, objs...)

	return out, nil
}

// renderMultusConfig returns the manifests of Multus
func renderMultusConfig(delegates string, manifestDir string) ([]*uns.Unstructured, error) {
	objs := []*uns.Unstructured{}

	// render the manifests on disk
	data := render.MakeRenderData()
	data.Data["MultusImage"] = os.Getenv("MULTUS_IMAGE")
	data.Data["MultusDelegates"] = delegates

	manifests, err := render.RenderDir(filepath.Join(manifestDir, "multus"), &data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to render multus manifests")
	}
	objs = append(objs, manifests...)
	return objs, nil
}

// TODO: validate multus change - test if it is possible to upgrade from one
// default net to another
