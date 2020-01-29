package network

import (
	"log"
	"os"
	"path/filepath"
	"reflect"

	"github.com/kubevirt/cluster-network-addons-operator/pkg/render"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	opv1alpha1 "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/v1alpha1"
)

func changeSafeNMState(prev, next *opv1alpha1.NetworkAddonsConfigSpec) []error {
	if prev.NMState != nil && !reflect.DeepEqual(prev.NMState, next.NMState) {
		if !(prev.NMState != nil && next.NMState == nil) {
		return []error{errors.Errorf("cannot modify NMState state handler configuration once it is deployed")}
		}
		log.Printf("DBGDBG Detected that NMState was removed")
	}

	return nil
}

// renderNMState generates the manifests of NMState handler
func renderNMState(prev, conf *opv1alpha1.NetworkAddonsConfigSpec, manifestDir string, clusterInfo *ClusterInfo) ([]*unstructured.Unstructured, []*unstructured.Unstructured, error) {
	if conf.NMState == nil && prev.NMState == nil {
	//if conf.NMState == nil {
		return nil, nil, nil
	}

	// render the manifests on disk
	data := render.MakeRenderData()
	data.Data["Namespace"] = os.Getenv("OPERAND_NAMESPACE")
	data.Data["NMStateHandlerImage"] = os.Getenv("NMSTATE_HANDLER_IMAGE")
	data.Data["ImagePullPolicy"] = conf.ImagePullPolicy
	data.Data["EnableSCC"] = clusterInfo.SCCAvailable

	objs, err := render.RenderDir(filepath.Join(manifestDir, "nmstate"), &data)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to render nmstate state handler manifests")
	}

	if conf.NMState == nil && prev.NMState != nil {
		return nil, objs, nil
	}
	return objs, nil, nil
}
