package network

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
)

var _ = Describe("Testing SelfSignConfiguration", func() {
	type validateCase struct {
		config        *cnao.NetworkAddonsConfigSpec
		expectedError string
	}
	DescribeTable("validation function",
		func(c validateCase) {
			errorList := validateSelfSignConfiguration(c.config)
			if len(c.expectedError) > 0 {
				Expect(errorList).To(HaveLen(1), "validation failed due to an unexpected error: %v", errorList)
				Expect(errorList[0].Error()).To(MatchRegexp(c.expectedError))
			} else {
				Expect(errorList).To(BeEmpty())
			}
		},
		Entry("When CARotateInterval is the only empty one should return an error", validateCase{
			config: &cnao.NetworkAddonsConfigSpec{
				SelfSignConfiguration: &cnao.SelfSignConfiguration{
					CARotateInterval:    "",
					CAOverlapInterval:   "48h30m50s",
					CertRotateInterval:  "48h30m50s",
					CertOverlapInterval: "48h30m50s",
				},
			},
			expectedError: "failed to validate selfSignConfiguration: caRotateInterval is missing",
		}),
		Entry("When CAOverlapInterval is the only empty one should return an error", validateCase{
			config: &cnao.NetworkAddonsConfigSpec{
				SelfSignConfiguration: &cnao.SelfSignConfiguration{
					CARotateInterval:    "48h30m50s",
					CAOverlapInterval:   "",
					CertRotateInterval:  "48h30m50s",
					CertOverlapInterval: "48h30m50s",
				},
			},
			expectedError: "failed to validate selfSignConfiguration: caOverlapInterval is missing",
		}),
		Entry("When CertRotateInterval is the only empty one should return an error", validateCase{
			config: &cnao.NetworkAddonsConfigSpec{
				SelfSignConfiguration: &cnao.SelfSignConfiguration{
					CARotateInterval:    "48h30m50s",
					CAOverlapInterval:   "48h30m50s",
					CertRotateInterval:  "",
					CertOverlapInterval: "48h30m50s",
				},
			},
			expectedError: "failed to validate selfSignConfiguration: certRotateInterval is missing",
		}),
		Entry("When CertOverlapInterval is the only empty one should return an error", validateCase{
			config: &cnao.NetworkAddonsConfigSpec{
				SelfSignConfiguration: &cnao.SelfSignConfiguration{
					CARotateInterval:    "48h30m50s",
					CAOverlapInterval:   "48h30m50s",
					CertRotateInterval:  "48h30m50s",
					CertOverlapInterval: "",
				},
			},
			expectedError: "failed to validate selfSignConfiguration: certOverlapInterval is missing",
		}),
		Entry("When selfSignConfiguration is valid should not return an error", validateCase{
			config: &cnao.NetworkAddonsConfigSpec{
				SelfSignConfiguration: &cnao.SelfSignConfiguration{
					CARotateInterval:    "24h",
					CAOverlapInterval:   "12h",
					CertRotateInterval:  "10h",
					CertOverlapInterval: "8h",
				},
			},
			expectedError: "",
		}),
		Entry("When CARotateInterval is invalid duration string should return an error", validateCase{
			config: &cnao.NetworkAddonsConfigSpec{
				SelfSignConfiguration: &cnao.SelfSignConfiguration{
					CARotateInterval:    "asdfasfda",
					CAOverlapInterval:   "12h",
					CertRotateInterval:  "10h",
					CertOverlapInterval: "8h",
				},
			},
			expectedError: "failed to validate selfSignConfiguration: error parsing caRotateInterval: time",
		}),
		Entry("When CAOverlapInterval is invalid duration string should return an error", validateCase{
			config: &cnao.NetworkAddonsConfigSpec{
				SelfSignConfiguration: &cnao.SelfSignConfiguration{
					CARotateInterval:    "12h",
					CAOverlapInterval:   "asdfasfsa",
					CertRotateInterval:  "10h",
					CertOverlapInterval: "8h",
				},
			},
			expectedError: "failed to validate selfSignConfiguration: error parsing caOverlapInterval: time",
		}),
		Entry("When CertRotateInterval is invalid duration string should return an error", validateCase{
			config: &cnao.NetworkAddonsConfigSpec{
				SelfSignConfiguration: &cnao.SelfSignConfiguration{
					CARotateInterval:    "12h",
					CAOverlapInterval:   "10h",
					CertRotateInterval:  "asdfasfa",
					CertOverlapInterval: "8h",
				},
			},
			expectedError: "failed to validate selfSignConfiguration: error parsing certRotateInterval: time",
		}),
		Entry("When CertOverlapInterval is invalid duration string should return an error", validateCase{
			config: &cnao.NetworkAddonsConfigSpec{
				SelfSignConfiguration: &cnao.SelfSignConfiguration{
					CARotateInterval:    "12h",
					CAOverlapInterval:   "10h",
					CertRotateInterval:  "10h",
					CertOverlapInterval: "asdfasfa",
				},
			},
			expectedError: "failed to validate selfSignConfiguration: error parsing certOverlapInterval: time",
		}),
		Entry("When CARotateInterval is zero should return an error", validateCase{
			config: &cnao.NetworkAddonsConfigSpec{
				SelfSignConfiguration: &cnao.SelfSignConfiguration{
					CARotateInterval:    "0",
					CAOverlapInterval:   "12h",
					CertRotateInterval:  "10h",
					CertOverlapInterval: "8h",
				},
			},
			expectedError: "failed to validate selfSignConfiguration: caRotateInterval duration has to be > 0",
		}),
		Entry("When CAOverlapInterval is zero should return an error", validateCase{
			config: &cnao.NetworkAddonsConfigSpec{
				SelfSignConfiguration: &cnao.SelfSignConfiguration{
					CARotateInterval:    "12h",
					CAOverlapInterval:   "0",
					CertRotateInterval:  "10h",
					CertOverlapInterval: "8h",
				},
			},
			expectedError: "failed to validate selfSignConfiguration: caOverlapInterval duration has to be > 0",
		}),
		Entry("When CertRotateInterval is zero should return an error", validateCase{
			config: &cnao.NetworkAddonsConfigSpec{
				SelfSignConfiguration: &cnao.SelfSignConfiguration{
					CARotateInterval:    "12h",
					CAOverlapInterval:   "10h",
					CertRotateInterval:  "0",
					CertOverlapInterval: "8h",
				},
			},
			expectedError: "failed to validate selfSignConfiguration: certRotateInterval duration has to be > 0",
		}),
		Entry("When CertOverlapInterval is zero should return an error", validateCase{
			config: &cnao.NetworkAddonsConfigSpec{
				SelfSignConfiguration: &cnao.SelfSignConfiguration{
					CARotateInterval:    "12h",
					CAOverlapInterval:   "10h",
					CertRotateInterval:  "10h",
					CertOverlapInterval: "0",
				},
			},
			expectedError: "failed to validate selfSignConfiguration: certOverlapInterval duration has to be > 0",
		}),
		Entry("When CAOverlapInterval == CARotateInterval == CertRotateInterval == CertOverlapInterval should not return an error", validateCase{
			config: &cnao.NetworkAddonsConfigSpec{
				SelfSignConfiguration: &cnao.SelfSignConfiguration{
					CARotateInterval:    "1h",
					CAOverlapInterval:   "1h",
					CertRotateInterval:  "1h",
					CertOverlapInterval: "1h",
				},
			},
			expectedError: "",
		}),
		Entry("When CAOverlapInterval is > CARotateInterval should return an error", validateCase{
			config: &cnao.NetworkAddonsConfigSpec{
				SelfSignConfiguration: &cnao.SelfSignConfiguration{
					CARotateInterval:    "1h",
					CAOverlapInterval:   "2h",
					CertRotateInterval:  "30m",
					CertOverlapInterval: "30m",
				},
			},
			expectedError: "failed to validate selfSignConfiguration: caOverlapInterval[(]2h0m0s[)] has to be <= caRotateInterval[(]1h0m0s[)]",
		}),
		Entry("When CertRotateInterval is > CARotateInterval should return an error", validateCase{
			config: &cnao.NetworkAddonsConfigSpec{
				SelfSignConfiguration: &cnao.SelfSignConfiguration{
					CARotateInterval:    "1h",
					CAOverlapInterval:   "30m",
					CertRotateInterval:  "2h",
					CertOverlapInterval: "30m",
				},
			},
			expectedError: "failed to validate selfSignConfiguration: certRotateInterval[(]2h0m0s[)] has to be <= caRotateInterval[(]1h0m0s[)]",
		}),
		Entry("When CertRotateInterval is > CARotateInterval should return an error", validateCase{
			config: &cnao.NetworkAddonsConfigSpec{
				SelfSignConfiguration: &cnao.SelfSignConfiguration{
					CARotateInterval:    "1h",
					CAOverlapInterval:   "30m",
					CertRotateInterval:  "2h",
					CertOverlapInterval: "30m",
				},
			},
			expectedError: "failed to validate selfSignConfiguration: certRotateInterval[(]2h0m0s[)] has to be <= caRotateInterval[(]1h0m0s[)]",
		}),
		Entry("When CertOverlapInterval is > CertRotateInterval should return an error", validateCase{
			config: &cnao.NetworkAddonsConfigSpec{
				SelfSignConfiguration: &cnao.SelfSignConfiguration{
					CARotateInterval:    "1h",
					CAOverlapInterval:   "30m",
					CertRotateInterval:  "1h",
					CertOverlapInterval: "2h",
				},
			},
			expectedError: "failed to validate selfSignConfiguration: certOverlapInterval[(]2h0m0s[)] has to be <= certRotateInterval[(]1h0m0s[)]",
		}),
	)
	var (
		defaultSelfSignConfiguration = cnao.SelfSignConfiguration{
			CARotateInterval:    caRotateIntervalDefault.String(),
			CAOverlapInterval:   caOverlapIntervalDefault.String(),
			CertRotateInterval:  certRotateIntervalDefault.String(),
			CertOverlapInterval: certOverlapIntervalDefault.String(),
		}
	)
	type fillDefaultsCase struct {
		previousConfig *cnao.NetworkAddonsConfigSpec
		currentConfig  *cnao.NetworkAddonsConfigSpec
		expectedConfig cnao.SelfSignConfiguration
	}
	DescribeTable("fill defaults function",
		func(c fillDefaultsCase) {
			errorList := fillDefaultsSelfSignConfiguration(c.currentConfig, c.previousConfig)
			Expect(*c.currentConfig.SelfSignConfiguration).To(Equal(c.expectedConfig))
			Expect(errorList).To(BeEmpty())
		},
		Entry("When SelfSignConfiguration is nil should return default values", fillDefaultsCase{
			previousConfig: &cnao.NetworkAddonsConfigSpec{},
			currentConfig:  &cnao.NetworkAddonsConfigSpec{},
			expectedConfig: defaultSelfSignConfiguration,
		}),
		Entry("When CARotateInterval is empty should return default values", fillDefaultsCase{
			previousConfig: &cnao.NetworkAddonsConfigSpec{},
			currentConfig: &cnao.NetworkAddonsConfigSpec{
				SelfSignConfiguration: &cnao.SelfSignConfiguration{
					CARotateInterval:    "",
					CAOverlapInterval:   "3h",
					CertRotateInterval:  "4h",
					CertOverlapInterval: "5h",
				},
			},
			expectedConfig: defaultSelfSignConfiguration,
		}),
		Entry("When CAOverlapInterval is empty should return default values", fillDefaultsCase{
			previousConfig: &cnao.NetworkAddonsConfigSpec{},
			currentConfig: &cnao.NetworkAddonsConfigSpec{
				SelfSignConfiguration: &cnao.SelfSignConfiguration{
					CARotateInterval:    "4h",
					CAOverlapInterval:   "",
					CertRotateInterval:  "3h",
					CertOverlapInterval: "5h",
				},
			},
			expectedConfig: defaultSelfSignConfiguration,
		}),
		Entry("When CertRotateInterval is empty should return default values", fillDefaultsCase{
			previousConfig: &cnao.NetworkAddonsConfigSpec{},
			currentConfig: &cnao.NetworkAddonsConfigSpec{
				SelfSignConfiguration: &cnao.SelfSignConfiguration{
					CARotateInterval:    "4h",
					CAOverlapInterval:   "3h",
					CertRotateInterval:  "",
					CertOverlapInterval: "5h",
				},
			},
			expectedConfig: defaultSelfSignConfiguration,
		}),
		Entry("When CertOverlapInterval is empty should return default values", fillDefaultsCase{
			previousConfig: &cnao.NetworkAddonsConfigSpec{},
			currentConfig: &cnao.NetworkAddonsConfigSpec{
				SelfSignConfiguration: &cnao.SelfSignConfiguration{
					CARotateInterval:    "4h",
					CAOverlapInterval:   "3h",
					CertRotateInterval:  "3h",
					CertOverlapInterval: "",
				},
			},
			expectedConfig: defaultSelfSignConfiguration,
		}),

		Entry("When the user hasn't explicitly change certificate knobs and a previous selfSignConfiguration exits should use the previous knobs values on the current one", fillDefaultsCase{
			previousConfig: &cnao.NetworkAddonsConfigSpec{
				SelfSignConfiguration: &cnao.SelfSignConfiguration{
					CARotateInterval:    "2h",
					CAOverlapInterval:   "1h",
					CertRotateInterval:  "1h",
					CertOverlapInterval: "1h",
				},
			},
			currentConfig: &cnao.NetworkAddonsConfigSpec{
				SelfSignConfiguration: &cnao.SelfSignConfiguration{
					CARotateInterval:    "",
					CAOverlapInterval:   "",
					CertRotateInterval:  "",
					CertOverlapInterval: "",
				},
			},
			expectedConfig: cnao.SelfSignConfiguration{
				CARotateInterval:    "2h",
				CAOverlapInterval:   "1h",
				CertRotateInterval:  "1h",
				CertOverlapInterval: "1h",
			},
		}),
		Entry("When self sign is nil and a previous conf selfSignConfiguration is missing CARotateInterval should return default values", fillDefaultsCase{
			previousConfig: &cnao.NetworkAddonsConfigSpec{
				SelfSignConfiguration: &cnao.SelfSignConfiguration{
					CARotateInterval:    "",
					CAOverlapInterval:   "1h",
					CertRotateInterval:  "1h",
					CertOverlapInterval: "2h",
				},
			},
			currentConfig:  &cnao.NetworkAddonsConfigSpec{},
			expectedConfig: defaultSelfSignConfiguration,
		}),
		Entry("When self sign is nil and a previous conf selfSignConfiguration is missing CAOverlapInterval should return default values", fillDefaultsCase{
			previousConfig: &cnao.NetworkAddonsConfigSpec{
				SelfSignConfiguration: &cnao.SelfSignConfiguration{
					CARotateInterval:    "1h",
					CAOverlapInterval:   "",
					CertRotateInterval:  "1h",
					CertOverlapInterval: "2h",
				},
			},
			currentConfig:  &cnao.NetworkAddonsConfigSpec{},
			expectedConfig: defaultSelfSignConfiguration,
		}),
		Entry("When self sign is nil and a previous conf selfSignConfiguration is missing CertRotateInterval should return default values", fillDefaultsCase{
			previousConfig: &cnao.NetworkAddonsConfigSpec{
				SelfSignConfiguration: &cnao.SelfSignConfiguration{
					CARotateInterval:    "1h",
					CAOverlapInterval:   "3h",
					CertRotateInterval:  "",
					CertOverlapInterval: "2h",
				},
			},
			currentConfig:  &cnao.NetworkAddonsConfigSpec{},
			expectedConfig: defaultSelfSignConfiguration,
		}),
		Entry("When self sign is nil and a previous conf selfSignConfiguration is missing CertOverlapIntervalOverlapInterval  should return default values", fillDefaultsCase{
			previousConfig: &cnao.NetworkAddonsConfigSpec{
				SelfSignConfiguration: &cnao.SelfSignConfiguration{
					CARotateInterval:    "1h",
					CAOverlapInterval:   "3h",
					CertRotateInterval:  "4h",
					CertOverlapInterval: "",
				},
			},
			currentConfig:  &cnao.NetworkAddonsConfigSpec{},
			expectedConfig: defaultSelfSignConfiguration,
		}),
	)
})
