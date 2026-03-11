package controller

import (
	"sync/atomic"

	ocpv1 "github.com/openshift/api/config/v1"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/kubevirt/cluster-network-addons-operator/pkg/controller/networkaddonsconfig"
)

// AddToManager adds all Controllers to the Manager
func AddToManager(m manager.Manager, tlsProfile *atomic.Pointer[ocpv1.TLSSecurityProfile]) error {
	return networkaddonsconfig.Add(m, tlsProfile)
}
