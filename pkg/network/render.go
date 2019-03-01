package network

import (
	"log"
	"reflect"

	"github.com/pkg/errors"

	networkaddonsoperatorv1alpha1 "github.com/phoracek/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/v1alpha1"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func Render(conf *networkaddonsoperatorv1alpha1.NetworkAddonsConfigSpec, manifestDir string) ([]*unstructured.Unstructured, error) {
	log.Printf("Starting render phase")
	objs := []*unstructured.Unstructured{}

	// render Multus
	o, err := renderMultus(conf, manifestDir)
	if err != nil {
		return nil, err
	}
	objs = append(objs, o...)

	log.Printf("Render phase done, rendered %d objects", len(objs))
	return objs, nil
}

// Canonicalize converts configuration to a canonical form.
// Currently we only care about case.
func Canonicalize(conf *networkaddonsoperatorv1alpha1.NetworkAddonsConfigSpec) {
}

// Validate checks that the supplied configuration is reasonable.
// This should be called after Canonicalize
func Validate(conf *networkaddonsoperatorv1alpha1.NetworkAddonsConfigSpec) error {
	errs := []error{}

	errs = append(errs, validateMultus(conf)...)

	if len(errs) > 0 {
		return errors.Errorf("invalid configuration: %v", errs)
	}
	return nil
}

// FillDefaults computes any default values and applies them to the configuration
// This is a mutating operation. It should be called after Validate.
//
// Defaults are carried forward from previous if it is provided. This is so we
// can change defaults as we move forward, but won't disrupt existing clusters.
func FillDefaults(conf, previous *networkaddonsoperatorv1alpha1.NetworkAddonsConfigSpec) {
	// TODO
}

// IsChangeSafe checks to see if the change between prev and next are allowed
// FillDefaults and Validate should have been called.
func IsChangeSafe(prev, next *networkaddonsoperatorv1alpha1.NetworkAddonsConfigSpec) error {
	if prev == nil {
		return nil
	}

	// Easy way out: nothing changed.
	if reflect.DeepEqual(prev, next) {
		return nil
	}

	errs := []error{}

	// TODO

	if len(errs) > 0 {
		return errors.Errorf("invalid configuration: %v", errs)
	}
	return nil
}
