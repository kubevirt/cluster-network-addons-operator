package network

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	openshiftoperatorv1 "github.com/openshift/api/operator/v1"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
)

var _ = Describe("Testing multus-dynamic-networks", func() {
	Describe("validation function", func() {
		Context("the correct CNAO configuration is passed", func() {
			var config cnao.NetworkAddonsConfigSpec

			BeforeEach(func() {
				config = cnao.NetworkAddonsConfigSpec{
					Multus:                &cnao.Multus{},
					MultusDynamicNetworks: &cnao.MultusDynamicNetworks{},
				}
			})

			It("is successfully validated", func() {
				Expect(validateMultusDynamicNetworks(&config, nil)).To(BeEmpty())
			})
		})

		Context("the configuration is missing the multus dependency", func() {
			var config cnao.NetworkAddonsConfigSpec

			BeforeEach(func() {
				config = cnao.NetworkAddonsConfigSpec{
					MultusDynamicNetworks: &cnao.MultusDynamicNetworks{},
				}
			})

			It("fails to be validated", func() {
				Expect(
					validateMultusDynamicNetworks(&config, nil),
				).To(ConsistOf(MatchError("the `multus` configuration is required")))
			})
		})

		When("`multusDynamicNetworks` is not configured", func() {
			DescribeTable(
				"the configuration is validated",
				func(cnaoConfig *cnao.NetworkAddonsConfigSpec, openshiftNetworkConfig *openshiftoperatorv1.Network) {
					Expect(validateMultusDynamicNetworks(cnaoConfig, openshiftNetworkConfig)).To(BeEmpty())
				},
				Entry("without an openshift network configuration", &cnao.NetworkAddonsConfigSpec{}, nil),
				Entry("on an openshift cluster", &cnao.NetworkAddonsConfigSpec{}, &openshiftoperatorv1.Network{}),
			)
		})

		When("an openshift network configuration in passed", func() {
			It("a valid configuration fails", func() {
				openshiftNetworkConfig := &openshiftoperatorv1.Network{}
				cnaoConfig := &cnao.NetworkAddonsConfigSpec{
					Multus:                &cnao.Multus{},
					MultusDynamicNetworks: &cnao.MultusDynamicNetworks{},
				}
				Expect(validateMultusDynamicNetworks(cnaoConfig, openshiftNetworkConfig)).To(ConsistOf(MatchError("`multusDynamicNetworks` configuration is not supported on Openshift yet")))
			})
		})

		When("`hostCriSocketPath` is not configured", func() {
			config := &cnao.NetworkAddonsConfigSpec{
				Multus:                &cnao.Multus{},
				MultusDynamicNetworks: &cnao.MultusDynamicNetworks{},
			}
			It("is successfully validated", func() {
				Expect(validateMultusDynamicNetworks(config, nil)).To(BeEmpty())
			})
		})

		When("`hostCriSocketPath` is configured", func() {
			config := &cnao.NetworkAddonsConfigSpec{
				Multus:                &cnao.Multus{},
				MultusDynamicNetworks: &cnao.MultusDynamicNetworks{HostCRISocketPath: "/run/containerd/containerd.sock"},
			}
			It("is successfully validated", func() {
				Expect(validateMultusDynamicNetworks(config, nil)).To(BeEmpty())
			})
		})
	})

	Describe("fill defaults function", func() {
		When("`multusDynamicNetworks` is not configured", func() {
			It("should NOT return an error", func() {
				currentClusterConfig := &cnao.NetworkAddonsConfigSpec{}
				previousClusterConfig := &cnao.NetworkAddonsConfigSpec{
					Multus:                &cnao.Multus{},
					MultusDynamicNetworks: &cnao.MultusDynamicNetworks{},
				}
				errorList := fillDefaultsMultusDynamicNetworks(currentClusterConfig, previousClusterConfig)
				Expect(currentClusterConfig.MultusDynamicNetworks).To(BeNil())
				Expect(errorList).To(BeEmpty())
			})
		})

		When("the user hasn't configured hostCriSocketPath", func() {
			DescribeTable(
				"should use the default value of hostCriSocketPath, and not return an error",
				func(previousClusterConfig *cnao.NetworkAddonsConfigSpec) {
					currentClusterConfig := &cnao.NetworkAddonsConfigSpec{
						Multus:                &cnao.Multus{},
						MultusDynamicNetworks: &cnao.MultusDynamicNetworks{},
					}
					fillDefaultsMultusDynamicNetworks(currentClusterConfig, previousClusterConfig)
					Expect(currentClusterConfig.MultusDynamicNetworks.HostCRISocketPath).To(Equal("/run/crio/crio.sock"))
				},
				Entry("with no previous multusDynamicNetworks exists", &cnao.NetworkAddonsConfigSpec{}),
				Entry("with a previous multusDynamicNetworks exists", &cnao.NetworkAddonsConfigSpec{
					Multus:                &cnao.Multus{},
					MultusDynamicNetworks: &cnao.MultusDynamicNetworks{},
				}),
				Entry("with a previous hostCriSocketPath configured", &cnao.NetworkAddonsConfigSpec{
					Multus:                &cnao.Multus{},
					MultusDynamicNetworks: &cnao.MultusDynamicNetworks{HostCRISocketPath: "/run/containerd/containerd.sock"},
				}),
			)
		})

		When("the user has configured a valid hostCriSocketPath", func() {
			DescribeTable(
				"should use the value of current hostCriSocketPath, and not return an error",
				func(previousClusterConfig *cnao.NetworkAddonsConfigSpec) {
					currentClusterConfig := &cnao.NetworkAddonsConfigSpec{
						Multus:                &cnao.Multus{},
						MultusDynamicNetworks: &cnao.MultusDynamicNetworks{HostCRISocketPath: "/run/containerd/containerd.sock"},
					}
					fillDefaultsMultusDynamicNetworks(currentClusterConfig, previousClusterConfig)
					Expect(currentClusterConfig.MultusDynamicNetworks.HostCRISocketPath).To(Equal("/run/containerd/containerd.sock"))
				},
				Entry("with no previous multusDynamicNetworks exists", &cnao.NetworkAddonsConfigSpec{}),
				Entry("with a previous multusDynamicNetworks exists", &cnao.NetworkAddonsConfigSpec{
					Multus:                &cnao.Multus{},
					MultusDynamicNetworks: &cnao.MultusDynamicNetworks{},
				}),
				Entry("with a different previous hostCriSocketPath configured", &cnao.NetworkAddonsConfigSpec{
					Multus:                &cnao.Multus{},
					MultusDynamicNetworks: &cnao.MultusDynamicNetworks{HostCRISocketPath: "/run/crio/crio.sock"},
				}),
			)
		})
	})

	Describe("change safe function", func() {
		Context("When they are equal", func() {
			It("should NOT return an error", func() {
				previousClusterConfig := &cnao.NetworkAddonsConfigSpec{
					Multus:                &cnao.Multus{},
					MultusDynamicNetworks: &cnao.MultusDynamicNetworks{HostCRISocketPath: "/run/containerd/containerd.sock"},
				}
				currentClusterConfig := &cnao.NetworkAddonsConfigSpec{
					Multus:                &cnao.Multus{},
					MultusDynamicNetworks: &cnao.MultusDynamicNetworks{HostCRISocketPath: "/run/containerd/containerd.sock"},
				}

				errorList := changeSafeMultusDynamicNetworks(previousClusterConfig, currentClusterConfig)
				Expect(errorList).To(BeEmpty())
			})
		})

		Context("When only hostCriSocketPath different", func() {
			It("should NOT return an error", func() {
				previousClusterConfig := &cnao.NetworkAddonsConfigSpec{
					Multus:                &cnao.Multus{},
					MultusDynamicNetworks: &cnao.MultusDynamicNetworks{HostCRISocketPath: "/run/crio/crio.sock"},
				}
				currentClusterConfig := &cnao.NetworkAddonsConfigSpec{
					Multus:                &cnao.Multus{},
					MultusDynamicNetworks: &cnao.MultusDynamicNetworks{HostCRISocketPath: "/run/containerd/containerd.sock"},
				}
				errorList := changeSafeMultusDynamicNetworks(previousClusterConfig, currentClusterConfig)
				Expect(errorList).To(BeEmpty(), "hostCriSocketPath changes should be allowed")
			})
		})

		Context("When trying to remove multusDynamicNetworks", func() {
			It("should NOT return an error", func() {
				previousClusterConfig := &cnao.NetworkAddonsConfigSpec{
					Multus:                &cnao.Multus{},
					MultusDynamicNetworks: &cnao.MultusDynamicNetworks{HostCRISocketPath: "/run/crio/crio.sock"},
				}
				currentClusterConfig := &cnao.NetworkAddonsConfigSpec{
					Multus: &cnao.Multus{},
				}

				errorList := changeSafeMultusDynamicNetworks(previousClusterConfig, currentClusterConfig)
				Expect(errorList).To(BeEmpty())
			})
		})
	})
})
