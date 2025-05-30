package networkaddonsconfig

import (
	"context"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"time"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	osconfv1 "github.com/openshift/api/config/v1"
	osv1 "github.com/openshift/api/operator/v1"
	osnetnames "github.com/openshift/cluster-network-operator/pkg/names"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
	cnaov1 "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/v1"
	cnaov1alpha1 "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/v1alpha1"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/apply"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/controller/statusmanager"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/eventemitter"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/monitoring/metrics"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/names"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/network"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/util/k8s"
)

// ManifestPath is the path to the manifest templates
const ManifestPath = "./data"

var operatorNamespace string
var operatorVersion string
var operatorVersionLabel string

func init() {
	operatorNamespace = os.Getenv("OPERATOR_NAMESPACE")
	operatorVersion = os.Getenv("OPERATOR_VERSION")
	operatorVersionLabel = k8s.StringToLabel(operatorVersion)
}

// Add creates a new NetworkAddonsConfig Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	cfg, err := config.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to get apiserver config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize apiserver client: %v", err)
	}

	namespace, namespaceSet := os.LookupEnv("OPERATOR_NAMESPACE")
	if !namespaceSet {
		return fmt.Errorf("environment variable OPERATOR_NAMESPACE has to be set")
	}

	clusterInfo := &network.ClusterInfo{}

	openShift4, err := isRunningOnOpenShift4(clientset)
	if err != nil {
		return fmt.Errorf("failed to check whether running on OpenShift 4: %v", err)
	}
	if openShift4 {
		log.Printf("Running on OpenShift 4")
	}
	clusterInfo.OpenShift4 = openShift4

	sccAvailable, err := isSCCAvailable(clientset)
	if err != nil {
		return fmt.Errorf("failed to check for availability of SCC: %v", err)
	}
	clusterInfo.SCCAvailable = sccAvailable

	addMonitorServiceResources, err := IsMonitoringAvailable(clientset)
	if err != nil {
		// we don't want CNAO to fail only if monitoring cannot be activated.
		addMonitorServiceResources = false
		log.Printf("failed to check for availability of Monitoring namespace: %v", err)
	}
	clusterInfo.MonitoringAvailable = addMonitorServiceResources

	return add(mgr, newReconciler(mgr, namespace, clusterInfo))
}

// newReconciler returns a new ReconcileNetworkAddonsConfig
func newReconciler(mgr manager.Manager, namespace string, clusterInfo *network.ClusterInfo) *ReconcileNetworkAddonsConfig {
	// Status manager is shared between both reconcilers, and it is used to update conditions of
	// NetworkAddonsConfig.State. NetworkAddonsConfig reconciler updates it with progress of rendering
	// and applying of manifests. Pods reconciler updates it with progress of deployed pods.
	statusManager := statusmanager.New(mgr, names.OperatorConfig)
	return &ReconcileNetworkAddonsConfig{
		client:        mgr.GetClient(),
		scheme:        mgr.GetScheme(),
		namespace:     namespace,
		podReconciler: newPodReconciler(statusManager, mgr),
		statusManager: statusManager,
		clusterInfo:   clusterInfo,
		eventEmitter:  eventemitter.New(mgr),
	}
}

type ctrlPredicate[T metav1.Object] struct {
	predicate.TypedFuncs[T]
}

func (p ctrlPredicate[T]) Update(e event.TypedUpdateEvent[T]) bool {
	oldConfig, err := runtimeObjectToNetworkAddonsConfig(e.ObjectOld)
	if err != nil {
		log.Printf("Failed to convert runtime.Object to NetworkAddonsConfig (old): %v", err)
		return false
	}
	newConfig, err := runtimeObjectToNetworkAddonsConfig(e.ObjectNew)
	if err != nil {
		log.Printf("Failed to convert runtime.Object to NetworkAddonsConfig (new): %v", err)
		return false
	}
	return !reflect.DeepEqual(oldConfig.Spec, newConfig.Spec)
}

