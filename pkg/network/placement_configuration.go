package network

import (
	corev1 "k8s.io/api/core/v1"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
)

func GetDefaultPlacementConfiguration() cnao.PlacementConfiguration {
	return cnao.PlacementConfiguration{
		Infra: &cnao.Placement{
			Tolerations: []corev1.Toleration{
				{
					Key:      "node-role.kubernetes.io/control-plane",
					Operator: corev1.TolerationOpExists,
					Effect:   corev1.TaintEffectNoSchedule,
				},
				{
					Key:      "node-role.kubernetes.io/master",
					Operator: corev1.TolerationOpExists,
					Effect:   corev1.TaintEffectNoSchedule,
				},
			},
			Affinity: corev1.Affinity{
				NodeAffinity: &corev1.NodeAffinity{
					PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{
						{
							Weight: 10,
							Preference: corev1.NodeSelectorTerm{
								MatchExpressions: []corev1.NodeSelectorRequirement{
									{
										Key:      "node-role.kubernetes.io/control-plane",
										Operator: corev1.NodeSelectorOpExists,
									},
								},
							},
						},
						{
							Weight: 1,
							Preference: corev1.NodeSelectorTerm{
								MatchExpressions: []corev1.NodeSelectorRequirement{
									{
										Key:      "node-role.kubernetes.io/master",
										Operator: corev1.NodeSelectorOpExists,
									},
								},
							},
						},
					},
				},
			},
		},
		Workloads: &cnao.Placement{
			NodeSelector: map[string]string{
				"beta.kubernetes.io/arch": "amd64",
			},
			Tolerations: []corev1.Toleration{
				{
					Operator: corev1.TolerationOpExists,
					Effect:   corev1.TaintEffectNoSchedule,
				},
			},
		},
	}
}

func fillDefaultsPlacementConfiguration(conf, previous *cnao.NetworkAddonsConfigSpec) []error {
	defaultPlacementConfiguration := GetDefaultPlacementConfiguration()
	if conf.PlacementConfiguration == nil {
		conf.PlacementConfiguration = &defaultPlacementConfiguration
		return []error{}
	}
	if conf.PlacementConfiguration.Infra == nil {
		conf.PlacementConfiguration.Infra = defaultPlacementConfiguration.Infra
	}
	if conf.PlacementConfiguration.Workloads == nil {
		conf.PlacementConfiguration.Workloads = defaultPlacementConfiguration.Workloads
	}
	return []error{}
}
