package network

import (
	corev1 "k8s.io/api/core/v1"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
)

func GetDefaultPlacementConfiguration() cnao.PlacementConfiguration {
	return cnao.PlacementConfiguration{
		Infra: &cnao.Placement{
			NodeSelector: map[string]string{
				"beta.kubernetes.io/arch":               "amd64",
				"node-role.kubernetes.io/control-plane": "",
			},
			Tolerations: []corev1.Toleration{
				corev1.Toleration{
					Key:      "node-role.kubernetes.io/control-plane",
					Operator: corev1.TolerationOpExists,
					Effect:   corev1.TaintEffectNoSchedule,
				},
			},
		},
		Workloads: &cnao.Placement{
			NodeSelector: map[string]string{
				"beta.kubernetes.io/arch": "amd64",
			},
			Tolerations: []corev1.Toleration{
				corev1.Toleration{
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