// add adds a new Controller to mgr with r as the ReconcileNetworkAddonsConfig
func add(mgr manager.Manager, r *ReconcileNetworkAddonsConfig) error {
	c, err := controller.New("networkaddonsconfig-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	if err := c.Watch(
		source.Kind(
			mgr.GetCache(),
			&cnaov1alpha1.NetworkAddonsConfig{},
			&handler.TypedEnqueueRequestForObject[*cnaov1alpha1.NetworkAddonsConfig]{},
			&ctrlPredicate[*cnaov1alpha1.NetworkAddonsConfig]{},
		),
	); err != nil {
		return fmt.Errorf("unable to watch NetworkAddonsConfig v1alpha1: %w", err)
	}

	if err := c.Watch(
		source.Kind(
			mgr.GetCache(),
			&cnaov1.NetworkAddonsConfig{},
			&handler.TypedEnqueueRequestForObject[*cnaov1.NetworkAddonsConfig]{},
			&ctrlPredicate[*cnaov1.NetworkAddonsConfig]{},
		),
	); err != nil {
		return fmt.Errorf("unable to watch NetworkAddonsConfig v1: %w", err)
	}

	// Create a new controller for Pod resources, this will be used to track state of deployed components
	c, err = controller.New("pod-controller", mgr, controller.Options{Reconciler: r.podReconciler})
	if err != nil {
		return err
	}

	if err := c.Watch(
		source.Kind(
			mgr.GetCache(),
			&appsv1.DaemonSet{},
			&handler.TypedEnqueueRequestForObject[*appsv1.DaemonSet]{},
		),
	); err != nil {
		return fmt.Errorf("unable to watch NetworkAddonsConfig v1: %w", err)
	}

	if err := c.Watch(
		source.Kind(
			mgr.GetCache(),
			&appsv1.Deployment{},
			&handler.TypedEnqueueRequestForObject[*appsv1.Deployment]{},
		),
	); err != nil {
		return fmt.Errorf("unable to watch NetworkAddonsConfig v1: %w", err)
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileNetworkAddonsConfig{}

// ReconcileNetworkAddonsConfig reconciles a NetworkAddonsConfig object
type ReconcileNetworkAddonsConfig struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client        client.Client
	scheme        *runtime.Scheme
	namespace     string
	podReconciler *ReconcilePods
	statusManager *statusmanager.StatusManager
	clusterInfo   *network.ClusterInfo
	eventEmitter  eventemitter.EventEmitter
}

// Reconcile reads that state of the cluster for a NetworkAddonsConfig object and makes changes based on the state read
// and what is in the NetworkAddonsConfig.Spec
func (r *ReconcileNetworkAddonsConfig) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	log.Print("reconciling NetworkAddonsConfig")

	// We won't create more than one network addons instance
	if request.Name != names.OperatorConfig {
		log.Print("ignoring NetworkAddonsConfig without default name")
		return reconcile.Result{}, nil
	}

	if r.clusterInfo.OpenShift4 {
		isSingleReplica, err := isOpenshiftSingleReplica(r.client)
		if err != nil {
			log.Printf("failed to check if running on a singleReplica infrastrcuture: %v", err)
		}
		r.clusterInfo.IsSingleReplica = isSingleReplica
	}

	// Fetch the NetworkAddonsConfig instance
	networkAddonsConfigStorageVersion := &cnaov1.NetworkAddonsConfig{}
	err := r.client.Get(context.TODO(), request.NamespacedName, networkAddonsConfigStorageVersion)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Reset list of tracked objects.
			// TODO: This can be dropped once we implement a finalizer waiting for all components to be removed
			r.stopTrackingObjects()

			if r.clusterInfo.MonitoringAvailable {
				metrics.ResetMonitoredComponents()
			}

			// Owned objects are automatically garbage collected. Return and don't requeue
			return reconcile.Result{}, nil
		}

		log.Printf("Error reading NetworkAddonsConfig. err = %v", err)
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	networkAddonsConfig, err := r.ConvertNetworkAddonsConfigV1ToShared(networkAddonsConfigStorageVersion)
	if err != nil {
		// If failed, set NetworkAddonsConfig to failing and requeue
		err = errors.Wrap(err, "failed converting NetworkAddonsConfig to internal structure")
		r.statusManager.SetFailing(statusmanager.OperatorConfig, "updateNetworkAddonsConfigToV1", err.Error())
		return reconcile.Result{}, err
	}

	// Convert to a canonicalized form
	network.Canonicalize(&networkAddonsConfig.Spec)

	// Read OpenShift network operator configuration (if exists)
	openshiftNetworkConfig, err := getOpenShiftNetworkConfig(context.TODO(), r.client)
	if err != nil {
		log.Printf("failed to load OpenShift NetworkConfig: %v", err)
		err = errors.Wrapf(err, "failed to load OpenShift NetworkConfig")
		r.statusManager.SetFailing(statusmanager.OperatorConfig, "FailedToGetOpenShiftNetworkConfig", err.Error())
		return reconcile.Result{}, err
	}

	// Validate the configuration
	if err := network.Validate(&networkAddonsConfig.Spec, openshiftNetworkConfig); err != nil {
		log.Printf("failed to validate NetworkConfig.Spec: %v", err)
		err = errors.Wrapf(err, "failed to validate NetworkConfig.Spec")
		r.statusManager.SetFailing(statusmanager.OperatorConfig, "FailedToValidate", err.Error())
		return reconcile.Result{}, err
	}
	prev, err := r.getPreviousConfigSpec(networkAddonsConfig)
	if err != nil {
		// If failed, set NetworkAddonsConfig to failing and requeue
		r.statusManager.SetFailing(statusmanager.OperatorConfig, "FailedToGetPreviousConfigSpec", err.Error())
		return reconcile.Result{}, err
	}

	// Canonicalize and validate NetworkAddonsConfig, finally render objects of requested components
	objs, err := r.renderObjectsV1(networkAddonsConfig, openshiftNetworkConfig)
	if err != nil {
		// If failed, set NetworkAddonsConfig to failing and requeue
		r.statusManager.SetFailing(statusmanager.OperatorConfig, "FailedToRender", err.Error())
		return reconcile.Result{}, err
	}

	objsToRemove, err := r.renderObjectsToDelete(networkAddonsConfig, openshiftNetworkConfig, prev)
	if err != nil {
		// If failed, set NetworkAddonsConfig to failing and requeue
		r.statusManager.SetFailing(statusmanager.OperatorConfig, "FailedToRenderDelete", err.Error())
		return reconcile.Result{}, err
	}

	// Apply generated objects on Kubernetes API server
	err = r.applyObjects(networkAddonsConfigStorageVersion, objs)
	if err != nil {
		// If failed, set NetworkAddonsConfig to failing and requeue
		r.statusManager.SetFailing(statusmanager.OperatorConfig, "FailedToApply", err.Error())
		return reconcile.Result{}, err
	}

	// Track state of all deployed pods
	r.trackDeployedObjects(objs, networkAddonsConfig.GetGeneration())

	// Delete generated objsToRemove on Kubernetes API server
	err = r.deleteOwnedObjects(objsToRemove)
	if err != nil {
		// If failed, set NetworkAddonsConfig to failing and requeue
		r.statusManager.SetFailing(statusmanager.OperatorConfig, "FailedToDeleteObjects", err.Error())
		return reconcile.Result{}, err
	}

	// Everything went smooth, remove failures from NetworkAddonsConfig if there are any from
	// previous runs.
	r.statusManager.MarkStatusLevelNotFailing(statusmanager.OperatorConfig)
	r.statusManager.SetFromOperator()

	// From now on, r.podReconciler takes over NetworkAddonsConfig handling, it will track deployed
	// objects if needed and set NetworkAddonsConfig.Status accordingly. However, if no pod was
	// deployed, there is nothing that would trigger initial reconciliation. Therefore, let's
	// perform this check manually.
	r.statusManager.SetFromPods()

	if r.clusterInfo.MonitoringAvailable {
		metrics.TrackMonitoredComponents(&networkAddonsConfig.Spec, r.statusManager)
	}

	// Kubernetes sometimes fails to apply objects while we remove and recreate
	// components, despite reporting success. In order to self-heal after these
	// incidents, keep requeing.
	return reconcile.Result{RequeueAfter: time.Minute}, nil
}

