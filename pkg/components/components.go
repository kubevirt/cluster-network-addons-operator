package components

import (
	"fmt"
	"os"
	"regexp"

	ocpv1 "github.com/openshift/api/config/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
	cnaov1 "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/v1"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/monitoring/rules/alerts"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/names"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/util/k8s"
)

const (
	Name      = "cluster-network-addons-operator"
	Namespace = "cluster-network-addons"
)

var (
	imageSplitRe = regexp.MustCompile(`(?:.+/)*([^/:@]+)(?:[:@]?.*)?`)
)

const (
	MultusImageDefault                 = "ghcr.io/k8snetworkplumbingwg/multus-cni@sha256:3fbcc32bd4e4d15bd93c96def784a229cd84cca27942bf4858b581f31c97ee02"
	MultusDynamicNetworksImageDefault  = "ghcr.io/k8snetworkplumbingwg/multus-dynamic-networks-controller@sha256:83b460502671fb4f34116363a1a39b2ddfc9d14a920ee0a6413bfc3bd0580404"
	LinuxBridgeCniImageDefault         = "quay.io/kubevirt/cni-default-plugins@sha256:0c354fa9d695b8cab97b459e8afea2f7662407a987e83f6f6f1a8af4b45726be"
	LinuxBridgeMarkerImageDefault      = "quay.io/kubevirt/bridge-marker@sha256:bba066e3b5ff3fb8c5e20861fe8abe51e3c9b50ad6ce3b2616af9cb5479a06d0"
	KubeMacPoolImageDefault            = "quay.io/kubevirt/kubemacpool@sha256:20e156be33e6d3692c456081acbb91a4349f94de448f7d1f1cddd0228931b31f"
	OvsCniImageDefault                 = "quay.io/kubevirt/ovs-cni-plugin@sha256:e16ac74343da21abb8fb668ce71e728053d00503a992dae2164b9e94a280113e"
	MacvtapCniImageDefault             = "quay.io/kubevirt/macvtap-cni@sha256:850b89343ace7c7ea6b18dd8e11964613974e9d1f7377af03854d407fb15230a"
	KubeRbacProxyImageDefault          = "quay.io/openshift/origin-kube-rbac-proxy@sha256:e2def4213ec0657e72eb790ae8a115511d5b8f164a62d3568d2f1bff189917e8"
	KubeSecondaryDNSImageDefault       = "ghcr.io/kubevirt/kubesecondarydns@sha256:6268d84154e2483fbce8c1adacbdaf6f0839117b2d48d9fa4687cc8f76bd5130"
	CoreDNSImageDefault                = "registry.k8s.io/coredns/coredns@sha256:a0ead06651cf580044aeb0a0feba63591858fb2e43ade8c9dea45a6a89ae7e5e"
	KubevirtIpamControllerImageDefault = "ghcr.io/maiqueb/kubevirt-ipam-claims@sha256:4667b37655dc87b79a63ca0b127517d0f9b971808b8fda8f5ea01398e8833294"
)

type AddonsImages struct {
	Multus                 string
	MultusDynamicNetworks  string
	LinuxBridgeCni         string
	LinuxBridgeMarker      string
	KubeMacPool            string
	OvsCni                 string
	MacvtapCni             string
	KubeRbacProxy          string
	KubeSecondaryDNS       string
	CoreDNS                string
	KubevirtIpamController string
}

type RelatedImage struct {
	Name string
	Ref  string
}

type RelatedImages []RelatedImage

func NewRelatedImages(images ...string) RelatedImages {
	ris := RelatedImages{}
	for _, image := range images {
		ris = append(ris, NewRelatedImage(image))
	}

	return ris
}

func (ris *RelatedImages) Add(image string) {
	ri := NewRelatedImage(image)
	*ris = append(*ris, ri)
}

func (ai *AddonsImages) FillDefaults() *AddonsImages {
	if ai.Multus == "" {
		ai.Multus = MultusImageDefault
	}
	if ai.MultusDynamicNetworks == "" {
		ai.Multus = MultusDynamicNetworksImageDefault
	}
	if ai.LinuxBridgeCni == "" {
		ai.LinuxBridgeCni = LinuxBridgeCniImageDefault
	}
	if ai.LinuxBridgeMarker == "" {
		ai.LinuxBridgeMarker = LinuxBridgeMarkerImageDefault
	}
	if ai.KubeMacPool == "" {
		ai.KubeMacPool = KubeMacPoolImageDefault
	}
	if ai.OvsCni == "" {
		ai.OvsCni = OvsCniImageDefault
	}
	if ai.MacvtapCni == "" {
		ai.MacvtapCni = MacvtapCniImageDefault
	}
	if ai.KubeRbacProxy == "" {
		ai.KubeRbacProxy = KubeRbacProxyImageDefault
	}
	if ai.KubeSecondaryDNS == "" {
		ai.KubeSecondaryDNS = KubeSecondaryDNSImageDefault
	}
	if ai.CoreDNS == "" {
		ai.CoreDNS = CoreDNSImageDefault
	}
	if ai.KubevirtIpamController == "" {
		ai.KubevirtIpamController = KubevirtIpamControllerImageDefault
	}
	return ai
}

func (ai AddonsImages) ToRelatedImages() RelatedImages {
	return NewRelatedImages(
		ai.Multus,
		ai.MultusDynamicNetworks,
		ai.LinuxBridgeCni,
		ai.LinuxBridgeMarker,
		ai.KubeMacPool,
		ai.OvsCni,
		ai.MacvtapCni,
		ai.KubeRbacProxy,
		ai.KubeSecondaryDNS,
		ai.CoreDNS,
		ai.KubevirtIpamController,
	)
}

func NewRelatedImage(image string) RelatedImage {
	// find the basic image name - with no registry and tag
	name := image
	if names := imageSplitRe.FindStringSubmatch(image); len(names) > 1 {
		name = names[1]
	}

	return RelatedImage{
		Name: name,
		Ref:  image,
	}
}

