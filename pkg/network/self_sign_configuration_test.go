package network

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	opv1alpha1 "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/v1alpha1"
)

var _ = Describe("Testing SelfSignConfiguration", func() {
	type validateCase struct {
		config        *opv1alpha1.NetworkAddonsConfigSpec
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
		Entry("when CARotateInterval is the only empty one should return an error", validateCase{
			config: &opv1alpha1.NetworkAddonsConfigSpec{
				SelfSignConfiguration: &opv1alpha1.SelfSignConfiguration{
					CARotateInterval:   "",
					CAOverlapInterval:  "48h30m50s",
					CertRotateInterval: "48h30m50s",
				},
			},
			expectedError: "failed to validate selfSignConfiguration: caRotateInterval is missing",
		}),
		Entry("when CAOverlapInterval is the only empty one should return an error", validateCase{
			config: &opv1alpha1.NetworkAddonsConfigSpec{
				SelfSignConfiguration: &opv1alpha1.SelfSignConfiguration{
					CARotateInterval:   "48h30m50s",
					CAOverlapInterval:  "",
					CertRotateInterval: "48h30m50s",
				},
			},
			expectedError: "failed to validate selfSignConfiguration: caOverlapInterval is missing",
		}),
		Entry("when CertRotateInterval is the only empty one should return an error", validateCase{
			config: &opv1alpha1.NetworkAddonsConfigSpec{
				SelfSignConfiguration: &opv1alpha1.SelfSignConfiguration{
					CARotateInterval:   "48h30m50s",
					CAOverlapInterval:  "48h30m50s",
					CertRotateInterval: "",
				},
			},
			expectedError: "failed to validate selfSignConfiguration: certRotateInterval is missing",
		}),
		Entry("when selfSignConfiguration is valid sould not return an error", validateCase{
			config: &opv1alpha1.NetworkAddonsConfigSpec{
				SelfSignConfiguration: &opv1alpha1.SelfSignConfiguration{
					CARotateInterval:   "24h",
					CAOverlapInterval:  "12h",
					CertRotateInterval: "10h",
				},
			},
			expectedError: "",
		}),
		Entry("when CARotateInterval is invalid duration string should return an error", validateCase{
			config: &opv1alpha1.NetworkAddonsConfigSpec{
				SelfSignConfiguration: &opv1alpha1.SelfSignConfiguration{
					CARotateInterval:   "asdfasfda",
					CAOverlapInterval:  "12h",
					CertRotateInterval: "10h",
				},
			},
			expectedError: "failed to validate selfSignConfiguration: error parsing caRotateInterval: time",
		}),
		Entry("when CAOverlapInterval is invalid duration string should return an error", validateCase{
			config: &opv1alpha1.NetworkAddonsConfigSpec{
				SelfSignConfiguration: &opv1alpha1.SelfSignConfiguration{
					CARotateInterval:   "12h",
					CAOverlapInterval:  "asdfasfsa",
					CertRotateInterval: "10h",
				},
			},
			expectedError: "failed to validate selfSignConfiguration: error parsing caOverlapInterval: time",
		}),
		Entry("when CertRotateInterval is invalid duration string should return an error", validateCase{
			config: &opv1alpha1.NetworkAddonsConfigSpec{
				SelfSignConfiguration: &opv1alpha1.SelfSignConfiguration{
					CARotateInterval:   "12h",
					CAOverlapInterval:  "10h",
					CertRotateInterval: "asdfasfa",
				},
			},
			expectedError: "failed to validate selfSignConfiguration: error parsing certRotateInterval: time",
		}),
		Entry("when CARotateInterval is zero should return an error", validateCase{
			config: &opv1alpha1.NetworkAddonsConfigSpec{
				SelfSignConfiguration: &opv1alpha1.SelfSignConfiguration{
					CARotateInterval:   "0",
					CAOverlapInterval:  "12h",
					CertRotateInterval: "10h",
				},
			},
			expectedError: "failed to validate selfSignConfiguration: caRotateInterval duration has to be > 0",
		}),
		Entry("when CAOverlapInterval is zero should return an error", validateCase{
			config: &opv1alpha1.NetworkAddonsConfigSpec{
				SelfSignConfiguration: &opv1alpha1.SelfSignConfiguration{
					CARotateInterval:   "12h",
					CAOverlapInterval:  "0",
					CertRotateInterval: "10h",
				},
			},
			expectedError: "failed to validate selfSignConfiguration: caOverlapInterval duration has to be > 0",
		}),
		Entry("when CertRotateInterval is zero should return an error", validateCase{
			config: &opv1alpha1.NetworkAddonsConfigSpec{
				SelfSignConfiguration: &opv1alpha1.SelfSignConfiguration{
					CARotateInterval:   "12h",
					CAOverlapInterval:  "10h",
					CertRotateInterval: "0",
				},
			},
			expectedError: "failed to validate selfSignConfiguration: certRotateInterval duration has to be > 0",
		}),
		Entry("when CAOverlapInterval == CARotateInterval == CertRotateInterval should not return an error", validateCase{
			config: &opv1alpha1.NetworkAddonsConfigSpec{
				SelfSignConfiguration: &opv1alpha1.SelfSignConfiguration{
					CARotateInterval:   "1h",
					CAOverlapInterval:  "1h",
					CertRotateInterval: "1h",
				},
			},
			expectedError: "",
		}),
		Entry("when CAOverlapInterval is > CARotateInterval should return an error", validateCase{
			config: &opv1alpha1.NetworkAddonsConfigSpec{
				SelfSignConfiguration: &opv1alpha1.SelfSignConfiguration{
					CARotateInterval:   "1h",
					CAOverlapInterval:  "2h",
					CertRotateInterval: "30m",
				},
			},
			expectedError: "failed to validate selfSignConfiguration: caOverlapInterval[(]2h0m0s[)] has to be <= caRotateInterval[(]1h0m0s[)]",
		}),
		Entry("when CertRotateInterval is > CARotateInterval should return an error", validateCase{
			config: &opv1alpha1.NetworkAddonsConfigSpec{
				SelfSignConfiguration: &opv1alpha1.SelfSignConfiguration{
					CARotateInterval:   "1h",
					CAOverlapInterval:  "30m",
					CertRotateInterval: "2h",
				},
			},
			expectedError: "failed to validate selfSignConfiguration: certRotateInterval[(]2h0m0s[)] has to be <= caRotateInterval[(]1h0m0s[)]",
		}),
	)
	var (
		defaultSelfSignConfiguration = opv1alpha1.SelfSignConfiguration{
			CARotateInterval:   caRotateIntervalDefault.String(),
			CAOverlapInterval:  caOverlapIntervalDefault.String(),
			CertRotateInterval: certRotateIntervalDefault.String(),
		}
	)
	type fillDefaultsCase struct {
		previousConfig *opv1alpha1.NetworkAddonsConfigSpec
		currentConfig  *opv1alpha1.NetworkAddonsConfigSpec
		expectedConfig opv1alpha1.SelfSignConfiguration
	}
	DescribeTable("fill defaults function",
		func(c fillDefaultsCase) {
			errorList := fillDefaultsSelfSignConfiguration(c.currentConfig, c.previousConfig)
			Expect(*c.currentConfig.SelfSignConfiguration).To(Equal(c.expectedConfig))
			Expect(errorList).To(BeEmpty())
		},
		Entry("when SelfSignConfiguration is nil should return default values", fillDefaultsCase{
			previousConfig: &opv1alpha1.NetworkAddonsConfigSpec{},
			currentConfig:  &opv1alpha1.NetworkAddonsConfigSpec{},
			expectedConfig: defaultSelfSignConfiguration,
		}),
		Entry("when CARotateInterval is empty should return default values", fillDefaultsCase{
			previousConfig: &opv1alpha1.NetworkAddonsConfigSpec{},
			currentConfig: &opv1alpha1.NetworkAddonsConfigSpec{
				SelfSignConfiguration: &opv1alpha1.SelfSignConfiguration{
					CARotateInterval:   "",
					CAOverlapInterval:  "3h",
					CertRotateInterval: "4h",
				},
			},
			expectedConfig: defaultSelfSignConfiguration,
		}),
		Entry("when CAOverlapInterval is empty should return default values", fillDefaultsCase{
			previousConfig: &opv1alpha1.NetworkAddonsConfigSpec{},
			currentConfig: &opv1alpha1.NetworkAddonsConfigSpec{
				SelfSignConfiguration: &opv1alpha1.SelfSignConfiguration{
					CARotateInterval:   "4h",
					CAOverlapInterval:  "",
					CertRotateInterval: "3h",
				},
			},
			expectedConfig: defaultSelfSignConfiguration,
		}),
		Entry("when CertRotateInterval is empty should return default values", fillDefaultsCase{
			previousConfig: &opv1alpha1.NetworkAddonsConfigSpec{},
			currentConfig: &opv1alpha1.NetworkAddonsConfigSpec{
				SelfSignConfiguration: &opv1alpha1.SelfSignConfiguration{
					CARotateInterval:   "4h",
					CAOverlapInterval:  "3h",
					CertRotateInterval: "",
				},
			},
			expectedConfig: defaultSelfSignConfiguration,
		}),
		Entry("when the user hasn't explicitly change certificate knobs and a previous selfSignConfiguration exits should use the previous knobs values on the current one", fillDefaultsCase{
			previousConfig: &opv1alpha1.NetworkAddonsConfigSpec{
				SelfSignConfiguration: &opv1alpha1.SelfSignConfiguration{
					CARotateInterval:   "2h",
					CAOverlapInterval:  "1h",
					CertRotateInterval: "1h",
				},
			},
			currentConfig: &opv1alpha1.NetworkAddonsConfigSpec{
				SelfSignConfiguration: &opv1alpha1.SelfSignConfiguration{
					CARotateInterval:   "",
					CAOverlapInterval:  "",
					CertRotateInterval: "",
				},
			},
			expectedConfig: opv1alpha1.SelfSignConfiguration{
				CARotateInterval:   "2h",
				CAOverlapInterval:  "1h",
				CertRotateInterval: "1h",
			},
		}),
	)
	Describe("change safe function", func() {
		Context("When they are equal", func() {
			It("should NOT return an error", func() {
				previousClusterConfig := &opv1alpha1.NetworkAddonsConfigSpec{
					SelfSignConfiguration: &opv1alpha1.SelfSignConfiguration{
						CARotateInterval:   "2h",
						CAOverlapInterval:  "1h",
						CertRotateInterval: "1h",
					},
				}
				currentClusterConfig := &opv1alpha1.NetworkAddonsConfigSpec{
					SelfSignConfiguration: &opv1alpha1.SelfSignConfiguration{
						CARotateInterval:   "2h",
						CAOverlapInterval:  "1h",
						CertRotateInterval: "1h",
					},
				}
				errorList := changeSafeSelfSignConfiguration(previousClusterConfig, currentClusterConfig)
				Expect(errorList).To(BeEmpty())
			})
		})

		Context("When they are not equal", func() {
			It("should return an error", func() {
				previousClusterConfig := &opv1alpha1.NetworkAddonsConfigSpec{
					SelfSignConfiguration: &opv1alpha1.SelfSignConfiguration{
						CARotateInterval:   "2h",
						CAOverlapInterval:  "1h",
						CertRotateInterval: "1h",
					},
				}
				currentClusterConfig := &opv1alpha1.NetworkAddonsConfigSpec{
					SelfSignConfiguration: &opv1alpha1.SelfSignConfiguration{
						CARotateInterval:   "4h",
						CAOverlapInterval:  "1h",
						CertRotateInterval: "1h",
					},
				}
				errorList := changeSafeSelfSignConfiguration(previousClusterConfig, currentClusterConfig)
				Expect(len(errorList)).To(Equal(1), "validation failed due to an unexpected error: %v", errorList)
				Expect(errorList[0].Error()).To(Equal("cannot modify SelfSignConfiguration configuration once it is deployed"))
			})
		})

		Context("When trying to remove selfSignConfiguration", func() {
			It("should NOT return an error", func() {
				previousClusterConfig := &opv1alpha1.NetworkAddonsConfigSpec{
					SelfSignConfiguration: &opv1alpha1.SelfSignConfiguration{
						CARotateInterval:   "2h",
						CAOverlapInterval:  "1h",
						CertRotateInterval: "1h",
					},
				}
				currentClusterConfig := &opv1alpha1.NetworkAddonsConfigSpec{}
				errorList := changeSafeSelfSignConfiguration(previousClusterConfig, currentClusterConfig)
				Expect(errorList).To(BeEmpty())
			})
		})
	})
})