// ConvertNetworkAddonsConfigV1ToShared converts NetworkAddonsConfig to shared type
func (r *ReconcileNetworkAddonsConfig) ConvertNetworkAddonsConfigV1ToShared(networkAddonsConfig *cnaov1.NetworkAddonsConfig) (*cnao.NetworkAddonsConfig, error) {
	return &cnao.NetworkAddonsConfig{
		TypeMeta:   networkAddonsConfig.TypeMeta,
		ObjectMeta: networkAddonsConfig.ObjectMeta,
		Spec:       networkAddonsConfig.Spec,
		Status:     networkAddonsConfig.Status,
	}, nil
}

// Render objects for all desired components
func (r *ReconcileNetworkAddonsConfig) renderObjectsV1(networkAddonsConfig *cnao.NetworkAddonsConfig, openshiftNetworkConfig *osv1.Network) ([]*unstructured.Unstructured, error) {
	// Generate the objects
	objs, err := network.Render(&networkAddonsConfig.Spec, ManifestPath, openshiftNetworkConfig, r.clusterInfo)
	if err != nil {
		log.Printf("failed to render: %v", err)
		err = errors.Wrapf(err, "failed to render")
		return objs, err
	}

	// Perform any special object changes that are impossible to do with regular Apply. e.g. Remove outdated objects
	// and objects that cannot be modified by Apply method due to incompatible changes.
	if err := network.SpecialCleanUp(&networkAddonsConfig.Spec, r.client, r.clusterInfo); err != nil {
		log.Printf("failed to Clean Up outdated objects: %v", err)
		return objs, err
	}

	// The first object we create should be the record of our applied configuration
	applied, err := appliedConfiguration(networkAddonsConfig, r.namespace)
	if err != nil {
		log.Printf("failed to render applied: %v", err)
		err = errors.Wrapf(err, "failed to render applied")
		return objs, err
	}
	objs = append([]*unstructured.Unstructured{applied}, objs...)

	err = updateObjectsLabels(networkAddonsConfig.GetLabels(), objs)
	if err != nil {
		log.Printf("failed to update objects labels: %v", err)
		err = errors.Wrapf(err, "failed to update objects labels")
		return objs, err
	}

	return objs, nil
}

