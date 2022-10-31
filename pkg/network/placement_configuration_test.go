package network

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
)

var _ = Describe("Testing PlacementConfiguration", func() {
	type fillDefaultsCase struct {
		previousConfig *cnao.NetworkAddonsConfigSpec
		currentConfig  *cnao.NetworkAddonsConfigSpec
		expectedConfig cnao.PlacementConfiguration
	}
	defaultPlacementConfiguration := GetDefaultPlacementConfiguration()
	DescribeTable("Fill defaults function",
		func(c fillDefaultsCase) {
			errorList := fillDefaultsPlacementConfiguration(c.currentConfig, c.previousConfig)
			Expect(*c.currentConfig.PlacementConfiguration).To(Equal(c.expectedConfig))
			Expect(errorList).To(BeEmpty())
		},
		Entry("When PlacementConfiguration is nil should return default values", fillDefaultsCase{
			previousConfig: &cnao.NetworkAddonsConfigSpec{},
			currentConfig:  &cnao.NetworkAddonsConfigSpec{},
			expectedConfig: defaultPlacementConfiguration,
		}),
		Entry("When Workloads PlacementConfiguration is nil should return default workloads values", fillDefaultsCase{
			previousConfig: &cnao.NetworkAddonsConfigSpec{},
			currentConfig: &cnao.NetworkAddonsConfigSpec{
				PlacementConfiguration: &cnao.PlacementConfiguration{
					Infra: &cnao.Placement{},
				},
			},
			expectedConfig: cnao.PlacementConfiguration{
				Workloads: defaultPlacementConfiguration.Workloads,
				Infra:     &cnao.Placement{},
			},
		}),
		Entry("When Infra PlacementConfiguration is nil should return default infra values", fillDefaultsCase{
			previousConfig: &cnao.NetworkAddonsConfigSpec{},
			currentConfig: &cnao.NetworkAddonsConfigSpec{
				PlacementConfiguration: &cnao.PlacementConfiguration{
					Workloads: &cnao.Placement{},
				},
			},
			expectedConfig: cnao.PlacementConfiguration{
				Workloads: &cnao.Placement{},
				Infra:     defaultPlacementConfiguration.Infra,
			},
		}),
		Entry("When new Infra & Workloads PlacementConfiguration is not nil should keep new values", fillDefaultsCase{
			previousConfig: &cnao.NetworkAddonsConfigSpec{},
			currentConfig: &cnao.NetworkAddonsConfigSpec{
				PlacementConfiguration: &cnao.PlacementConfiguration{
					Workloads: &cnao.Placement{
						NodeSelector: map[string]string{
							"kubernetes.io/arch": "amd64",
						},
					},
					Infra: &cnao.Placement{
						NodeSelector: map[string]string{
							"node-role.kubernetes.io/control-plane": "",
						},
					},
				},
			},
			expectedConfig: cnao.PlacementConfiguration{
				Workloads: &cnao.Placement{
					NodeSelector: map[string]string{
						"kubernetes.io/arch": "amd64",
					},
				},
				Infra: &cnao.Placement{
					NodeSelector: map[string]string{
						"node-role.kubernetes.io/control-plane": "",
					},
				},
			},
		}),
	)
})
