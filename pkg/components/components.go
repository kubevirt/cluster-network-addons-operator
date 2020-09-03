package components

import (
	"fmt"
	"regexp"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	extv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
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
	MultusImageDefault            = "nfvpe/multus@sha256:167722b954355361bd69829466f27172b871dbdbf86b85a95816362885dc0aba"
	LinuxBridgeCniImageDefault    = "quay.io/kubevirt/cni-default-plugins@sha256:3dd438117076016d6d2acd508b93f106ca80a28c0af6e2e914d812f9a1d55142"
	LinuxBridgeMarkerImageDefault = "quay.io/kubevirt/bridge-marker@sha256:e55f73526468fee46a35ae41aa860f492d208b8a7a132832c5b9a76d4a51566a"
	KubeMacPoolImageDefault       = "quay.io/kubevirt/kubemacpool@sha256:ad8ca6d379d495804969ba4d03da9a6936ff8f413f6f6c7bd20e0138dc0303c4"
	NMStateHandlerImageDefault    = "quay.io/nmstate/kubernetes-nmstate-handler@sha256:76cc13fb4a60943dca6038619599b6a49fe451852aba23ad3046658429a9af30"
	OvsCniImageDefault            = "quay.io/kubevirt/ovs-cni-plugin@sha256:4101c52617efb54a45181548c257a08e3689f634b79b9dfcff42bffd8b25af53"
	OvsMarkerImageDefault         = "quay.io/kubevirt/ovs-cni-marker@sha256:0f08d6b1550a90c9f10221f2bb07709d1090e7c675ee1a711981bd429074d620"
	MacvtapCniImageDefault        = "quay.io/kubevirt/macvtap-cni@sha256:2606ac545cc69af5766d929bd82f86df5098a5e1c8ac26cfd2cdf99c46da8067"
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

func GetCrd() *extv1beta1.CustomResourceDefinition {
	crd := &extv1beta1.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apiextensions.k8s.io/v1beta1",
			Kind:       "CustomResourceDefinition",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "networkaddonsconfigs.networkaddonsoperator.network.kubevirt.io",
		},
		Spec: extv1beta1.CustomResourceDefinitionSpec{
			Group: "networkaddonsoperator.network.kubevirt.io",
			Scope: "Cluster",

			Subresources: &extv1beta1.CustomResourceSubresources{
				Status: &extv1beta1.CustomResourceSubresourceStatus{},
			},

			Names: extv1beta1.CustomResourceDefinitionNames{
				Plural:   "networkaddonsconfigs",
				Singular: "networkaddonsconfig",
				Kind:     "NetworkAddonsConfig",
				ListKind: "NetworkAddonsConfigList",
			},

			Versions: []extv1beta1.CustomResourceDefinitionVersion{
				{
					Name:    "v1",
					Served:  true,
					Storage: true,
				},
				{
					Name:    "v1alpha1",
					Served:  true,
					Storage: false,
				},
			},

			Validation: &extv1beta1.CustomResourceValidation{
				OpenAPIV3Schema: &extv1beta1.JSONSchemaProps{
					Type: "object",
					Properties: map[string]extv1beta1.JSONSchemaProps{
						"apiVersion": extv1beta1.JSONSchemaProps{
							Type: "string",
						},
						"kind": extv1beta1.JSONSchemaProps{
							Type: "string",
						},
						"metadata": extv1beta1.JSONSchemaProps{
							Type: "object",
						},
						"spec": extv1beta1.JSONSchemaProps{
							Type: "object",
						},
						"status": extv1beta1.JSONSchemaProps{
							Type: "object",
						},
					},
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