// Validate and returns the previous configuration spec
func (r *ReconcileNetworkAddonsConfig) getPreviousConfigSpec(networkAddonsConfig *cnao.NetworkAddonsConfig) (*cnao.NetworkAddonsConfigSpec, error) {
	// Retrieve the previously applied operator configuration
	prev, err := getAppliedConfiguration(context.TODO(), r.client, networkAddonsConfig.ObjectMeta.Name, r.namespace)
	if err != nil {
		log.Printf("failed to retrieve previously applied configuration: %v", err)
		err = errors.Wrapf(err, "failed to retrieve previously applied configuration")
		return nil, err
	}

	// Fill all defaults explicitly
	if err := network.FillDefaults(&networkAddonsConfig.Spec, prev); err != nil {
		log.Printf("failed to fill defaults: %v", err)
		err = errors.Wrapf(err, "failed to fill defaults")
		return nil, err
	}

	// Compare against previous applied configuration to see if this change
	// is safe.
	if prev != nil {
		// We may need to fill defaults here -- sort of as a poor-man's
		// upconversion scheme -- if we add additional fields to the config.
		err = network.IsChangeSafe(prev, &networkAddonsConfig.Spec)
		if err != nil {
			log.Printf("not applying unsafe change: %v", err)
			err = errors.Wrapf(err, "not applying unsafe change")
			return nil, err
		}
	}

	return prev, nil
}

