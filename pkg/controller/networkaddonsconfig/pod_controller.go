package networkaddonsconfig

import (
	"log"
	"time"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var resyncPeriod = 5 * time.Second

// newPodReconciler returns a new reconcile.Reconciler
func newPodReconciler() *ReconcilePods {
	return &ReconcilePods{}
}

var _ reconcile.Reconciler = &ReconcilePods{}

// ReconcilePods watches for updates to specified resources and then updates its StatusManager
type ReconcilePods struct {
	//status *statusmanager.StatusManager

	resources []types.NamespacedName
}

func (r *ReconcilePods) SetResources(resources []types.NamespacedName) {
	log.Printf("XXXX Set resources %v", resources)
	r.resources = resources
}

// Reconcile updates the ClusterOperator.Status to match the current state of the
// watched Deployments/DaemonSets
func (r *ReconcilePods) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log.Printf("XXXX Reconcile 1")
	found := false
	for _, name := range r.resources {
		log.Printf("XXXX Reconcile 2 %s %s", request.Name, request.Namespace)
		if name.Namespace == request.Namespace && name.Name == request.Name {
			found = true
			break
		}
	}
	if !found {
		log.Printf("XXXX Reconcile 3")
		return reconcile.Result{}, nil
	}

	log.Printf("XXXX Reconcile 4")
	log.Printf("Reconciling update to %s/%s\n", request.Namespace, request.Name)

	return reconcile.Result{RequeueAfter: resyncPeriod}, nil
}
