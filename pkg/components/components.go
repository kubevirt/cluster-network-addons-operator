package components

import (
	"fmt"

	cnav1alpha1 "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/v1alpha1"
	names "github.com/kubevirt/cluster-network-addons-operator/pkg/names"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	extv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetDeployment(repository string, tag string, imagePullPolicy string) *appsv1.Deployment {
	name := names.APPLIED_NAMESPACE
	image := fmt.Sprintf("%s/%s:%s", repository, name, tag)
	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: name,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"name": name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"name": name,
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: name,
					Containers: []corev1.Container{
						{
							Name:            name,
							Image:           image,
							ImagePullPolicy: corev1.PullPolicy(imagePullPolicy),
							Env: []corev1.EnvVar{
								{
									Name:  "MULTUS_IMAGE",
									Value: "docker.io/nfvpe/multus:latest",
								},
								{
									Name:  "LINUX_BRIDGE_IMAGE",
									Value: "quay.io/kubevirt/cni-default-plugins",
								},
								{
									Name:  "SRIOV_DP_IMAGE",
									Value: "quay.io/booxter/sriov-device-plugin:latest",
								},
								{
									Name:  "SRIOV_CNI_IMAGE",
									Value: "docker.io/nfvpe/sriov-cni:latest",
								},
								{
									Name:  "SRIOV_ROOT_DEVICES",
									Value: "",
								},
								{
									Name:  "KUBEMACPOOL_IMAGE",
									Value: "quay.io/schseba/mac-controller:latest",
								},
								{
									Name:  "OPERATOR_IMAGE",
									Value: image,
								},
								{
									Name:  "OPERATOR_NAME",
									Value: name,
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
									Name: "WATCH_NAMESPACE",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
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

func GetRole() *rbacv1.Role {
	name := names.APPLIED_NAMESPACE
	role := &rbacv1.Role{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "Role",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: name,
			Labels: map[string]string{
				"name": name,
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
	name := names.APPLIED_NAMESPACE
	role := &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "ClusterRole",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"name": name,
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
			Group:   "networkaddonsoperator.network.kubevirt.io",
			Version: "v1alpha1",
			Scope:   "Cluster",

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
					Name:    "v1alpha1",
					Served:  true,
					Storage: true,
				},
			},

			Validation: &extv1beta1.CustomResourceValidation{
				OpenAPIV3Schema: &extv1beta1.JSONSchemaProps{
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

func GetCR() *cnav1alpha1.NetworkAddonsConfig {
	return &cnav1alpha1.NetworkAddonsConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networkaddonsoperator.network.kubevirt.io/v1alpha1",
			Kind:       "NetworkAddonsConfig",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster",
		},
		Spec: cnav1alpha1.NetworkAddonsConfigSpec{
			Multus:      &cnav1alpha1.Multus{},
			LinuxBridge: &cnav1alpha1.LinuxBridge{},
			Sriov:       &cnav1alpha1.Sriov{},
			KubeMacPool: &cnav1alpha1.KubeMacPool{
				StartPoolRange: "02:00:00:00:00:00",
				EndPoolRange:   "FD:FF:FF:FF:FF:FF",
			},
			ImagePullPolicy: "Always",
		},
	}
}

func int32Ptr(i int32) *int32 {
	return &i
}