func GetDeployment(version string, operatorVersion string, namespace string, repository string, imageName string, tag string, imagePullPolicy string, addonsImages *AddonsImages) *appsv1.Deployment {
	image := fmt.Sprintf("%s/%s:%s", repository, imageName, tag)
	runAsNonRoot := true
	allowPrivilegeEscalation := false

	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      Name,
			Namespace: namespace,
			Annotations: map[string]string{
				cnaov1.GroupVersion.Group + "/version": k8s.StringToLabel(operatorVersion),
			},
			Labels: map[string]string{
				names.PROMETHEUS_LABEL_KEY: names.PROMETHEUS_LABEL_VALUE,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"name": Name,
				},
			},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RecreateDeploymentStrategyType,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"name":                     Name,
						names.PROMETHEUS_LABEL_KEY: names.PROMETHEUS_LABEL_VALUE,
					},
					Annotations: map[string]string{
						"description": "cluster-network-addons-operator manages the lifecycle of different Kubernetes network components on top of Kubernetes cluster",
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: Name,
					SecurityContext: &corev1.PodSecurityContext{
						RunAsNonRoot: &runAsNonRoot,
						SeccompProfile: &corev1.SeccompProfile{
							Type: corev1.SeccompProfileTypeRuntimeDefault,
						},
					},
					PriorityClassName: "system-cluster-critical",
					Containers: []corev1.Container{
						{
							Name:            Name,
							Image:           image,
							ImagePullPolicy: corev1.PullPolicy(imagePullPolicy),
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("50m"),
									corev1.ResourceMemory: resource.MustParse("30Mi"),
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "MULTUS_IMAGE",
									Value: addonsImages.Multus,
								},
								{
									Name:  "MULTUS_DYNAMIC_NETWORKS_CONTROLLER_IMAGE",
									Value: addonsImages.MultusDynamicNetworks,
								},
								{
									Name:  "LINUX_BRIDGE_IMAGE",
									Value: addonsImages.LinuxBridgeCni,
								},
								{
									Name:  "LINUX_BRIDGE_MARKER_IMAGE",
									Value: addonsImages.LinuxBridgeMarker,
								},
								{
									Name:  "OVS_CNI_IMAGE",
									Value: addonsImages.OvsCni,
								},
								{
									Name:  "KUBEMACPOOL_IMAGE",
									Value: addonsImages.KubeMacPool,
								},
								{
									Name:  "MACVTAP_CNI_IMAGE",
									Value: addonsImages.MacvtapCni,
								},
								{
									Name:  "KUBE_RBAC_PROXY_IMAGE",
									Value: addonsImages.KubeRbacProxy,
								},
								{
									Name:  "KUBE_SECONDARY_DNS_IMAGE",
									Value: addonsImages.KubeSecondaryDNS,
								},
								{
									Name:  "CORE_DNS_IMAGE",
									Value: addonsImages.CoreDNS,
								},
								{
									Name:  "KUBEVIRT_IPAM_CONTROLLER_IMAGE",
									Value: addonsImages.KubevirtIpamController,
								},
								{
									Name:  "OPERATOR_IMAGE",
									Value: image,
								},
								{
									Name:  "OPERATOR_NAME",
									Value: Name,
								},
								{
									Name:  "OPERATOR_VERSION",
									Value: operatorVersion,
								},
								{
									Name: "OPERATOR_NAMESPACE",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
								{
									Name: "OPERAND_NAMESPACE",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
								{
									Name: "POD_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
								{
									Name:  "WATCH_NAMESPACE",
									Value: "",
								},
								{
									Name:  "MONITORING_NAMESPACE",
									Value: getMonitoringNamespace(),
								},
								{
									Name:  "MONITORING_SERVICE_ACCOUNT",
									Value: "prometheus-k8s",
								},
								{
									Name:  "RUNBOOK_URL_TEMPLATE",
									Value: alerts.GetRunbookURLTemplate(),
								},
							},
							SecurityContext: &corev1.SecurityContext{
								AllowPrivilegeEscalation: &allowPrivilegeEscalation,
								Capabilities: &corev1.Capabilities{
									Drop: []corev1.Capability{corev1.Capability("ALL")},
								},
							},
						},
						{
							Name:            "kube-rbac-proxy",
							Image:           addonsImages.KubeRbacProxy,
							ImagePullPolicy: corev1.PullPolicy(imagePullPolicy),
							Ports: []corev1.ContainerPort{
								corev1.ContainerPort{
									Name:          "metrics",
									Protocol:      "TCP",
									ContainerPort: 8443,
								},
							},
							Args: []string{
								"--logtostderr",
								"--secure-listen-address=:8443",
								"--upstream=http://127.0.0.1:8080",
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("10m"),
									corev1.ResourceMemory: resource.MustParse("20Mi"),
								},
							},
							TerminationMessagePolicy: corev1.TerminationMessageFallbackToLogsOnError,
							SecurityContext: &corev1.SecurityContext{
								AllowPrivilegeEscalation: &allowPrivilegeEscalation,
								Capabilities: &corev1.Capabilities{
									Drop: []corev1.Capability{corev1.Capability("ALL")},
								},
							},
						},
					},
				},
			},
		},
	}
	return deployment
}

func GetRole(namespace string) *rbacv1.Role {
	role := &rbacv1.Role{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "Role",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      Name,
			Namespace: namespace,
			Labels: map[string]string{
				"name": Name,
			},
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{
					"apps",
				},
				Resources: []string{
					"daemonsets",
				},
				Verbs: []string{
					"get",
					"create",
					"update",
					"delete",
				},
			},
			{
				APIGroups: []string{
					"",
				},
				Resources: []string{
					"configmaps",
				},
				Verbs: []string{
					"get",
					"create",
					"update",
				},
			},
			{
				APIGroups: []string{
					"apps",
				},
				Resources: []string{
					"deployments",
				},
				Verbs: []string{
					"delete",
				},
			},
			{
				APIGroups: []string{
					"",
				},
				Resources: []string{
					"namespaces",
				},
				Verbs: []string{
					"update",
					"get",
					"patch",
				},
			},
			{
				APIGroups: []string{
					"",
				},
				Resources: []string{
					"serviceaccounts",
				},
				Verbs: []string{
					"get",
					"create",
					"update",
					"delete",
				},
			},
			{
				APIGroups: []string{
					"monitoring.coreos.com",
				},
				Resources: []string{
					"prometheusrules",
					"servicemonitors",
				},
				Verbs: []string{
					"get",
					"create",
					"update",
					"delete",
				},
			},
			{
				APIGroups: []string{
					"rbac.authorization.k8s.io",
				},
				Resources: []string{
					"roles",
					"rolebindings",
				},
				Verbs: []string{
					"get",
					"create",
					"update",
					"delete",
				},
			},
			{
				APIGroups: []string{
					"policy",
				},
				Resources: []string{
					"poddisruptionbudgets",
				},
				Verbs: []string{
					"get",
					"delete",
				},
			},
			{
				APIGroups: []string{
					"",
				},
				Resources: []string{
					"configmaps",
				},
				Verbs: []string{
					"patch",
				},
			},
			{
				APIGroups: []string{
					"coordination.k8s.io",
				},
				Resources: []string{
					"leases",
				},
				Verbs: []string{
					"get",
					"list",
					"watch",
					"create",
					"update",
					"patch",
					"delete",
				},
			},
			{
				APIGroups: []string{
					"cert-manager.io",
				},
				Resources: []string{
					"certificates",
					"issuers",
				},
				Verbs: []string{
					"get",
					"create",
					"update",
					"delete",
				},
			},
		},
	}
	return role
}

