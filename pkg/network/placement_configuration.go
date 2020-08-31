package network

import (
	"reflect"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
)

var (
	defaultPlacementConfiguration = cnao.PlacementConfiguration{
		Infra: &cnao.Placement{
			NodeSelector: map[string]string{
				"beta.kubernetes.io/arch":        "amd64",
				"node-role.kubernetes.io/master": "",
			},
			Tolerations: []corev1.Toleration{
				corev1.Toleration{
					Key:      "node-role.kubernetes.io/master",
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
)

func fillDefaultsPlacementConfiguration(conf, previous *cnao.NetworkAddonsConfigSpec) []error {
	if conf.PlacementConfiguration == nil || conf.PlacementConfiguration.Infra == nil || conf.PlacementConfiguration.Workloads == nil {
		if previous != nil && previous.PlacementConfiguration != nil {
			conf.PlacementConfiguration = previous.PlacementConfiguration
			return []error{}
		}

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
	}
	return []error{}
}

func changeSafePlacementConfiguration(prev, next *cnao.NetworkAddonsConfigSpec) []error {
	if prev.PlacementConfiguration != nil && next.PlacementConfiguration != nil && !reflect.DeepEqual(prev.PlacementConfiguration, next.PlacementConfiguration) {
		return []error{errors.Errorf("cannot modify PlacementConfiguration once it is deployed")}
	}
	return []error{}
}
