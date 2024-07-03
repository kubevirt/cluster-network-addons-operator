package network

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/render"
)

// renderKubevirtIPAMController generates the manifests of kubevirt-ipam-controller
func renderKubevirtIPAMController(conf *cnao.NetworkAddonsConfigSpec, manifestDir string, clusterInfo *ClusterInfo) ([]*unstructured.Unstructured, error) {
	if conf.KubevirtIpamController == nil {
		return nil, nil
	}

	// render the manifests on disk
	data := render.MakeRenderData()
	data.Data["Namespace"] = os.Getenv("OPERAND_NAMESPACE")
	data.Data["ImagePullPolicy"] = conf.ImagePullPolicy
	data.Data["Placement"] = conf.PlacementConfiguration.Workloads
	data.Data["KubevirtIpamControllerImage"] = os.Getenv("KUBEVIRT_IPAM_CONTROLLER_IMAGE")

	if clusterInfo.OpenShift4 {
		data.Data["WebhookAnnotation"] = `service.beta.openshift.io/inject-cabundle: "true"`
		data.Data["CertDir"] = "/etc/ipam-controller/certificates"
		data.Data["MountPath"] = data.Data["CertDir"]
		data.Data["SecretName"] = "kubevirt-ipam-controller-webhook-service"
	} else {
		data.Data["WebhookAnnotation"] =
			"cert-manager.io/inject-ca-from: " + os.Getenv("OPERAND_NAMESPACE") + "/kubevirt-ipam-controller-serving-cert"
		data.Data["CertDir"] = ""
		data.Data["MountPath"] = "/tmp/k8s-webhook-server/serving-certs"
		data.Data["SecretName"] = "webhook-server-cert"
	}
	data.Data["IsOpenshift"] = clusterInfo.OpenShift4

	objs, err := render.RenderDir(filepath.Join(manifestDir, "kubevirt-ipam-controller"), &data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to render kubevirt-ipam-controller state handler manifests")
	}

	return objs, nil
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
