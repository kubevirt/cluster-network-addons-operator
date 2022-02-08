package networkaddonsconfig

import (
	"context"
	"log"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/kubevirt/cluster-network-addons-operator/pkg/controller/statusmanager"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/eventemitter"
)

// newPodReconciler returns a new reconcile.Reconciler
func newPodReconciler(statusManager *statusmanager.StatusManager, mgr manager.Manager) *ReconcilePods {
	return &ReconcilePods{
		statusManager: statusManager,
		eventEmitter:  eventemitter.New(mgr),
	}
}

var _ reconcile.Reconciler = &ReconcilePods{}

// ReconcilePods watches for updates to specified resources and then updates its StatusManager
type ReconcilePods struct {
	statusManager *statusmanager.StatusManager
	resources     []types.NamespacedName
	eventEmitter  eventemitter.EventEmitter
}

// SetResources updates context's resources
func (r *ReconcilePods) SetResources(resources []types.NamespacedName) {
	r.resources = resources
}

// Reconcile updates the NetworkAddonsConfig.Status to match the current state of the
// watched Deployments/DaemonSets
func (r *ReconcilePods) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	for _, name := range r.resources {
		if name.Namespace == request.Namespace && name.Name == request.Name {
			log.Printf("Reconciling update to %s/%s\n", request.Namespace, request.Name)
			r.eventEmitter.EmitModifiedForConfig()
			r.statusManager.SetFromPods()
			return reconcile.Result{}, nil
		}
	}
	return reconcile.Result{}, nil
}