// Generate the removal object list
func (r *ReconcileNetworkAddonsConfig) renderObjectsToDelete(networkAddonsConfig *cnao.NetworkAddonsConfig, openshiftNetworkConfig *osv1.Network, prev *cnao.NetworkAddonsConfigSpec) ([]*unstructured.Unstructured, error) {
	objsToRemove, err := network.RenderObjsToRemove(r.scheme, prev, &networkAddonsConfig.Spec, ManifestPath, openshiftNetworkConfig, r.clusterInfo)
	if err != nil {
		log.Printf("failed to render for removal: %v", err)
		err = errors.Wrapf(err, "failed to render for removal")
		return objsToRemove, err
	}

	return objsToRemove, nil
}

// Apply the objects to the cluster. Set their controller reference to NetworkAddonsConfig, so they
// are removed when NetworkAddonsConfig config is
func (r *ReconcileNetworkAddonsConfig) applyObjects(networkAddonsConfig metav1.Object, objs []*unstructured.Unstructured) error {
	for _, obj := range objs {
		// Mark the object to be GC'd if the owner is deleted.
		// Don't set owner reference on namespaces if they are used by the operator itself
		// Don't set owner reference on CRDs, they should survive removal of the operator
		// Don't set owner reference on objects that explicitly rejected an owner
		isCRD := obj.GetKind() == "CustomResourceDefinition"
		_, isRejectingOwner := obj.GetAnnotations()[names.RejectOwnerAnnotation]
		if !isCRD && !isOperatorNamespace(obj) && !isRejectingOwner {
			if err := controllerutil.SetControllerReference(networkAddonsConfig, obj, r.scheme); err != nil {
				log.Printf("could not set reference for (%s) %s/%s: %v", obj.GroupVersionKind(), obj.GetNamespace(), obj.GetName(), err)
				err = errors.Wrapf(err, "could not set reference for (%s) %s/%s", obj.GroupVersionKind(), obj.GetNamespace(), obj.GetName())
				return err
			}
		}

		// Apply all objects on apiserver
		if err := apply.ApplyObject(context.TODO(), r.client, obj); err != nil {
			log.Printf("could not apply (%s) %s/%s: %v", obj.GroupVersionKind(), obj.GetNamespace(), obj.GetName(), err)
			err = errors.Wrapf(err, "could not apply (%s) %s/%s", obj.GroupVersionKind(), obj.GetNamespace(), obj.GetName())
			return err
		}
	}

	return nil
}

// Delete removed objects
func (r *ReconcileNetworkAddonsConfig) deleteOwnedObjects(objs []*unstructured.Unstructured) error {
	for _, obj := range objs {
		if err := apply.DeleteOwnedObject(context.TODO(), r.client, obj); err != nil {
			log.Printf("could not delete (%s) %s/%s: %v", obj.GroupVersionKind(), obj.GetNamespace(), obj.GetName(), err)
			err = errors.Wrapf(err, "could not delete (%s) %s/%s", obj.GroupVersionKind(), obj.GetNamespace(), obj.GetName())
			return err
		}
	}

	return nil
}

// Track current state of Deployments and DaemonSets deployed by the operator. This is needed to
// keep state of NetworkAddonsConfig up-to-date, e.g. mark as Ready once all objects are successfully
// created. This also exposes all containers and their images used by deployed components in Status.
func (r *ReconcileNetworkAddonsConfig) trackDeployedObjects(objs []*unstructured.Unstructured, generation int64) {
	daemonSets := []types.NamespacedName{}
	deployments := []types.NamespacedName{}
	containers := []cnao.Container{}

	for _, obj := range objs {
		if obj.GetAPIVersion() == "apps/v1" && obj.GetKind() == "DaemonSet" {
			daemonSets = append(daemonSets, types.NamespacedName{Namespace: obj.GetNamespace(), Name: obj.GetName()})

			daemonSet, err := unstructuredToDaemonSet(obj)
			if err != nil {
				log.Printf("Failed to detect images used in DaemonSet %q: %v", obj.GetName(), err)
				continue
			}

			containers = append(containers, collectContainersInfo(obj.GetKind(), daemonSet.GetName(), daemonSet.Spec.Template.Spec.InitContainers)...)
			containers = append(containers, collectContainersInfo(obj.GetKind(), daemonSet.GetName(), daemonSet.Spec.Template.Spec.Containers)...)
		} else if obj.GetAPIVersion() == "apps/v1" && obj.GetKind() == "Deployment" {
			deployments = append(deployments, types.NamespacedName{Namespace: obj.GetNamespace(), Name: obj.GetName()})

			deployment, err := unstructuredToDeployment(obj)
			if err != nil {
				log.Printf("Failed to detect images used in Deployment %q: %v", obj.GetName(), err)
				continue
			}

			containers = append(containers, collectContainersInfo(obj.GetKind(), deployment.GetName(), deployment.Spec.Template.Spec.InitContainers)...)
			containers = append(containers, collectContainersInfo(obj.GetKind(), deployment.GetName(), deployment.Spec.Template.Spec.Containers)...)
		}
	}

	r.statusManager.SetAttributes(daemonSets, deployments, containers, generation)

	allResources := []types.NamespacedName{}
	allResources = append(allResources, daemonSets...)
	allResources = append(allResources, deployments...)

	r.podReconciler.SetResources(allResources)
}

