package network

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
)

var _ = Describe("Testing PlacementConfiguration", func() {
	type fillDefaultsCase struct {
		previousConfig *cnao.NetworkAddonsConfigSpec
		currentConfig  *cnao.NetworkAddonsConfigSpec
		expectedConfig cnao.PlacementConfiguration
	}
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
		Entry("When the user hasn't explicitly changed PlacementConfiguration and a previous configuration exits should use the previous values on the current one", fillDefaultsCase{
			previousConfig: &cnao.NetworkAddonsConfigSpec{
				PlacementConfiguration: &cnao.PlacementConfiguration{
					Workloads: &cnao.Placement{
						NodeSelector: map[string]string{
							"beta.kubernetes.io/arch": "amd64",
						},
					},
				},
			},
			currentConfig: &cnao.NetworkAddonsConfigSpec{},
			expectedConfig: cnao.PlacementConfiguration{
				Workloads: &cnao.Placement{
					NodeSelector: map[string]string{
						"beta.kubernetes.io/arch": "amd64",
					},
				},
			},
		}),
	)
	type changeSafeCase struct {
		previousConfig *cnao.NetworkAddonsConfigSpec
		currentConfig  *cnao.NetworkAddonsConfigSpec
		expectedError  string
	}
	DescribeTable("Change safe function",
		func(c changeSafeCase) {
			errorList := changeSafePlacementConfiguration(c.previousConfig, c.currentConfig)
			if c.expectedError == "" {
				Expect(errorList).To(BeEmpty())
			} else {
				Expect(len(errorList)).To(Equal(1), "validation failed due to an unexpected error: %v", errorList)
				Expect(errorList[0].Error()).To(MatchRegexp(c.expectedError))
			}
		},
		Entry("When they are equal should NOT return an error", changeSafeCase{
			previousConfig: &cnao.NetworkAddonsConfigSpec{
				PlacementConfiguration: &cnao.PlacementConfiguration{
					Infra: &cnao.Placement{
						NodeSelector: map[string]string{
							"label": "value",
						},
					},
				},
			},
			currentConfig: &cnao.NetworkAddonsConfigSpec{
				PlacementConfiguration: &cnao.PlacementConfiguration{
					Infra: &cnao.Placement{
						NodeSelector: map[string]string{
							"label": "value",
						},
					},
				},
			},
			expectedError: "",
		}),
		Entry("When they are not equal should return an error", changeSafeCase{
			previousConfig: &cnao.NetworkAddonsConfigSpec{
				PlacementConfiguration: &cnao.PlacementConfiguration{
					Infra: &cnao.Placement{
						NodeSelector: map[string]string{
							"label": "value",
						},
					},
				},
			},
			currentConfig: &cnao.NetworkAddonsConfigSpec{
				PlacementConfiguration: &cnao.PlacementConfiguration{
					Infra: &cnao.Placement{
						NodeSelector: map[string]string{
							"newlabel": "newvalue",
						},
					},
				},
			},
			expectedError: "cannot modify PlacementConfiguration once it is deployed",
		}),
		Entry("When trying to remove PlacementConfiguration should NOT return an error", changeSafeCase{
			previousConfig: &cnao.NetworkAddonsConfigSpec{
				PlacementConfiguration: &cnao.PlacementConfiguration{
					Infra: &cnao.Placement{
						NodeSelector: map[string]string{
							"label": "value",
						},
					},
				},
			},
			currentConfig: &cnao.NetworkAddonsConfigSpec{},
			expectedError: "",
		}),
	)
})