func GetClusterRole(allowMultus bool) *rbacv1.ClusterRole {
	role := &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "ClusterRole",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: Name,
			Labels: map[string]string{
				"name": Name,
			},
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{
					"operator.openshift.io",
				},
				Resources: []string{
					"networks",
				},
				Verbs: []string{
					"list",
					"watch",
				},
			},
			{
				APIGroups: []string{
					"security.openshift.io",
				},
				Resources: []string{
					"securitycontextconstraints",
				},
				Verbs: []string{
					"get",
					"list",
					"create",
					"update",
				},
			},
			{
				APIGroups: []string{
					"apiextensions.k8s.io",
				},
				Resources: []string{
					"customresourcedefinitions",
				},
				Verbs: []string{
					"get",
					"create",
					"update",
				},
			},
			{
				APIGroups: []string{
					"networkaddonsoperator.network.kubevirt.io",
				},
				Resources: []string{
					"networkaddonsconfigs",
				},
				Verbs: []string{
					"list",
					"watch",
				},
			},
			{
				APIGroups: []string{
					"networkaddonsoperator.network.kubevirt.io",
				},
				Resources: []string{
					"networkaddonsconfigs/status",
				},
				Verbs: []string{
					"patch",
				},
			},
			{
				APIGroups: []string{
					"networkaddonsoperator.network.kubevirt.io",
				},
				Resources: []string{
					"networkaddonsconfigs/finalizers",
				},
				Verbs: []string{
					"update",
				},
			},
			{
				APIGroups: []string{
					"apps",
				},
				Resources: []string{
					"deployments",
					"daemonsets",
				},
				Verbs: []string{
					"list",
					"watch",
				},
			},
			{
				APIGroups: []string{
					"",
				},
				Resources: []string{
					"configmaps",
					"namespaces",
				},
				Verbs: []string{
					"list",
					"watch",
				},
			},
			{
				APIGroups: []string{
					"rbac.authorization.k8s.io",
				},
				Resources: []string{
					"clusterroles",
				},
				Verbs: []string{
					"get",
					"create",
					"update",
					"bind",
					"delete",
				},
			},
			{
				APIGroups: []string{
					"rbac.authorization.k8s.io",
				},
				Resources: []string{
					"clusterrolebindings",
				},
				Verbs: []string{
					"get",
					"create",
					"update",
					"delete",
				},
			},
			{
				APIGroups: []string{
					"",
				},
				Resources: []string{
					"events",
				},
				Verbs: []string{
					"create",
					"patch",
				},
			},
			{
				APIGroups: []string{
					"admissionregistration.k8s.io",
				},
				Resources: []string{
					"mutatingwebhookconfigurations",
				},
				Verbs: []string{
					"get",
					"create",
					"update",
					"delete",
				},
			},
			{
				APIGroups: []string{
					"config.openshift.io",
				},
				Resources: []string{
					"infrastructures",
				},
				Verbs: []string{
					"list",
					"watch",
				},
			},
			{
				APIGroups: []string{
					"",
				},
				Resources: []string{
					"services",
				},
				Verbs: []string{
					"delete",
				},
			},
			{
				APIGroups: []string{
					"k8s.cni.cncf.io",
				},
				Resources: []string{
					"ipamclaims",
				},
				Verbs: []string{
					"get",
					"list",
					"watch",
					"create",
					"update",
				},
			},
			{
				APIGroups: []string{
					"k8s.cni.cncf.io",
				},
				Resources: []string{
					"network-attachment-definitions",
				},
				Verbs: []string{
					"get",
					"list",
					"watch",
				},
			},
		},
	}

	role.Rules = append(role.Rules, componentsClusterRoles()...)
	if allowMultus {
		role.Rules = append(role.Rules, multusClusterRoles()...)
	}

	return role
}

func componentsClusterRoles() []rbacv1.PolicyRule {
	return []rbacv1.PolicyRule{
		{
			APIGroups: []string{
				"",
			},
			Resources: []string{
				"events",
			},
			Verbs: []string{
				"update",
			},
		},
		{
			APIGroups: []string{
				"",
			},
			Resources: []string{
				"pods",
				"pods/status",
			},
			Verbs: []string{
				"get",
				"update",
				"list",
				"watch",
			},
		},
		{
			APIGroups: []string{
				"events.k8s.io",
			},
			Resources: []string{
				"events",
			},
			Verbs: []string{
				"create",
				"patch",
				"update",
			},
		},
		{
			APIGroups: []string{
				"",
			},
			Resources: []string{
				"nodes",
				"nodes/status",
			},
			Verbs: []string{
				"get",
				"update",
				"patch",
			},
		},
		{
			APIGroups: []string{
				"",
			},
			Resources: []string{
				"configmaps",
			},
			Verbs: []string{
				"get",
				"delete",
			},
		},
		{
			APIGroups: []string{
				"",
			},
			Resources: []string{
				"secrets",
			},
			Verbs: []string{
				"list",
				"watch",
				"create",
				"update",
			},
		},
		{
			APIGroups: []string{
				"admissionregistration.k8s.io",
			},
			Resources: []string{
				"validatingwebhookconfigurations",
				"mutatingwebhookconfigurations",
			},
			Verbs: []string{
				"list",
				"watch",
			},
		},
		{
			APIGroups: []string{
				"",
			},
			Resources: []string{
				"services",
			},
			Verbs: []string{
				"get",
				"create",
				"update",
				"list",
				"watch",
			},
		},
		{
			APIGroups: []string{
				"kubevirt.io",
			},
			Resources: []string{
				"virtualmachines",
			},
			Verbs: []string{
				"get",
				"list",
				"watch",
				"update",
			},
		},
		{
			APIGroups: []string{
				"authentication.k8s.io",
			},
			Resources: []string{
				"tokenreviews",
			},
			Verbs: []string{
				"create",
			},
		},
		{
			APIGroups: []string{
				"authorization.k8s.io",
			},
			Resources: []string{
				"subjectaccessreviews",
			},
			Verbs: []string{
				"create",
			},
		},
		{
			APIGroups: []string{
				"apps",
			},
			Resources: []string{
				"deployments",
			},
			Verbs: []string{
				"get",
				"create",
				"update",
			},
		},
		{
			APIGroups: []string{
				"kubevirt.io",
			},
			Resources: []string{
				"virtualmachineinstances",
			},
			Verbs: []string{
				"get",
				"list",
				"watch",
			},
		},
		{
			APIGroups: []string{
				"",
			},
			Resources: []string{
				"endpoints",
			},
			Verbs: []string{
				"get",
				"list",
				"watch",
			},
		},
	}
}

