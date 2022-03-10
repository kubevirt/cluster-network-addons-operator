package network

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/kubevirt/cluster-network-addons-operator/pkg/render"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
)

const (
	kubernetesOsSelector = "kubernetes.io/os"
)

// renderNMState generates the manifests of NMState handler
func renderNMState(conf *cnao.NetworkAddonsConfigSpec, manifestDir string, clusterInfo *ClusterInfo) ([]*unstructured.Unstructured, error) {
	if conf.NMState == nil {
		return nil, nil
	}

	confWithDefaults := pupulateNmstateDefaultConfiguration(conf)

	// render the manifests on disk
	data := render.MakeRenderData()
	data.Data["HandlerPrefix"] = ""
	data.Data["HandlerNamespace"] = os.Getenv("OPERAND_NAMESPACE")
	data.Data["HandlerImage"] = os.Getenv("NMSTATE_HANDLER_IMAGE")
	data.Data["HandlerPullPolicy"] = confWithDefaults.ImagePullPolicy
	data.Data["EnableSCC"] = clusterInfo.SCCAvailable
	data.Data["CARotateInterval"] = confWithDefaults.SelfSignConfiguration.CARotateInterval
	data.Data["CAOverlapInterval"] = confWithDefaults.SelfSignConfiguration.CAOverlapInterval
	data.Data["CertRotateInterval"] = confWithDefaults.SelfSignConfiguration.CertRotateInterval
	data.Data["CertOverlapInterval"] = confWithDefaults.SelfSignConfiguration.CertOverlapInterval
	data.Data["PlacementConfiguration"] = confWithDefaults.PlacementConfiguration
	data.Data["WebhookReplicas"] = getNumberOfWebhookReplicas(clusterInfo)
	data.Data["WebhookMinReplicas"] = getMinNumberOfWebhookReplicas(clusterInfo)
	data.Data["HandlerNodeSelector"] = confWithDefaults.PlacementConfiguration.Workloads.NodeSelector
	data.Data["HandlerTolerations"] = confWithDefaults.PlacementConfiguration.Workloads.Tolerations
	data.Data["HandlerAffinity"] = confWithDefaults.PlacementConfiguration.Workloads.Affinity
	data.Data["InfraNodeSelector"] = confWithDefaults.PlacementConfiguration.Infra.NodeSelector
	data.Data["InfraTolerations"] = confWithDefaults.PlacementConfiguration.Infra.Tolerations
	data.Data["InfraAffinity"] = confWithDefaults.PlacementConfiguration.Infra.Affinity

	_, enableOVS := os.LookupEnv("NMSTATE_ENABLE_OVS")
	data.Data["EnableOVS"] = enableOVS

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
	return removeResource(ctx, client, types.NamespacedName{Namespace: namespace, Name: name}, schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: kind})
}

func removeCoreV1Resource(ctx context.Context, client k8sclient.Client, name, namespace, kind string) []error {
	return removeResource(ctx, client, types.NamespacedName{Namespace: namespace, Name: name}, schema.GroupVersionKind{Group: "", Version: "v1", Kind: kind})
}

func removeResource(ctx context.Context, client k8sclient.Client, key types.NamespacedName, gvk schema.GroupVersionKind) []error {
	// Get existing
	existing := &unstructured.Unstructured{}
	existing.SetGroupVersionKind(gvk)

	err := client.Get(ctx, key, existing)

	// if we found the object
	if err == nil {
		objDesc := fmt.Sprintf("(%s) %s", gvk.String(), key)
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
	namespace := os.Getenv("OPERAND_NAMESPACE")
	name := "nmstate-handler-worker"
	kind := "DaemonSet"

	return removeAppsV1Resource(ctx, client, name, namespace, kind)
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

func removeConfig(ctx context.Context, client k8sclient.Client) []error {
	namespace := os.Getenv("OPERAND_NAMESPACE")
	name := "nmstate-config"
	kind := "ConfigMap"

	return removeCoreV1Resource(ctx, client, name, namespace, kind)
}

func cleanUpNMState(conf *cnao.NetworkAddonsConfigSpec, ctx context.Context, client k8sclient.Client, clusterInfo *ClusterInfo) []error {
	if conf.NMState == nil {
		return []error{}
	}

	errList := []error{}
	errList = append(errList, removeDaemonSetHandlerWorker(ctx, client)...)
	errList = append(errList, removeConfig(ctx, client)...)
	if clusterInfo.NmstateOperator {
		errList = append(errList, removeStandaloneHandler(ctx, client)...)
		errList = append(errList, removeStandaloneWebhook(ctx, client)...)
		errList = append(errList, removeStandaloneCertManager(ctx, client)...)
	}

	return errList
}

func getNumberOfWebhookReplicas(clusterInfo *ClusterInfo) int32 {
	const (
		highlyAvailableWebhookReplicas = int32(2)
		singleReplicaWebhookReplicas   = int32(1)
	)

	if clusterInfo.OpenShift4 && clusterInfo.IsSingleReplica {
		return singleReplicaWebhookReplicas
	}

	return highlyAvailableWebhookReplicas
}

func getMinNumberOfWebhookReplicas(clusterInfo *ClusterInfo) int32 {
	const (
		highlyAvailableMinWebhookReplicas = int32(1)
		singleReplicaMinWebhookReplicas   = int32(0)
	)

	if clusterInfo.OpenShift4 && clusterInfo.IsSingleReplica {
		return singleReplicaMinWebhookReplicas
	}

	return highlyAvailableMinWebhookReplicas
}

func pupulateNmstateDefaultConfiguration(conf *cnao.NetworkAddonsConfigSpec) *cnao.NetworkAddonsConfigSpec {
	confWithDefaults := conf.DeepCopy()
	if confWithDefaults.PlacementConfiguration.Workloads.NodeSelector == nil {
		confWithDefaults.PlacementConfiguration.Workloads.NodeSelector = map[string]string{}
	}
	if _, present := confWithDefaults.PlacementConfiguration.Workloads.NodeSelector[kubernetesOsSelector]; !present {
		confWithDefaults.PlacementConfiguration.Workloads.NodeSelector[kubernetesOsSelector] = "linux"
	}

	return confWithDefaults
}
