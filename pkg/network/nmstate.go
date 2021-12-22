package network

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/kubevirt/cluster-network-addons-operator/pkg/render"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
)

// renderNMState generates the manifests of NMState handler
func renderNMState(conf *cnao.NetworkAddonsConfigSpec, manifestDir string, clusterInfo *ClusterInfo) ([]*unstructured.Unstructured, error) {
	if conf.NMState == nil {
		return nil, nil
	}

	// render the manifests on disk
	data := render.MakeRenderData()
	data.Data["HandlerPrefix"] = ""
	data.Data["HandlerNamespace"] = os.Getenv("OPERAND_NAMESPACE")
	data.Data["HandlerImage"] = os.Getenv("NMSTATE_HANDLER_IMAGE")
	data.Data["HandlerPullPolicy"] = conf.ImagePullPolicy
	data.Data["HandlerNodeSelector"] = map[string]string{}
	data.Data["EnableSCC"] = clusterInfo.SCCAvailable
	data.Data["CARotateInterval"] = conf.SelfSignConfiguration.CARotateInterval
	data.Data["CAOverlapInterval"] = conf.SelfSignConfiguration.CAOverlapInterval
	data.Data["CertRotateInterval"] = conf.SelfSignConfiguration.CertRotateInterval
	data.Data["CertOverlapInterval"] = conf.SelfSignConfiguration.CertOverlapInterval
	data.Data["PlacementConfiguration"] = conf.PlacementConfiguration

	log.Printf("NMStateOperator == %t", clusterInfo.NmstateOperator)
	fullManifestDir := filepath.Join(manifestDir, "nmstate", "operand")
	if clusterInfo.NmstateOperator {
		fullManifestDir = filepath.Join(manifestDir, "nmstate", "operator")
	}
	log.Printf("Rendering NMState directory: %s", fullManifestDir)

	objs, err := render.RenderDir(fullManifestDir, &data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to render nmstate state handler manifests")
	}

	return objs, nil
}

func removeAppsV1Resource(ctx context.Context, client k8sclient.Client, name, namespace, kind string) []error {
	// Get existing
	existing := &unstructured.Unstructured{}
	gvk := schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: kind}
	existing.SetGroupVersionKind(gvk)

	err := client.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, existing)

	// if we found the object
	if err == nil {
		objDesc := fmt.Sprintf("(%s) %s/%s", gvk.String(), namespace, name)
		log.Printf("Cleanup up %s Object", objDesc)

		// Delete the object
		err = client.Delete(ctx, existing)
		if err != nil {
			log.Printf("Failed Cleaning up %s Object", objDesc)
			return []error{err}
		}
	}
	return []error{}
}

func removeDaemonSetHandlerWorker(ctx context.Context, client k8sclient.Client) []error {
	// Get existing
	existing := &unstructured.Unstructured{}
	gvk := schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "DaemonSet"}
	existing.SetGroupVersionKind(gvk)
	namespace := os.Getenv("OPERAND_NAMESPACE")
	name := "nmstate-handler-worker"

	err := client.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, existing)
	// if we found the object
	if err == nil {
		objDesc := fmt.Sprintf("(%s) %s/%s", gvk.String(), namespace, name)
		log.Printf("Cleaning up %s Object", objDesc)

		// Delete the object
		err = client.Delete(ctx, existing)
		if err != nil {
			log.Printf("Failed Cleaning up %s Object", objDesc)
			return []error{err}
		}

	} else if apierrors.IsNotFound(err) {
		// object not found, no need for action.
		return []error{}
	}
	return []error{}
}

func removeStandaloneHandler(ctx context.Context, client k8sclient.Client) []error {
	namespace := os.Getenv("OPERAND_NAMESPACE")
	name := "nmstate-handler"
	kind := "DaemonSet"

	return removeAppsV1Resource(ctx, client, name, namespace, kind)
}

func removeStandaloneWebhook(ctx context.Context, client k8sclient.Client) []error {
	namespace := os.Getenv("OPERAND_NAMESPACE")
	name := "nmstate-webhook"
	kind := "Deployment"

	return removeAppsV1Resource(ctx, client, name, namespace, kind)
}

func removeStandaloneCertManager(ctx context.Context, client k8sclient.Client) []error {
	namespace := os.Getenv("OPERAND_NAMESPACE")
	name := "nmstate-cert-manager"
	kind := "Deployment"

	return removeAppsV1Resource(ctx, client, name, namespace, kind)
}

func cleanUpNMState(conf *cnao.NetworkAddonsConfigSpec, ctx context.Context, client k8sclient.Client, clusterInfo *ClusterInfo) []error {
	if conf.NMState == nil {
		return []error{}
	}

	errList := []error{}
	errList = append(errList, removeDaemonSetHandlerWorker(ctx, client)...)
	if clusterInfo.NmstateOperator {
		errList = append(errList, removeStandaloneHandler(ctx, client)...)
		errList = append(errList, removeStandaloneWebhook(ctx, client)...)
		errList = append(errList, removeStandaloneCertManager(ctx, client)...)
	}

	return errList
}
