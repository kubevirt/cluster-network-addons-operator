package components

import (
	"fmt"
	"regexp"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
	cnaov1 "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/v1"
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
	MultusImageDefault            = "nfvpe/multus@sha256:ac1266b87ba44c09dc2a336f0d5dad968fccd389ce1944a85e87b32cd21f7224"
	LinuxBridgeCniImageDefault    = "quay.io/kubevirt/cni-default-plugins@sha256:3dd438117076016d6d2acd508b93f106ca80a28c0af6e2e914d812f9a1d55142"
	LinuxBridgeMarkerImageDefault = "quay.io/kubevirt/bridge-marker@sha256:99d0fbd707deaf2136968aaa34862b54c73c9e5e963dc44140e726cf9ad41b58"
	KubeMacPoolImageDefault       = "quay.io/kubevirt/kubemacpool@sha256:79c4534d418c4a350a663e38499c22d54dc68c400f517aead4479f6d862b408e"
	NMStateHandlerImageDefault    = "quay.io/nmstate/kubernetes-nmstate-handler@sha256:b4bc41ce7f9e5fa7eef6a25c27ba13770c53fc6710ea8e4c4b5f6cf68e97821c"
	OvsCniImageDefault            = "quay.io/kubevirt/ovs-cni-plugin@sha256:283fcdff34b6dc726f88a7c4a6a0cbc35f8779960d54d69e54845c37f5da3121"
	OvsMarkerImageDefault         = "quay.io/kubevirt/ovs-cni-marker@sha256:e377a4bd0119de9363fc4a3772bf320afb95f36e10c034cf19b86de47e7fcca3"
	MacvtapCniImageDefault        = "quay.io/kubevirt/macvtap-cni@sha256:0fbb0f3cde7970c0786aec8213bbf28a9d6328e7644d965262783a8248a9ded9"
)

type AddonsImages struct {
	Multus            string
	LinuxBridgeCni    string
	LinuxBridgeMarker string
	KubeMacPool       string
	NMStateHandler    string
	OvsCni            string
	OvsMarker         string
	MacvtapCni        string
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
	if ai.LinuxBridgeCni == "" {
		ai.LinuxBridgeCni = LinuxBridgeCniImageDefault
	}
	if ai.LinuxBridgeMarker == "" {
		ai.LinuxBridgeMarker = LinuxBridgeMarkerImageDefault
	}
	if ai.KubeMacPool == "" {
		ai.KubeMacPool = KubeMacPoolImageDefault
	}
	if ai.NMStateHandler == "" {
		ai.NMStateHandler = NMStateHandlerImageDefault
	}
	if ai.OvsCni == "" {
		ai.OvsCni = OvsCniImageDefault
	}
	if ai.OvsMarker == "" {
		ai.OvsMarker = OvsMarkerImageDefault
	}
	if ai.MacvtapCni == "" {
		ai.MacvtapCni = MacvtapCniImageDefault
	}
	return ai
}

func (ai AddonsImages) ToRelatedImages() RelatedImages {
	return NewRelatedImages(
		ai.Multus,
		ai.LinuxBridgeCni,
		ai.LinuxBridgeMarker,
		ai.KubeMacPool,
		ai.NMStateHandler,
		ai.OvsCni,
		ai.OvsMarker,
		ai.MacvtapCni,
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

	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      Name,
			Namespace: namespace,
			Annotations: map[string]string{
				cnaov1.SchemeGroupVersion.Group + "/version": k8s.StringToLabel(operatorVersion),
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
						"name": Name,
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: Name,
					SecurityContext: &corev1.PodSecurityContext{
						RunAsNonRoot: &runAsNonRoot,
					},
					Containers: []corev1.Container{
						{
							Name:            Name,
							Image:           image,
							ImagePullPolicy: corev1.PullPolicy(imagePullPolicy),
							Env: []corev1.EnvVar{
								{
									Name:  "MULTUS_IMAGE",
									Value: addonsImages.Multus,
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
									Name:  "NMSTATE_HANDLER_IMAGE",
									Value: addonsImages.NMStateHandler,
								},
								{
									Name:  "OVS_CNI_IMAGE",
									Value: addonsImages.OvsCni,
								},
								{
									Name:  "OVS_MARKER_IMAGE",
									Value: addonsImages.OvsMarker,
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
					"",
				},
				Resources: []string{
					"pods",
					"configmaps",
				},
				Verbs: []string{
					"get",
					"list",
					"watch",
					"create",
					"patch",
					"update",
					"delete",
				},
			},
			{
				APIGroups: []string{
					"apps",
				},
				Resources: []string{
					"deployments",
					"replicasets",
				},
				Verbs: []string{
					"get",
					"list",
					"watch",
					"create",
					"patch",
					"update",
					"delete",
				},
			},
		},
	}
	return role
}

func GetClusterRole() *rbacv1.ClusterRole {
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
					"security.openshift.io",
				},
				Resources: []string{
					"securitycontextconstraints",
				},
				ResourceNames: []string{
					"privileged",
				},
				Verbs: []string{
					"get",
					"list",
					"watch",
				},
			},
			{
				APIGroups: []string{
					"operator.openshift.io",
				},
				Resources: []string{
					"networks",
				},
				Verbs: []string{
					"get",
					"list",
					"watch",
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
					"get",
					"list",
					"watch",
				},
			},
			{
				APIGroups: []string{
					"*",
				},
				Resources: []string{
					"*",
				},
				Verbs: []string{
					"*",
				},
			},
		},
	}
	return role
}

