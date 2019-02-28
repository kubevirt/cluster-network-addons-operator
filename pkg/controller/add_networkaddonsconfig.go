package controller

import (
	"github.com/phoracek/cluster-network-addons-operator/pkg/controller/networkaddonsconfig"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, networkaddonsconfig.Add)
}
