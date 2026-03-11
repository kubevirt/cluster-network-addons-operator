package controller

import (
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/kubevirt/cluster-network-addons-operator/pkg/controller/networkaddonsconfig"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/tlsconfig"
)

// AddToManager adds all Controllers to the Manager
func AddToManager(m manager.Manager, tlsCache *tlsconfig.Cache) error {
	return networkaddonsconfig.Add(m, tlsCache)
}
