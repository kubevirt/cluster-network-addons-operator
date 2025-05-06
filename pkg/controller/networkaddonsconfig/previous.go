package networkaddonsconfig

import (
	"context"
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	uns "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/names"
	k8sutil "github.com/kubevirt/cluster-network-addons-operator/pkg/util/k8s"
)

// GetAppliedConfiguration retrieves the configuration we applied.
// Returns nil with no error if no previous configuration was observed.
func getAppliedConfiguration(ctx context.Context, client k8sclient.Client, name string, namespace string) (*cnao.NetworkAddonsConfigSpec, error) {
	cm := &corev1.ConfigMap{}
	err := client.Get(ctx, types.NamespacedName{Name: names.AppliedPrefix + name, Namespace: namespace}, cm)
	if err != nil && apierrors.IsNotFound(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	spec := &cnao.NetworkAddonsConfigSpec{}
	err = json.Unmarshal([]byte(cm.Data["applied"]), spec)
	if err != nil {
		return nil, err
	}
	return spec, nil
}

// AppliedConfiguration renders the ConfigMap in which we store the configuration
// we've applied.
func appliedConfiguration(applied *cnao.NetworkAddonsConfig, namespace string) (*uns.Unstructured, error) {
	app, err := json.Marshal(applied.Spec)
	if err != nil {
		return nil, err
	}
	cm := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      names.AppliedPrefix + applied.Name,
			Namespace: namespace,
		},
		Data: map[string]string{
			"applied": string(app),
		},
	}

	// transmute to unstructured
	return k8sutil.ToUnstructured(cm)
}