// Stop tracking current state of Deployments and DaemonSets deployed by the operator.
func (r *ReconcileNetworkAddonsConfig) stopTrackingObjects() {
	// reset generation number by using invalid generation value
	r.statusManager.SetAttributes([]types.NamespacedName{}, []types.NamespacedName{}, []cnao.Container{}, -1)

	r.podReconciler.SetResources([]types.NamespacedName{})

	// Trigger status manager to notice the change
	r.statusManager.SetFromPods()
}

func updateObjectsLabels(crLabels map[string]string, objs []*unstructured.Unstructured) error {
	var err error
	for _, obj := range objs {
		labels := obj.GetLabels()
		if labels == nil {
			labels = map[string]string{}
		}
		if !isOperatorNamespace(obj) {
			// Label objects with version of the operator they were created by
			labels[cnaov1.GroupVersion.Group+"/version"] = operatorVersionLabel
			labels[names.PrometheusLabelKey] = names.PrometheusLabelValue
			labels[names.ManagedByLabelKey] = names.ManagedByLabelDefaultValue

			appLabelKeys := []string{names.ComponentLabelKey, names.PartOfLabelKey, names.VersionLabelKey}
			labels = updateLabelsFromCR(labels, crLabels, appLabelKeys)

			templateLabelKeys := append(appLabelKeys, names.PrometheusLabelKey, names.ManagedByLabelKey)
			err = updateObjectTemplateLabels(obj, labels, templateLabelKeys)
			if err != nil {
				return err
			}
		} else {
			delete(labels, names.KubemacpoolControlPlaneKey)
		}

		obj.SetLabels(labels)
	}

	return nil
}

func updateLabelsFromCR(labels, crLabels map[string]string, appLabelKeys []string) map[string]string {
	labels[names.ComponentLabelKey] = names.ComponentLabelDefaultValue
	for _, key := range appLabelKeys {
		if value, exist := crLabels[key]; exist == true {
			labels[key] = value
		}
	}

	return labels
}

func updateObjectTemplateLabels(obj *unstructured.Unstructured, labels map[string]string, templateLabelKeys []string) error {
	kind := obj.GetKind()
	if kind == "DaemonSet" || kind == "ReplicaSet" || kind == "Deployment" || kind == "StatefulSet" {
		for _, key := range templateLabelKeys {
			if value, exist := labels[key]; exist == true {
				err := unstructured.SetNestedField(obj.Object, value, "spec", "template", "metadata", "labels", key)
				if err != nil {
					log.Printf("failed to add relationship label %s: %v", key, err)
					err = errors.Wrapf(err, "failed to add relationship labels")
					return err
				}
			}
		}
	}

	return nil
}

func collectContainersInfo(parentKind string, parentName string, containers []v1.Container) []cnao.Container {
	containersInfo := []cnao.Container{}

	for _, container := range containers {
		containersInfo = append(containersInfo, cnao.Container{
			ParentKind: parentKind,
			ParentName: parentName,
			Image:      container.Image,
			Name:       container.Name,
		})
	}

	return containersInfo
}

