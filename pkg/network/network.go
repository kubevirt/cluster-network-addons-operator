package network

import (
	"context"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"

	osv1 "github.com/openshift/api/operator/v1"
	"github.com/pkg/errors"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubectl/pkg/scheme"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/monitoring"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/util/k8s"
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
	errs = append(errs, validateMultusDynamicNetworks(conf, openshiftNetworkConfig)...)
	errs = append(errs, validateMacvtap(conf, openshiftNetworkConfig)...)

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
	fillMacvtapDefaults(conf, previous)

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
	errs = append(errs, cleanUpNamespaceLabels(ctx, client)...)

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

	o, err = renderMultusDynamicNetworks(conf, manifestDir, clusterInfo)
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
	o, err = renderKubeMacPool(conf, manifestDir, clusterInfo)
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

	// render KubeSecondaryDNS
	o, err = renderKubeSecondaryDNS(conf, manifestDir, clusterInfo)
	if err != nil {
		return nil, err
	}
	objs = append(objs, o...)

	// render KubevirtIPAMController
	o, err = renderKubevirtIPAMController(conf, manifestDir, clusterInfo)
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

	if conf.MultusDynamicNetworks == nil {
		o, err := renderMultusDynamicNetworks(prev, manifestDir, clusterInfo)
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
		o, err := renderKubeMacPool(prev, manifestDir, clusterInfo)
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
		o, err := renderMacvtapCni(prev, manifestDir, clusterInfo)
		if err != nil {
			return nil, err
		}
		objsToRemove = append(objsToRemove, o...)
	}

	// render KubeSecondaryDNS
	if conf.KubeSecondaryDNS == nil {
		o, err := renderKubeSecondaryDNS(prev, manifestDir, clusterInfo)
		if err != nil {
			return nil, err
		}
		objsToRemove = append(objsToRemove, o...)
	}

	if conf.KubevirtIpamController == nil {
		o, err := renderKubevirtIPAMController(prev, manifestDir, clusterInfo)
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

	// Remove old CNAO managed kubernetes-nmstate
	oldKNMStateObjects, err := cnaoKNMStateObjects(operandNamespace)
	if err != nil {
		return nil, err
	}
	objsToRemove = append(objsToRemove, oldKNMStateObjects...)

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

// cleanUpNamespaceLabels removes relation labels from the operator namespace
// It is done in order to support upgrading from versions where the labels were added to
// to operator namespace
func cleanUpNamespaceLabels(ctx context.Context, client k8sclient.Client) []error {
	namespace := &v1.Namespace{}
	err := client.Get(context.Background(), types.NamespacedName{Name: os.Getenv("OPERATOR_NAMESPACE")}, namespace)
	if err != nil {
		return []error{err}
	}

	labels := namespace.GetLabels()
	if len(labels) == 0 {
		return []error{}
	}

	patch := k8sclient.MergeFrom(namespace.DeepCopy())
	labelFound := false
	for _, key := range k8s.RemovedLabels() {
		if _, exist := labels[key]; exist {
			delete(labels, key)
			labelFound = true
		}
	}

	if !labelFound {
		return []error{}
	}

	namespace.SetLabels(labels)
	err = client.Patch(ctx, namespace, patch)
	if err != nil {
		return []error{err}
	}

	return []error{}
}

func cnaoKNMStateObjects(operandNamespace string) ([]*unstructured.Unstructured, error) {
	objects := []runtime.Object{
		&v1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{Namespace: operandNamespace, Name: "nmstate-handler"},
		},
		&v1.Service{
			ObjectMeta: metav1.ObjectMeta{Namespace: operandNamespace, Name: "nmstate-webhook"},
		},
		&v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Namespace: operandNamespace, Name: "nmstate-config"},
		},
		&appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{Namespace: operandNamespace, Name: "nmstate-handler"},
		},
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Namespace: operandNamespace, Name: "nmstate-webhook"},
		},
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Namespace: operandNamespace, Name: "nmstate-cert-manager"},
		},
		&rbacv1.Role{
			ObjectMeta: metav1.ObjectMeta{Namespace: operandNamespace, Name: "nmstate-handler"},
		},
		&rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{Namespace: operandNamespace, Name: "nmstate-handler"},
		},
		&rbacv1.ClusterRole{
			ObjectMeta: metav1.ObjectMeta{Name: "nmstate-handler"},
		},
		&rbacv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{Name: "nmstate-handler"},
		},
		&policyv1.PodDisruptionBudget{
			ObjectMeta: metav1.ObjectMeta{Namespace: operandNamespace, Name: "nmstate-webhook"},
		},
		&admissionregistrationv1.MutatingWebhookConfiguration{
			ObjectMeta: metav1.ObjectMeta{Name: "nmstate"},
		},
	}

	convertedObjects := []*unstructured.Unstructured{}
	for _, object := range objects {
		err := addTypeInformationToObject(object)
		if err != nil {
			return nil, err
		}
		convertedObject, err := runtime.DefaultUnstructuredConverter.ToUnstructured(object)
		if err != nil {
			return nil, err
		}
		convertedObjects = append(convertedObjects, &unstructured.Unstructured{Object: convertedObject})
	}
	return convertedObjects, nil
}

// addTypeInformationToObject adds TypeMeta information to a runtime.Object based upon the loaded scheme.Scheme
// Related to issue https://github.com/kubernetes/kubernetes/issues/3030
func addTypeInformationToObject(obj runtime.Object) error {
	gvks, _, err := scheme.Scheme.ObjectKinds(obj)
	if err != nil {
		return fmt.Errorf("missing apiVersion or kind and cannot assign it; %w", err)
	}

	for _, gvk := range gvks {
		if len(gvk.Kind) == 0 {
			continue
		}
		if len(gvk.Version) == 0 || gvk.Version == runtime.APIVersionInternal {
			continue
		}
		obj.GetObjectKind().SetGroupVersionKind(gvk)
		break
	}

	return nil
}
