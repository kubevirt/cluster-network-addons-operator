package network

import (
	"context"
	"os"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// cleanUpKubeSecondaryDNS removes legacy kube-secondary-dns objects left from
// older releases now that the component is no longer supported by CNAO.
func cleanUpKubeSecondaryDNS(ctx context.Context, client k8sclient.Client) []error {
	namespace := os.Getenv("OPERAND_NAMESPACE")

	resources := []struct {
		gvk  schema.GroupVersionKind
		name string
	}{
		{
			gvk:  schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ConfigMap"},
			name: "secondary-dns",
		},
		{
			gvk:  schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ServiceAccount"},
			name: "secondary",
		},
		{
			gvk:  schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
			name: "secondary-dns",
		},
		{
			gvk:  schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "ClusterRole"},
			name: "secondary",
		},
		{
			gvk:  schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "ClusterRoleBinding"},
			name: "secondary",
		},
		{
			gvk:  schema.GroupVersionKind{Group: "networking.k8s.io", Version: "v1", Kind: "NetworkPolicy"},
			name: "allow-ingress-to-secondary-dns",
		},
	}

	var errors []error
	for _, resource := range resources {
		existing := &unstructured.Unstructured{}
		existing.SetGroupVersionKind(resource.gvk)

		err := client.Get(ctx, types.NamespacedName{Name: resource.name, Namespace: namespace}, existing)
		if err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
			errors = append(errors, err)
			continue
		}

		renderLog.Info("cleaning up object", "kind", resource.gvk.String(), "namespace", namespace, "name", resource.name)

		err = client.Delete(ctx, existing)
		if err != nil {
			renderLog.Error(err, "failed cleaning up object", "kind", resource.gvk.String(), "namespace", namespace, "name", resource.name)
			errors = append(errors, err)
		}
	}

	return errors
}
