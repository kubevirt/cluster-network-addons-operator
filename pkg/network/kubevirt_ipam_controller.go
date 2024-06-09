package network

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

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
		data.Data["SecretName"] = "kubevirt-ipam-claims-webhook-service"
	} else {
		data.Data["WebhookAnnotation"] =
			"cert-manager.io/inject-ca-from: " + os.Getenv("OPERAND_NAMESPACE") + "/kubevirt-ipam-claims-serving-cert"
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