func getOpenShiftNetworkConfig(ctx context.Context, c k8sclient.Client) (*osv1.Network, error) {
	nc := &osv1.Network{}

	err := c.Get(ctx, types.NamespacedName{Namespace: "", Name: osnetnames.OPERATOR_CONFIG}, nc)
	if err != nil {
		if apierrors.IsNotFound(err) || strings.Contains(err.Error(), "no matches for kind") {
			log.Printf("OpenShift cluster network configuration resource has not been found: %v", err)
			return nil, nil
		}
		log.Printf("failed to obtain OpenShift cluster network configuration with unexpected error: %v", err)
		return nil, err
	}

	return nc, nil
}

// Check whether running on OpenShift 4 by looking for operator objects that has been introduced
// only in OpenShift 4
func isRunningOnOpenShift4(c kubernetes.Interface) (bool, error) {
	return isResourceAvailable(c, "networks", "operator.openshift.io", "v1")
}

func isSCCAvailable(c kubernetes.Interface) (bool, error) {
	return isResourceAvailable(c, "securitycontextconstraints", "security.openshift.io", "v1")
}

// IsMonitoringAvailable checks if we can deploy the monitoring component
func IsMonitoringAvailable(c kubernetes.Interface) (bool, error) {
	prometheusRuleResourceAvailable, err := isResourceAvailable(c, "customresourcedefinitions/prometheusrules.monitoring.coreos.com", "apiextensions.k8s.io", "v1")
	if err != nil {
		return false, errors.Wrap(err, "failed to check if prometheusRule resource is available")
	}

	serviceMonitorResourceAvailable, err := isResourceAvailable(c, "customresourcedefinitions/servicemonitors.monitoring.coreos.com", "apiextensions.k8s.io", "v1")
	if err != nil {
		return false, errors.Wrap(err, "failed to check if serviceMonitor resource is available")
	}

	if prometheusRuleResourceAvailable && serviceMonitorResourceAvailable {
		return true, nil
	}

	log.Printf("will not deploy monitoring manifests: not all monitoring resources are available: %s: %v, %s, %v", monitoringv1.PrometheusRuleKind, prometheusRuleResourceAvailable, monitoringv1.ServiceMonitorsKind, serviceMonitorResourceAvailable)
	return false, nil
}

func isResourceAvailable(kubeClient kubernetes.Interface, name string, group string, version string) (bool, error) {
	result := kubeClient.ExtensionsV1beta1().RESTClient().Get().RequestURI("/apis/" + group + "/" + version + "/" + name).Do(context.TODO())
	if result.Error() != nil {
		if strings.Contains(result.Error().Error(), "the server could not find the requested resource") {
			return false, nil
		}
		return false, result.Error()
	}

	return true, nil
}

func runtimeObjectToNetworkAddonsConfig(obj interface{}) (*cnao.NetworkAddonsConfig, error) {
	unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, err
	}

	networkAddonsConfig := &cnao.NetworkAddonsConfig{}
	if err = runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredObj, networkAddonsConfig); err != nil {
		return nil, err
	}

	return networkAddonsConfig, nil
}

func unstructuredToDaemonSet(obj *unstructured.Unstructured) (*appsv1.DaemonSet, error) {
	daemonSet := &appsv1.DaemonSet{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, daemonSet); err != nil {
		return nil, err
	}
	return daemonSet, nil
}

func unstructuredToDeployment(obj *unstructured.Unstructured) (*appsv1.Deployment, error) {
	deployment := &appsv1.Deployment{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, deployment); err != nil {
		return nil, err
	}
	return deployment, nil
}

func isOpenshiftSingleReplica(c k8sclient.Client) (bool, error) {
	infraConfig := &osconfv1.Infrastructure{}
	if err := c.Get(context.TODO(), types.NamespacedName{Name: "cluster"}, infraConfig); err != nil {
		return false, err
	}

	return infraConfig.Status.InfrastructureTopology == osconfv1.SingleReplicaTopologyMode, nil
}

func isOperatorNamespace(obj *unstructured.Unstructured) bool {
	const namespaceKind = "Namespace"
	return obj.GetKind() == namespaceKind && obj.GetName() == operatorNamespace
}