// Multus CNCF rules are kept separate and optional to allow vendors
// who ship Multus through different channels to not deploy Multus'
// highly privileged cluster-wide RBAC with CNAO.
func multusClusterRoles() []rbacv1.PolicyRule {
	return []rbacv1.PolicyRule{
		{
			APIGroups: []string{
				"k8s.cni.cncf.io",
			},
			Resources: []string{
				"*",
			},
			Verbs: []string{
				"*",
			},
		},
	}
}

func GetCrd() *extv1.CustomResourceDefinition {
	subResouceSchema := &extv1.CustomResourceSubresources{Status: &extv1.CustomResourceSubresourceStatus{}}
	labelSelectorRequirement := map[string]extv1.JSONSchemaProps{
		"key": extv1.JSONSchemaProps{
			Description: "key is the label key that the selector applies to.",
			Type:        "string",
		},
		"operator": extv1.JSONSchemaProps{
			Description: "operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.",
			Type:        "string",
		},
		"values": extv1.JSONSchemaProps{
			Description: "values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.",
			Type:        "array",
			Items: &extv1.JSONSchemaPropsOrArray{
				Schema: &extv1.JSONSchemaProps{
					Type: "string",
				},
			},
		},
	}
	podLabelSelector := extv1.JSONSchemaProps{
		Description: "A label query over a set of resources, in this case pods.",
		Type:        "object",
		Properties: map[string]extv1.JSONSchemaProps{
			"matchExpressions": extv1.JSONSchemaProps{
				Description: "matchExpressions is a list of label selector requirements. The requirements are ANDed.",
				Type:        "array",
				Items: &extv1.JSONSchemaPropsOrArray{
					Schema: &extv1.JSONSchemaProps{
						Description: "A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.",
						Type:        "object",
						Properties:  labelSelectorRequirement,
						Required: []string{
							"key",
							"operator",
						},
					},
				},
			},
			"matchLabels": extv1.JSONSchemaProps{
				Description: "matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is \"key\", the operator is \"In\", and the values array contains only \"value\". The requirements are ANDed.",
				Type:        "object",
				AdditionalProperties: &extv1.JSONSchemaPropsOrBool{
					Schema: &extv1.JSONSchemaProps{
						Type: "string",
					},
				},
			},
		},
	}
	getPodPreferredDuringSchedulingIgnoredDuringExecution := func(affinityPolarity string) extv1.JSONSchemaProps {
		return extv1.JSONSchemaProps{
			Description: fmt.Sprintf("The scheduler will prefer to schedule pods to nodes that satisfy the %s expressions specified by this field, but it may choose a node that violates one or more of the expressions. The node that is most preferred is the one with the greatest sum of weights, i.e. for each node that meets all of the scheduling requirements (resource request, requiredDuringScheduling %s expressions, etc.), compute a sum by iterating through the elements of this field and adding \"weight\" to the sum if the node has pods which matches the corresponding podAffinityTerm; the node(s) with the highest sum are the most preferred.", affinityPolarity, affinityPolarity),
			Type:        "array",
			Items: &extv1.JSONSchemaPropsOrArray{
				Schema: &extv1.JSONSchemaProps{
					Description: "The weights of all of the matched WeightedPodAffinityTerm fields are added per-node to find the most preferred node(s)",
					Type:        "object",
					Properties: map[string]extv1.JSONSchemaProps{
						"podAffinityTerm": extv1.JSONSchemaProps{
							Description: "Required. A pod affinity term, associated with the corresponding weight.",
							Type:        "object",
							Required: []string{
								"topologyKey",
							},
							Properties: map[string]extv1.JSONSchemaProps{
								"labelSelector": podLabelSelector,
								"namespaces": extv1.JSONSchemaProps{
									Description: "namespaces specifies which namespaces the labelSelector applies to (matches against); null or empty list means \"this pod's namespace\"",
									Type:        "array",
									Items: &extv1.JSONSchemaPropsOrArray{
										Schema: &extv1.JSONSchemaProps{
											Type: "string",
										},
									},
								},
								"topologyKey": extv1.JSONSchemaProps{
									Description: "This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching the labelSelector in the specified namespaces, where co-located is defined as running on a node whose value of the label with key topologyKey matches that of any node on which any of the selected pods is running. Empty topologyKey is not allowed.",
									Type:        "string",
								},
							},
						},
						"weight": extv1.JSONSchemaProps{
							Description: "weight associated with matching the corresponding podAffinityTerm, in the range 1-100.",
							Type:        "integer",
							Format:      "int32",
						},
					},
					Required: []string{
						"podAffinityTerm",
						"weight",
					},
				},
			},
		}
	}
	getPodRequiredDuringSchedulingIgnoredDuringExecution := func(affinityPolarity string) extv1.JSONSchemaProps {
		return extv1.JSONSchemaProps{
			Description: fmt.Sprintf("If the %s requirements specified by this field are not met at scheduling time, the pod will not be scheduled onto the node. If the %s requirements specified by this field cease to be met at some point during pod execution (e.g. due to a pod label update), the system may or may not try to eventually evict the pod from its node. When there are multiple elements, the lists of nodes corresponding to each podAffinityTerm are intersected, i.e. all terms must be satisfied.", affinityPolarity, affinityPolarity),
			Type:        "array",
			Items: &extv1.JSONSchemaPropsOrArray{
				Schema: &extv1.JSONSchemaProps{
					Description: "Defines a set of pods (namely those matching the labelSelector relative to the given namespace(s)) that this pod should be co-located (affinity) or not co-located (anti-affinity) with, where co-located is defined as running on a node whose value of the label with key <topologyKey> matches that of any node on which a pod of the set of pods is running",
					Type:        "object",
					Required: []string{
						"topologyKey",
					},
					Properties: map[string]extv1.JSONSchemaProps{
						"labelSelector": podLabelSelector,
						"namespaces": extv1.JSONSchemaProps{
							Description: "namespaces specifies which namespaces the labelSelector applies to (matches against); null or empty list means \"this pod's namespace\"",
							Type:        "array",
							Items: &extv1.JSONSchemaPropsOrArray{
								Schema: &extv1.JSONSchemaProps{
									Type: "string",
								},
							},
						},
						"topologyKey": extv1.JSONSchemaProps{
							Description: "This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching the labelSelector in the specified namespaces, where co-located is defined as running on a node whose value of the label with key topologyKey matches that of any node on which any of the selected pods is running. Empty topologyKey is not allowed.",
							Type:        "string",
						},
					},
				},
			},
		}
	}
	getNodeSelectorRequirement := func(description string) extv1.JSONSchemaProps {
		return extv1.JSONSchemaProps{
			Description: fmt.Sprintf("%s", description),
			Type:        "array",
			Items: &extv1.JSONSchemaPropsOrArray{
				Schema: &extv1.JSONSchemaProps{
					Description: "A node selector requirement is a selector that contains values, a key, and an operator that relates the key and values.",
					Type:        "object",
					Required: []string{
						"key",
						"operator",
					},
					Properties: map[string]extv1.JSONSchemaProps{
						"key": extv1.JSONSchemaProps{
							Description: "The label key that the selector applies to.",
							Type:        "string",
						},
						"operator": extv1.JSONSchemaProps{
							Description: "Represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.",
							Type:        "string",
						},
						"values": extv1.JSONSchemaProps{
							Description: "An array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be interpreted as an integer. This array is replaced during a strategic merge patch.",
							Type:        "array",
							Items: &extv1.JSONSchemaPropsOrArray{
								Schema: &extv1.JSONSchemaProps{
									Type: "string",
								},
							},
						},
					},
				},
			},
		}
	}

	cipherSuites := func() []extv1.JSON {
		suites := []extv1.JSON{}
		m := make(map[string]bool)
		for _, p := range tlsProfiles(ocpv1.TLSProfiles).sortedKeys() {
			for _, c := range ocpv1.TLSProfiles[p].Ciphers {
				if m[c] {
					continue
				}
				m[c] = true
				suites = append(suites, extv1.JSON{Raw: []byte(fmt.Sprintf(`"%s"`, c))})
			}
		}
		return suites
	}

	customSecurityProfileProps := map[string]extv1.JSONSchemaProps{
		"ciphers": extv1.JSONSchemaProps{
			Description: "ciphers is used to specify the cipher algorithms that are negotiated during the TLS handshake.  Operators may remove entries their operands do not support.  For example, to use DES-CBC3-SHA  (yaml):\n   ciphers:     - DES-CBC3-SHA",
			Items: &extv1.JSONSchemaPropsOrArray{
				Schema: &extv1.JSONSchemaProps{
					Type: "string",
					Enum: cipherSuites(),
				},
			},
			Type: "array",
		},
		"minTLSVersion": extv1.JSONSchemaProps{
			Description: "minTLSVersion is used to specify the minimal version of the TLS protocol that is negotiated during the TLS handshake. For example, to use TLS versions 1.1, 1.2 and 1.3 (yaml):\n   minTLSVersion: TLSv1.1\n NOTE: currently the highest minTLSVersion allowed is VersionTLS12",
			Type:        "string",
			Enum: []extv1.JSON{
				{Raw: []byte(fmt.Sprintf("\"%s\"", ocpv1.VersionTLS10))},
				{Raw: []byte(fmt.Sprintf("\"%s\"", ocpv1.VersionTLS11))},
				{Raw: []byte(fmt.Sprintf("\"%s\"", ocpv1.VersionTLS12))},
				{Raw: []byte(fmt.Sprintf("\"%s\"", ocpv1.VersionTLS13))},
			},
		},
	}

	placementProps := map[string]extv1.JSONSchemaProps{
		"nodeSelector": extv1.JSONSchemaProps{
			AdditionalProperties: &extv1.JSONSchemaPropsOrBool{
				Schema: &extv1.JSONSchemaProps{
					Type: "string",
				},
			},
			Type: "object",
		},
		"affinity": extv1.JSONSchemaProps{
			Description: "Affinity is a group of affinity scheduling rules.",
			Type:        "object",
			Properties: map[string]extv1.JSONSchemaProps{
				"nodeAffinity": extv1.JSONSchemaProps{
					Description: "Describes node affinity scheduling rules for the pod.",
					Type:        "object",
					Properties: map[string]extv1.JSONSchemaProps{
						"preferredDuringSchedulingIgnoredDuringExecution": extv1.JSONSchemaProps{
							Description: "The scheduler will prefer to schedule pods to nodes that satisfy the affinity expressions specified by this field, but it may choose a node that violates one or more of the expressions. The node that is most preferred is the one with the greatest sum of weights, i.e. for each node that meets all of the scheduling requirements (resource request, requiredDuringScheduling affinity expressions, etc.), compute a sum by iterating through the elements of this field and adding \"weight\" to the sum if the node matches the corresponding matchExpressions; the node(s) with the highest sum are the most preferred.",
							Type:        "array",
							Items: &extv1.JSONSchemaPropsOrArray{
								Schema: &extv1.JSONSchemaProps{
									Description: "An empty preferred scheduling term matches all objects with implicit weight 0 (i.e. it's a no-op). A null preferred scheduling term matches no objects (i.e. is also a no-op).",
									Type:        "object",
									Properties: map[string]extv1.JSONSchemaProps{
										"preference": extv1.JSONSchemaProps{
											Description: "A node selector term, associated with the corresponding weight.",
											Type:        "object",
											Properties: map[string]extv1.JSONSchemaProps{
												"matchExpressions": getNodeSelectorRequirement("A list of node selector requirements by node's labels."),
												"matchFields":      getNodeSelectorRequirement("A list of node selector requirements by node's fields."),
											},
										},
										"weight": extv1.JSONSchemaProps{
											Description: "Weight associated with matching the corresponding nodeSelectorTerm, in the range 1-100.",
											Type:        "integer",
											Format:      "int32",
										},
									},
									Required: []string{
										"preference",
										"weight",
									},
								},
							},
						},
						"requiredDuringSchedulingIgnoredDuringExecution": extv1.JSONSchemaProps{
							Description: "If the affinity requirements specified by this field are not met at scheduling time, the pod will not be scheduled onto the node. If the affinity requirements specified by this field cease to be met at some point during pod execution (e.g. due to an update), the system may or may not try to eventually evict the pod from its node.",
							Type:        "object",
							Required: []string{
								"nodeSelectorTerms",
							},
							Properties: map[string]extv1.JSONSchemaProps{
								"nodeSelectorTerms": extv1.JSONSchemaProps{
									Description: "Required. A list of node selector terms. The terms are ORed.",
									Type:        "array",
									Items: &extv1.JSONSchemaPropsOrArray{
										Schema: &extv1.JSONSchemaProps{
											Description: "A null or empty node selector term matches no objects. The requirements of them are ANDed. The TopologySelectorTerm type implements a subset of the NodeSelectorTerm.",
											Type:        "object",
											Properties: map[string]extv1.JSONSchemaProps{
												"matchExpressions": getNodeSelectorRequirement("A list of node selector requirements by node's labels."),
												"matchFields":      getNodeSelectorRequirement("A list of node selector requirements by node's fields."),
											},
										},
									},
								},
							},
						},
					},
				},
				"podAffinity": extv1.JSONSchemaProps{
					Description: "Describes pod affinity scheduling rules (e.g. co-locate this pod in the same node, zone, etc. as some other pod(s)).",
					Type:        "object",
					Properties: map[string]extv1.JSONSchemaProps{
						"preferredDuringSchedulingIgnoredDuringExecution": getPodPreferredDuringSchedulingIgnoredDuringExecution("affinity"),
						"requiredDuringSchedulingIgnoredDuringExecution":  getPodRequiredDuringSchedulingIgnoredDuringExecution("affinity"),
					},
				},
				"podAntiAffinity": extv1.JSONSchemaProps{
					Description: "Describes pod anti-affinity scheduling rules (e.g. avoid putting this pod in the same node, zone, etc. as some other pod(s)).",
					Type:        "object",
					Properties: map[string]extv1.JSONSchemaProps{
						"preferredDuringSchedulingIgnoredDuringExecution": getPodPreferredDuringSchedulingIgnoredDuringExecution("anti-affinity"),
						"requiredDuringSchedulingIgnoredDuringExecution":  getPodRequiredDuringSchedulingIgnoredDuringExecution("anti-affinity"),
					},
				},
			},
		},
		"tolerations": extv1.JSONSchemaProps{
			Type: "array",
			Items: &extv1.JSONSchemaPropsOrArray{
				Schema: &extv1.JSONSchemaProps{
					Description: "The pod this Toleration is attached to tolerates any taint that matches the triple <key,value,effect> using the matching operator <operator>.",
					Type:        "object",
					Properties: map[string]extv1.JSONSchemaProps{
						"effect": extv1.JSONSchemaProps{
							Description: "Effect indicates the taint effect to match. Empty means match all taint effects. When specified, allowed values are NoSchedule, PreferNoSchedule and NoExecute.",
							Type:        "string",
						},
						"key": extv1.JSONSchemaProps{
							Description: "Key is the taint key that the toleration applies to. Empty means match all taint keys. If the key is empty, operator must be Exists; this combination means to match all values and all keys.",
							Type:        "string",
						},
						"operator": extv1.JSONSchemaProps{
							Description: "Operator represents a key's relationship to the value. Valid operators are Exists and Equal. Defaults to Equal. Exists is equivalent to wildcard for value, so that a pod can tolerate all taints of a particular category.",
							Type:        "string",
						},
						"tolerationSeconds": extv1.JSONSchemaProps{
							Description: "TolerationSeconds represents the period of time the toleration (which must be of effect NoExecute, otherwise this field is ignored) tolerates the taint. By default, it is not set, which means tolerate the taint forever (do not evict). Zero and negative values will be treated as 0 (evict immediately) by the system.",
							Type:        "integer",
							Format:      "int64",
						},
						"value": extv1.JSONSchemaProps{
							Description: "Value is the taint value the toleration matches to. If the operator is Exists, the value should be empty, otherwise just a regular string.",
							Type:        "string",
						},
					},
				},
			},
		},
	}

	validationSchema := &extv1.CustomResourceValidation{
		OpenAPIV3Schema: &extv1.JSONSchemaProps{
			Description: "NetworkAddonsConfig is the Schema for the networkaddonsconfigs API",
			Type:        "object",
			Properties: map[string]extv1.JSONSchemaProps{
				"apiVersion": extv1.JSONSchemaProps{
					Description: "APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources",
					Type:        "string",
				},
				"kind": extv1.JSONSchemaProps{
					Description: "Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds",
					Type:        "string",
				},
				"metadata": extv1.JSONSchemaProps{
					Type: "object",
				},
				"spec": extv1.JSONSchemaProps{
					Description: "NetworkAddonsConfigSpec defines the desired state of NetworkAddonsConfig",
					Type:        "object",
					Properties: map[string]extv1.JSONSchemaProps{
						"imagePullPolicy": extv1.JSONSchemaProps{
							Description: "PullPolicy describes a policy for if/when to pull a container image",
							Type:        "string",
						},
						"kubeMacPool": extv1.JSONSchemaProps{
							Description: "KubeMacPool plugin manages MAC allocation to Pods and VMs in Kubernetes",
							Type:        "object",
							Properties: map[string]extv1.JSONSchemaProps{
								"rangeEnd": extv1.JSONSchemaProps{
									Description: "RangeEnd defines the first mac in range",
									Type:        "string",
								},
								"rangeStart": extv1.JSONSchemaProps{
									Description: "RangeStart defines the first mac in range",
									Type:        "string",
								},
							},
						},
						"linuxBridge": extv1.JSONSchemaProps{
							Description: "LinuxBridge plugin allows users to create a bridge and add the host and the container to it",
							Type:        "object",
						},
						"macvtap": extv1.JSONSchemaProps{
							Description: "MacvtapCni plugin allows users to define Kubernetes networks on top of existing host interfaces",
							Type:        "object",
							Properties: map[string]extv1.JSONSchemaProps{
								"devicePluginConfig": extv1.JSONSchemaProps{
									Description: "DevicePluginConfig allows the user to override the name of the `ConfigMap` where the device plugin configuration is held",
									Type:        "string",
								},
							},
						},
						"multus": extv1.JSONSchemaProps{
							Description: "Multus plugin enables attaching multiple network interfaces to Pods in Kubernetes",
							Type:        "object",
						},
						"multusDynamicNetworks": extv1.JSONSchemaProps{
							Description: "A multus extension enabling hot-plug and hot-unplug of Pod interfaces",
							Type:        "object",
						},
						"kubeSecondaryDNS": extv1.JSONSchemaProps{
							Description: "KubeSecondaryDNS plugin allows to support FQDN for VMI's secondary networks",
							Type:        "object",
							Properties: map[string]extv1.JSONSchemaProps{
								"domain": extv1.JSONSchemaProps{
									Description: "Domain defines the FQDN domain",
									Type:        "string",
								},
								"nameServerIP": extv1.JSONSchemaProps{
									Description: "NameServerIp defines the name server IP",
									Type:        "string",
								},
							},
						},
						"ovs": extv1.JSONSchemaProps{
							Description: "Ovs plugin allows users to define Kubernetes networks on top of Open vSwitch bridges available on nodes",
							Type:        "object",
						},
						"kubevirtIpamController": extv1.JSONSchemaProps{
							Description: "KubevirtIpamController plugin allows to support IPAM for secondary networks",
							Type:        "object",
						},
						"selfSignConfiguration": extv1.JSONSchemaProps{
							Description: "SelfSignConfiguration defines self sign configuration",
							Type:        "object",
							Properties: map[string]extv1.JSONSchemaProps{
								"caRotateInterval": extv1.JSONSchemaProps{
									Description: "CARotateInterval defines duration for CA expiration",
									Type:        "string",
								},
								"certRotateInterval": extv1.JSONSchemaProps{
									Description: "CertRotateInterval defines duration for of service certificate expiration",
									Type:        "string",
								},
								"caOverlapInterval": extv1.JSONSchemaProps{
									Description: "CAOverlapInterval defines the duration where expired CA certificate can overlap with new one, in order to allow fluent CA rotation transitioning",
									Type:        "string",
								},
								"certOverlapInterval": extv1.JSONSchemaProps{
									Description: "CertOverlapInterval defines the duration where expired service certificate can overlap with new one, in order to allow fluent service rotation transitioning",
									Type:        "string",
								},
							},
						},
						"placementConfiguration": extv1.JSONSchemaProps{
							Description: "PlacementConfiguration defines node placement configuration",
							Type:        "object",
							Properties: map[string]extv1.JSONSchemaProps{
								"infra": extv1.JSONSchemaProps{
									Description: "Infra defines placement configuration for control-plane nodes",
									Type:        "object",
									Properties:  placementProps,
								},
								"workloads": extv1.JSONSchemaProps{
									Type:       "object",
									Properties: placementProps,
								},
							},
						},
						"tlsSecurityProfile": extv1.JSONSchemaProps{
							Description: "TLSSecurityProfile defines the schema for a TLS security profile. This object is used by operators to apply TLS security settings to operands.",
							Type:        "object",
							Nullable:    true,
							Properties: map[string]extv1.JSONSchemaProps{
								"custom": extv1.JSONSchemaProps{
									Description: "custom is a user-defined TLS security profile. Be extremely careful using a custom profile as invalid configurations can be catastrophic. An example custom profile looks like this: ciphers: ECDHE-ECDSA-CHACHA20-POLY1305,ECDHE-RSA-CHACHA20-POLY1305,ECDHE-RSA-AES128-GCM-SHA256,ECDHE-ECDSA-AES128-GCM-SHA256 minTLSVersion: TLSv1.1",
									Nullable:    true,
									Properties:  customSecurityProfileProps,
									Type:        "object",
								},
								"intermediate": extv1.JSONSchemaProps{
									Description: "intermediate is a TLS security profile based on: https://wiki.mozilla.org/Security/Server_Side_TLS#Intermediate_compatibility_.28recommended.29 and looks like this (yaml):\n   ciphers: TLS_AES_128_GCM_SHA256,TLS_AES_256_GCM_SHA384,TLS_CHACHA20_POLY1305_SHA256,ECDHE-ECDSA-AES128-GCM-SHA256     - ECDHE-RSA-AE,SHA256,ECDHE-ECDSA-AES256-GCM-SHA384,ECDHE-RSA-AE,SHA384,ECDHE-ECDSA-CHACHA20-POLY1305,ECDHE,POLY1305,DHE-RSA-AES128-GCM-SHA256,DHE-RSA-AES256-GCM-SHA384 minTLSVersion: TLSv1.2",
									Nullable:    true,
									Type:        "object",
								},
								"modern": extv1.JSONSchemaProps{
									Description: "modern is a TLS security profile based on: https://wiki.mozilla.org/Security/Server_Side_TLS#Modern_compatibility and looks like this (yaml): ciphers: TLS_AES_128_GCM_SHA256,TLS_AES_256_GCM_SHA384,TLS_CHACHA20_POLY1305_SHA256 minTLSVersion: TLSv1.3 NOTE: Currently unsupported.",
									Nullable:    true,
									Type:        "object",
								},
								"old": extv1.JSONSchemaProps{
									Description: "old is a TLS security profile based on: https://wiki.mozilla.org/Security/Server_Side_TLS#Old_backward_compatibility and looks like this (yaml): ciphers: TLS_AES_128_GCM_SHA256,TLS_AES_256_GCM_SHA384,TLS_CHACHA20_POLY1305_SHA256,ECDHE-ECDSA-AES128-GCM-SHA256,ECDHE-RSA-AES128-GCM-SHA256,ECDHE-ECDSA-AES256-GCM-SHA384,ECDHE-RSA-AES256-GCM-SHA384,ECDHE-ECDSA-CHACHA20-POLY1305,ECDHE-RSA-CHACHA20-POLY1305,DHE-RSA-AES128-GCM-SHA256,DHE-RSA-AES256-GCM-SHA384,DHE-RSA-CHACHA20-POLY1305,ECDHE-ECDSA-AES128-SHA256,ECDHE-RSA-AES128-SHA256,ECDHE-ECDSA-AES128-SHA,ECDHE-RSA-AES128-SHA,ECDHE-ECDSA-AES256-SHA384,ECDHE-RSA-AES256-SHA384,ECDHE-ECDSA-AES256-SHA,ECDHE-RSA-AES256-SHA,DHE-RSA-AES128-SHA256,DHE-RSA-AES256-SHA256,AES128-GCM-SHA256,AES256-GCM-SHA384,AES128-SHA256,AES256-SHA256,AES128-SHA,AES256-SHA,DES-CBC3-SHA minTLSVersion: TLSv1.0",
									Nullable:    true,
									Type:        "object",
								},
								"type": extv1.JSONSchemaProps{
									Description: "type is one of Old, Intermediate, Modern or Custom. Custom provides the ability to specify individual TLS security profile parameters. Old, Intermediate and Modern are TLS security profiles based on:\n https://wiki.mozilla.org/Security/Server_Side_TLS#Recommended_configurations The profiles are intent based, so they may change over time as new ciphers are developed and existing ciphers are found to be insecure.  Depending on precisely which ciphers are available to a process, the list may be reduced.\n Note that the Modern profile is currently not supported because it is not yet well adopted by common software libraries.",
									Enum: []extv1.JSON{
										{Raw: []byte(fmt.Sprintf("\"%s\"", ocpv1.TLSProfileOldType))},
										{Raw: []byte(fmt.Sprintf("\"%s\"", ocpv1.TLSProfileIntermediateType))},
										{Raw: []byte(fmt.Sprintf("\"%s\"", ocpv1.TLSProfileModernType))},
										{Raw: []byte(fmt.Sprintf("\"%s\"", ocpv1.TLSProfileCustomType))},
									},
									Type: "string",
								},
							},
						},
					},
				},
				"status": extv1.JSONSchemaProps{
					Description: "NetworkAddonsConfigStatus defines the observed state of NetworkAddonsConfig",
					Type:        "object",
					Properties: map[string]extv1.JSONSchemaProps{
						"conditions": extv1.JSONSchemaProps{
							Type: "array",
							Items: &extv1.JSONSchemaPropsOrArray{
								Schema: &extv1.JSONSchemaProps{
									Description: "Condition represents the state of the operator's reconciliation functionality.",
									Type:        "object",
									Properties: map[string]extv1.JSONSchemaProps{
										"lastHeartbeatTime": extv1.JSONSchemaProps{
											Format:   "date-time",
											Type:     "string",
											Nullable: true,
										},
										"lastTransitionTime": extv1.JSONSchemaProps{
											Format:   "date-time",
											Type:     "string",
											Nullable: true,
										},
										"message": extv1.JSONSchemaProps{
											Type: "string",
										},
										"reason": extv1.JSONSchemaProps{
											Type: "string",
										},
										"status": extv1.JSONSchemaProps{
											Type: "string",
										},
										"type": extv1.JSONSchemaProps{
											Description: "ConditionType is the state of the operator's reconciliation functionality.",
											Type:        "string",
										},
									},
									Required: []string{
										"status",
										"type",
									},
								},
							},
						},
						"containers": extv1.JSONSchemaProps{
							Type: "array",
							Items: &extv1.JSONSchemaPropsOrArray{
								Schema: &extv1.JSONSchemaProps{
									Properties: map[string]extv1.JSONSchemaProps{
										"image": extv1.JSONSchemaProps{
											Type: "string",
										},
										"name": extv1.JSONSchemaProps{
											Type: "string",
										},
										"parentKind": extv1.JSONSchemaProps{
											Type: "string",
										},
										"parentName": extv1.JSONSchemaProps{
											Type: "string",
										},
									},
									Required: []string{
										"image",
										"name",
										"parentKind",
										"parentName",
									},
									Type: "object",
								},
							},
						},
						"observedVersion": extv1.JSONSchemaProps{
							Type: "string",
						},
						"operatorVersion": extv1.JSONSchemaProps{
							Type: "string",
						},
						"targetVersion": extv1.JSONSchemaProps{
							Type: "string",
						},
					},
				},
			},
		},
	}

	crd := &extv1.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apiextensions.k8s.io/v1",
			Kind:       "CustomResourceDefinition",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "networkaddonsconfigs.networkaddonsoperator.network.kubevirt.io",
		},
		Spec: extv1.CustomResourceDefinitionSpec{
			Group: "networkaddonsoperator.network.kubevirt.io",
			Scope: "Cluster",

			Names: extv1.CustomResourceDefinitionNames{
				Plural:   "networkaddonsconfigs",
				Singular: "networkaddonsconfig",
				Kind:     "NetworkAddonsConfig",
				ListKind: "NetworkAddonsConfigList",
			},

			Versions: []extv1.CustomResourceDefinitionVersion{
				{
					Name:         "v1",
					Served:       true,
					Storage:      true,
					Schema:       validationSchema,
					Subresources: subResouceSchema,
				},
				{
					Name:         "v1alpha1",
					Served:       true,
					Storage:      false,
					Schema:       validationSchema,
					Subresources: subResouceSchema,
				},
			},
		},
	}
	return crd
}

func GetCRV1() *cnaov1.NetworkAddonsConfig {
	return &cnaov1.NetworkAddonsConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networkaddonsoperator.network.kubevirt.io/v1",
			Kind:       "NetworkAddonsConfig",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster",
		},
		Spec: cnao.NetworkAddonsConfigSpec{
			Multus:                 &cnao.Multus{},
			MultusDynamicNetworks:  &cnao.MultusDynamicNetworks{},
			LinuxBridge:            &cnao.LinuxBridge{},
			KubeMacPool:            &cnao.KubeMacPool{},
			Ovs:                    &cnao.Ovs{},
			MacvtapCni:             &cnao.MacvtapCni{},
			KubeSecondaryDNS:       &cnao.KubeSecondaryDNS{},
			KubevirtIpamController: &cnao.KubevirtIpamController{},
			ImagePullPolicy:        corev1.PullIfNotPresent,
		},
	}
}

func getMonitoringNamespace() string {
	namespace := os.Getenv("MONITORING_NAMESPACE")
	if namespace == "" {
		return "openshift-monitoring"
	}
	return namespace
}

func int32Ptr(i int32) *int32 {
	return &i
}
