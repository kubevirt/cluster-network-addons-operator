package network

import (
	"context"
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/kubevirt/cluster-network-addons-operator/pkg/monitoring"

	osv1 "github.com/openshift/api/operator/v1"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
)

// Canonicalize converts configuration to a canonical form.
func Canonicalize(conf *cnao.NetworkAddonsConfigSpec) {
	// TODO
}

// Validate checks that the supplied configuration is reasonable.
// This should be called after Canonicalize
func Validate(conf *cnao.NetworkAddonsConfigSpec, openshiftNetworkConfig *osv1.Network) error {
	errs := []error{}

	errs = append(errs, validateMultus(conf, openshiftNetworkConfig)...)
	errs = append(errs, validateKubeMacPool(conf)...)
	errs = append(errs, validateImagePullPolicy(conf)...)
	errs = append(errs, validateSelfSignConfiguration(conf)...)

	if len(errs) > 0 {
		return errors.Errorf("invalid configuration:\n%s", errorListToMultiLineString(errs))
	}
	return nil
}

// FillDefaults computes any default values and applies them to the configuration
// This is a mutating operation. It should be called after Validate.
//
// Defaults are carried forward from previous if it is provided. This is so we
// can change defaults as we move forward, but won't disrupt existing clusters.
func FillDefaults(conf, previous *cnao.NetworkAddonsConfigSpec) error {
	errs := []error{}

	errs = append(errs, fillDefaultsPlacementConfiguration(conf, previous)...)
	errs = append(errs, fillDefaultsSelfSignConfiguration(conf, previous)...)
	errs = append(errs, fillDefaultsImagePullPolicy(conf, previous)...)
	errs = append(errs, fillDefaultsKubeMacPool(conf, previous)...)

	if len(errs) > 0 {
		return errors.Errorf("invalid configuration:\n%s", errorListToMultiLineString(errs))
	}

	return nil
}

// specialCleanUp checks if there are any specific outdated objects or ones that are no longer compatible and deletes them.
func SpecialCleanUp(conf *cnao.NetworkAddonsConfigSpec, client k8sclient.Client, clusterInfo *ClusterInfo) error {
	errs := []error{}
	ctx := context.TODO()

	errs = append(errs, cleanUpMultus(conf, ctx, client)...)
	errs = append(errs, cleanUpNMState(conf, ctx, client, clusterInfo)...)

	if len(errs) > 0 {
		return errors.Errorf("invalid configuration:\n%v", errorListToMultiLineString(errs))
	}

	return nil
}

// IsChangeSafe checks to see if the change between prev and next are allowed
// FillDefaults and Validate should have been called.
func IsChangeSafe(prev, next *cnao.NetworkAddonsConfigSpec) error {
	if prev == nil {
		return nil
	}

	// Easy way out: nothing changed.
	if reflect.DeepEqual(prev, next) {
		return nil
	}

	errs := []error{}

	errs = append(errs, changeSafeKubeMacPool(prev, next)...)
	errs = append(errs, changeSafeImagePullPolicy(prev, next)...)

	if len(errs) > 0 {
		return errors.Errorf("invalid configuration:\n%s", errorListToMultiLineString(errs))
	}
	return nil
}

// Render creates a list of components to be created
func Render(conf *cnao.NetworkAddonsConfigSpec, manifestDir string, openshiftNetworkConfig *osv1.Network, clusterInfo *ClusterInfo) ([]*unstructured.Unstructured, error) {
	log.Print("starting render phase")
	objs := []*unstructured.Unstructured{}

	// render Multus
	o, err := renderMultus(conf, manifestDir, openshiftNetworkConfig, clusterInfo)
	if err != nil {
		return nil, err
	}
	objs = append(objs, o...)

	// render Linux Bridge
	o, err = renderLinuxBridge(conf, manifestDir, clusterInfo)
	if err != nil {
		return nil, err
	}
	objs = append(objs, o...)

	// render kubeMacPool
	o, err = renderKubeMacPool(conf, manifestDir)
	if err != nil {
		return nil, err
	}
	objs = append(objs, o...)

	// render NMState
	o, err = renderNMState(conf, manifestDir, clusterInfo)
	if err != nil {
		return nil, err
	}
	objs = append(objs, o...)

	// render Ovs
	o, err = renderOvs(conf, manifestDir, clusterInfo)
	if err != nil {
		return nil, err
	}
	objs = append(objs, o...)

	// render MacvtapCni
	o, err = renderMacvtapCni(conf, manifestDir, clusterInfo)
	if err != nil {
		return nil, err
	}
	objs = append(objs, o...)

	// render Monitoring Service
	o, err = monitoring.RenderMonitoring(manifestDir, clusterInfo.MonitoringAvailable)
	if err != nil {
		return nil, err
	}
	objs = append(objs, o...)

	log.Printf("render phase done, rendered %d objects", len(objs))
	return objs, nil
}