func GetCrd() *extv1.CustomResourceDefinition {
	subResouceSchema := &extv1.CustomResourceSubresources{Status: &extv1.CustomResourceSubresourceStatus{}}
	placementProps := map[string]extv1.JSONSchemaProps{
		"NodeSelector" : extv1.JSONSchemaProps{
			Type:        "object",
		},
		"Affinity" : extv1.JSONSchemaProps{
			Description: "Affinity is a group of affinity scheduling rules.",
			Type:        "object",
		},
		"Tolerations" : extv1.JSONSchemaProps{
			Description: "The pod this Toleration is attached to tolerates any taint that matches the triple <key,value,effect> using the matching operator <operator>.",
			Type:        "object",
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
						},
						"multus": extv1.JSONSchemaProps{
							Description: "Multus plugin enables attaching multiple network interfaces to Pods in Kubernetes",
							Type:        "object",
						},
						"nmstate": extv1.JSONSchemaProps{
							Description: "NMState is a declarative node network configuration driven through Kubernetes API",
							Type:        "object",
						},
						"ovs": extv1.JSONSchemaProps{
							Description: "Ovs plugin allows users to define Kubernetes networks on top of Open vSwitch bridges available on nodes",
							Type:        "object",
						},
						"selfSignConfiguration": extv1.JSONSchemaProps{
							Description: "SelfSignConfiguration defines self sign configuration",
							Type: "object",
							Properties: map[string]extv1.JSONSchemaProps{
								"caRotateInterval": extv1.JSONSchemaProps{
									Description: "CARotateInterval defines duration for CA and certificate",
									Type:        "string",
								},
								"certRotateInterval": extv1.JSONSchemaProps{
									Description: "CertRotateInterval defines duration for of service certificate",
									Type:        "string",
								},
								"caOverlapInterval": extv1.JSONSchemaProps{
									Description: "CAOverlapInterval defines the duration of CA Certificates at CABundle if not set it will default to CARotateInterval",
									Type:        "string",
								},
							},
						},
						"PlacementConfiguration": extv1.JSONSchemaProps{
							Description: "PlacementConfiguration defines node placement configuration",
							Type: "object",
							Properties: map[string]extv1.JSONSchemaProps{
								"Infra": extv1.JSONSchemaProps{
									Description: "Infra defines placement configuration for master nodes",
									Type:        "object",
									Properties: placementProps,
								},
								"Workloads": extv1.JSONSchemaProps{
									Description: "Workloads defines placement configuration for worker nodes",
									Type:        "object",
									Properties: placementProps,
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
							//Description: "Condition represents the state of the operator's reconciliation functionality.",
							Type: "array",
							Items: &extv1.JSONSchemaPropsOrArray{
								Schema: &extv1.JSONSchemaProps{
									Type: "object",
									Properties: map[string]extv1.JSONSchemaProps{
										"lastHeartbeatTime": extv1.JSONSchemaProps{
											Format: "date-time",
											Type:   "string",
											Nullable: true,
										},
										"lastTransitionTime": extv1.JSONSchemaProps{
											Format: "date-time",
											Type:   "string",
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
			Multus:          &cnao.Multus{},
			LinuxBridge:     &cnao.LinuxBridge{},
			KubeMacPool:     &cnao.KubeMacPool{},
			NMState:         &cnao.NMState{},
			Ovs:             &cnao.Ovs{},
			MacvtapCni:      &cnao.MacvtapCni{},
			ImagePullPolicy: corev1.PullIfNotPresent,
		},
	}
}

func int32Ptr(i int32) *int32 {
	return &i
}
