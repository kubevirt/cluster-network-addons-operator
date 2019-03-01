package networkaddonsconfig

import (
	"context"

	networkaddonsoperatorv1alpha1 "github.com/phoracek/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/v1alpha1"
	"github.com/phoracek/cluster-network-addons-operator/pkg/apply"
	"github.com/phoracek/cluster-network-addons-operator/pkg/names"
	"github.com/phoracek/cluster-network-addons-operator/pkg/network"

	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	uns "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// ManifestPath is the path to the manifest templates
var ManifestPath = "./data"

var log = logf.Log.WithName("controller_networkaddonsconfig")

// Add creates a new NetworkAddonsConfig Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileNetworkAddonsConfig{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("networkaddonsconfig-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource NetworkAddonsConfig
	err = c.Watch(&source.Kind{Type: &networkaddonsoperatorv1alpha1.NetworkAddonsConfig{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileNetworkAddonsConfig{}

// ReconcileNetworkAddonsConfig reconciles a NetworkAddonsConfig object
type ReconcileNetworkAddonsConfig struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a NetworkAddonsConfig object and makes changes based on the state read
// and what is in the NetworkAddonsConfig.Spec
func (r *ReconcileNetworkAddonsConfig) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling NetworkAddonsConfig")

	// We won't create more than one network addons instance
	if request.Name != names.OPERATOR_CONFIG {
		log.Info("Ignoring NetworkAddonsConfig without default name")
		return reconcile.Result{}, nil
	}

	// Fetch the NetworkAddonsConfig instance
	networkAddonsConfig := &networkaddonsoperatorv1alpha1.NetworkAddonsConfig{}
	err := r.client.Get(context.TODO(), request.NamespacedName, networkAddonsConfig)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Convert to a canonicalized form
	network.Canonicalize(&networkAddonsConfig.Spec)

	// Validate the configuration
	if err := network.Validate(&networkAddonsConfig.Spec); err != nil {
		reqLogger.Info("Failed to validate NetworkConfig.Spec: %v", err)
		return reconcile.Result{}, err
	}

	// Retrieve the previously applied operator configuration
	prev, err := GetAppliedConfiguration(context.TODO(), r.client, networkAddonsConfig.ObjectMeta.Name)
	if err != nil {
		reqLogger.Info("Failed to retrieve previously applied configuration: %v", err)
		return reconcile.Result{}, err
	}

	// Fill all defaults explicitly
	network.FillDefaults(&networkAddonsConfig.Spec, prev)

	// Compare against previous applied configuration to see if this change
	// is safe.
	if prev != nil {
		// We may need to fill defaults here -- sort of as a poor-man's
		// upconversion scheme -- if we add additional fields to the config.
		err = network.IsChangeSafe(prev, &networkAddonsConfig.Spec)
		if err != nil {
			reqLogger.Info("Not applying unsafe change: %v", err)
			errors.Wrapf(err, "not applying unsafe change")
			return reconcile.Result{}, err
		}
	}

	// Generate the objects
	objs, err := network.Render(&networkAddonsConfig.Spec, ManifestPath)
	if err != nil {
		reqLogger.Info("Failed to render: %v", err)
		err = errors.Wrapf(err, "failed to render")
		return reconcile.Result{}, err
	}

	// The first object we create should be the record of our applied configuration
	applied, err := AppliedConfiguration(networkAddonsConfig)
	if err != nil {
		reqLogger.Info("Failed to render applied: %v", err)
		err = errors.Wrapf(err, "failed to render applied")
		return reconcile.Result{}, err
	}
	objs = append([]*uns.Unstructured{applied}, objs...)

	// Apply the objects to the cluster
	for _, obj := range objs {
		// Mark the object to be GC'd if the owner is deleted
		if err := controllerutil.SetControllerReference(networkAddonsConfig, obj, r.scheme); err != nil {
			err = errors.Wrapf(err, "could not set reference for (%s) %s/%s", obj.GroupVersionKind(), obj.GetNamespace(), obj.GetName())
			reqLogger.Info("%v", err)
			return reconcile.Result{}, err
		}

		// Open question: should an error here indicate we will never retry?
		if err := apply.ApplyObject(context.TODO(), r.client, obj); err != nil {
			err = errors.Wrapf(err, "could not apply (%s) %s/%s", obj.GroupVersionKind(), obj.GetNamespace(), obj.GetName())
			reqLogger.Info("%v", err)

			// Ignore errors if we've asked to do so.
			anno := obj.GetAnnotations()
			if anno != nil {
				if _, ok := anno[names.IgnoreObjectErrorAnnotation]; ok {
					reqLogger.Info("Object has ignore-errors annotation set, continuing")
					continue
				}
			}
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}