// RenderObjsToRemove creates list of components to be removed
func RenderObjsToRemove(prev, conf *cnao.NetworkAddonsConfigSpec, manifestDir string, openshiftNetworkConfig *osv1.Network, clusterInfo *ClusterInfo) ([]*unstructured.Unstructured, error) {
	log.Print("starting rendering objects to delete phase")
	objsToRemove := []*unstructured.Unstructured{}

	if prev == nil {
		return nil, nil
	}

	if conf.Multus == nil {
		o, err := renderMultus(prev, manifestDir, openshiftNetworkConfig, clusterInfo)
		if err != nil {
			return nil, err
		}
		objsToRemove = append(objsToRemove, o...)
	}

	if conf.LinuxBridge == nil {
		o, err := renderLinuxBridge(prev, manifestDir, clusterInfo)
		if err != nil {
			return nil, err
		}
		objsToRemove = append(objsToRemove, o...)
	}

	if conf.KubeMacPool == nil {
		o, err := renderKubeMacPool(prev, manifestDir)
		if err != nil {
			return nil, err
		}
		objsToRemove = append(objsToRemove, o...)
	}

	if conf.NMState == nil {
		o, err := renderNMState(prev, manifestDir, clusterInfo)
		if err != nil {
			return nil, err
		}
		objsToRemove = append(objsToRemove, o...)
	}

	if conf.Ovs == nil {
		o, err := renderOvs(prev, manifestDir, clusterInfo)
		if err != nil {
			return nil, err
		}
		objsToRemove = append(objsToRemove, o...)
	}

	// render MacvtapCni
	if conf.MacvtapCni == nil {
		o, err := renderMacvtapCni(conf, manifestDir, clusterInfo)
		if err != nil {
			return nil, err
		}
		objsToRemove = append(objsToRemove, o...)
	}

	// Remove OPERAND_NAMESPACE occurences
	// TODO cleanup OPERAND_NAMESPACE once there are no components using it.
	objsToRemoveWithoutNamespace := []*unstructured.Unstructured{}
	operandNamespace := os.Getenv("OPERAND_NAMESPACE")
	for _, obj := range objsToRemove {
		if !(obj.GetName() == operandNamespace && obj.GetKind() == "Namespace") {
			objsToRemoveWithoutNamespace = append(objsToRemoveWithoutNamespace, obj)
		}
	}
	objsToRemove = objsToRemoveWithoutNamespace

	// Do not remove CustomResourceDefinitions, they should be kept even after
	// removal of the operator
	objsToRemoveWithoutCRDs := []*unstructured.Unstructured{}
	for _, obj := range objsToRemove {
		if obj.GetKind() != "CustomResourceDefinition" {
			objsToRemoveWithoutCRDs = append(objsToRemoveWithoutCRDs, obj)
		}
	}
	objsToRemove = objsToRemoveWithoutCRDs

	log.Printf("object removal render phase done, rendered %d objects to remove", len(objsToRemove))
	return objsToRemove, nil
}

func errorListToMultiLineString(errs []error) string {
	stringErrs := []string{}
	for _, err := range errs {
		if err != nil {
			stringErrs = append(stringErrs, err.Error())
		}
	}
	return strings.Join(stringErrs, "\n")
}
