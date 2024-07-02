package network

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	osv1 "github.com/openshift/api/operator/v1"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/network/cni"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/render"
)

// ValidateMultus validates the combination of DisableMultiNetwork and AddtionalNetworks
func validateMultus(conf *cnao.NetworkAddonsConfigSpec, openshiftNetworkConfig *osv1.Network) []error {
	if conf.Multus == nil {
		return []error{}
	}

	if openshiftNetworkConfig != nil {
		if openshiftNetworkConfig.Spec.DisableMultiNetwork != nil && *openshiftNetworkConfig.Spec.DisableMultiNetwork == true {
			return []error{errors.Errorf("multus has been requested, but is disabled on OpenShift Cluster Network Operator")}
		}
	}

	return []error{}
}

// cleanUpMultus checks specific multus outdated objects or ones that are no longer compatible and deletes them.
func cleanUpMultus(conf *cnao.NetworkAddonsConfigSpec, ctx context.Context, client k8sclient.Client) []error {
	if conf.Multus == nil {
		return []error{}
	}

	errList := []error{}
	errList = append(errList, cleanUpMultusOldName(ctx, client)...)
	return errList
}

// cleanUpMultusOldName deletes multus ds object with old name after a new name was introduces in version 0.25.0.
// REQUIRED_FOR upgrade from multus <= 0.25.0.
func cleanUpMultusOldName(ctx context.Context, client k8sclient.Client) []error {
	// Get existing
	existing := &unstructured.Unstructured{}
	gvk := schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "DaemonSet"}
	existing.SetGroupVersionKind(gvk)
	namespace := os.Getenv("OPERAND_NAMESPACE")
	name := "kube-multus-ds-amd64"

	err := client.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, existing)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// object not found, no need for action.
			return []error{}
		}
		return []error{err}
	}

	// if we found the object
	objDesc := fmt.Sprintf("(%s) %s/%s", gvk.String(), namespace, name)
	log.Printf("Cleaning up %s Object", objDesc)

	// Delete the object
	err = client.Delete(ctx, existing)
	if err != nil {
		log.Printf("Failed Cleaning up %s Object", objDesc)
		return []error{err}
	}

	return []error{}
}

// cleanUpKubevirtIpamController checks specific kic outdated objects or ones that are no longer compatible and deletes them.
func cleanUpKubevirtIpamController(conf *cnao.NetworkAddonsConfigSpec, ctx context.Context, client k8sclient.Client) []error {
	if conf.KubevirtIpamController == nil {
		return []error{}
	}

	errList := []error{}
	errList = append(errList, cleanUpKubevirtIpamControllerOldNames(ctx, client)...)
	return errList
}

// cleanUpKubevirtIpamControllerOldNames deletes kic objects with old name after a new name was introduces in version 0.94.1
// REQUIRED_FOR upgrade from kubevirt-ipam-controller == 0.94.0
func cleanUpKubevirtIpamControllerOldNames(ctx context.Context, client k8sclient.Client) []error {
	namespace := os.Getenv("OPERAND_NAMESPACE")

	resources := []struct {
		gvk  schema.GroupVersionKind
		name string
	}{
		{
			gvk:  schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
			name: "kubevirt-ipam-claims-controller-manager",
		},
		{
			gvk:  schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ServiceAccount"},
			name: "kubevirt-ipam-claims-controller-manager",
		},
		{
			gvk:  schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "Role"},
			name: "kubevirt-ipam-claims-leader-election-role",
		},
		{
			gvk:  schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "ClusterRole"},
			name: "kubevirt-ipam-claims-manager-role",
		},
		{
			gvk:  schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "RoleBinding"},
			name: "kubevirt-ipam-claims-leader-election-rolebinding",
		},
		{
			gvk:  schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "ClusterRoleBinding"},
			name: "kubevirt-ipam-claims-manager-rolebinding",
		},
		{
			gvk:  schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Service"},
			name: "kubevirt-ipam-claims-webhook-service",
		},
		{
			gvk:  schema.GroupVersionKind{Group: "cert-manager.io", Version: "v1", Kind: "Certificate"},
			name: "kubevirt-ipam-claims-serving-cert",
		},
		{
			gvk:  schema.GroupVersionKind{Group: "cert-manager.io", Version: "v1", Kind: "Issuer"},
			name: "kubevirt-ipam-claims-selfsigned-issuer",
		},
		{
			gvk:  schema.GroupVersionKind{Group: "admissionregistration.k8s.io", Version: "v1", Kind: "MutatingWebhookConfiguration"},
			name: "kubevirt-ipam-claims-mutating-webhook-configuration",
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

		objDesc := fmt.Sprintf("(%s) %s/%s", resource.gvk.String(), namespace, resource.name)
		log.Printf("Cleaning up %s Object", objDesc)

		err = client.Delete(ctx, existing)
		if err != nil {
			log.Printf("Failed Cleaning up %s Object", objDesc)
			errors = append(errors, err)
		}
	}

	return errors
}

// RenderMultus generates the manifests of Multus
func renderMultus(conf *cnao.NetworkAddonsConfigSpec, manifestDir string, openshiftNetworkConfig *osv1.Network, clusterInfo *ClusterInfo) ([]*unstructured.Unstructured, error) {
	if conf.Multus == nil || openshiftNetworkConfig != nil {
		return nil, nil
	}

	// render manifests from disk
	data := render.MakeRenderData()
	data.Data["Namespace"] = os.Getenv("OPERAND_NAMESPACE")
	data.Data["MultusImage"] = os.Getenv("MULTUS_IMAGE")
	data.Data["ImagePullPolicy"] = conf.ImagePullPolicy
	data.Data["Placement"] = conf.PlacementConfiguration.Workloads
	if clusterInfo.OpenShift4 {
		data.Data["CNIConfigDir"] = cni.ConfigDirOpenShift4
		data.Data["CNIBinDir"] = cni.BinDirOpenShift4
	} else {
		data.Data["CNIConfigDir"] = cni.ConfigDir
		data.Data["CNIBinDir"] = cni.BinDir
	}
	data.Data["EnableSCC"] = clusterInfo.SCCAvailable

	objs, err := render.RenderDir(filepath.Join(manifestDir, "multus"), &data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to render multus manifests")
	}

	return objs, nil
}
