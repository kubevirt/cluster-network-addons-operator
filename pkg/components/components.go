package components

import (
	"fmt"
	"regexp"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
	cnaov1 "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/v1"
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
	MultusImageDefault            = "quay.io/kubevirt/cluster-network-addon-multus@sha256:32867c73cda4d605651b898dc85fea67d93191c47f27e1ad9e9f2b9041c518de"
	LinuxBridgeCniImageDefault    = "quay.io/kubevirt/cni-default-plugins@sha256:a90902cf3e5154424148bf3ba3c1bf90316cc77a54042cf6584fe8aedbe6daec"
	LinuxBridgeMarkerImageDefault = "quay.io/kubevirt/bridge-marker@sha256:3cc868fd1cc18e111775f6f00066486e040c481c44e7eecc1b9b385c0d1d1d77"
	KubeMacPoolImageDefault       = "quay.io/kubevirt/kubemacpool@sha256:ef759a6e8960d895e777621381c3e94d677f1401435bc00c7663dc1b828272cb"
	NMStateHandlerImageDefault    = "quay.io/nmstate/kubernetes-nmstate-handler@sha256:0e30f69b9568b252d9d86c46292821c76bf8f471fd81e099e55ff613727267be"
	OvsCniImageDefault            = "quay.io/kubevirt/ovs-cni-plugin@sha256:6129aa9b9d4d9c077cae47b20500ae7d128b56d1f4296b957eecfdd99bee134b"
	OvsMarkerImageDefault         = "quay.io/kubevirt/ovs-cni-marker@sha256:d0661bc26a56e0b7066ec056e5fc95f21dfa9d23aba6760d213951c8f50b3920"
	MacvtapCniImageDefault        = "quay.io/kubevirt/macvtap-cni@sha256:e34cc796dee2e300f866d6f5b563361253ce89226eaf9eb0c3bc792f5481b8df"
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
			Labels: map[string]string{
				names.PROMETHEUS_LABEL_KEY: "",
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
						names.PROMETHEUS_LABEL_KEY: "",
					},
					Annotations: map[string]string{
						"description": "cluster-network-addons-operator manages the lifecycle of different Kubernetes network components on top of Kubernetes cluster",
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: Name,
					SecurityContext: &corev1.PodSecurityContext{
						RunAsNonRoot: &runAsNonRoot,
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
									Description: "Infra defines placement configuration for master nodes",
									Type:        "object",
									Properties:  placementProps,
								},
								"workloads": extv1.JSONSchemaProps{
									Type:       "object",
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
