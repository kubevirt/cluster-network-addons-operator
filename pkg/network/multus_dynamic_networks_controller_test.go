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

		When("`hostCRISocketPath` is configured with a validated value", func() {
			DescribeTable(
				"should use the default value of hostCRISocketPath, and not return an error",
				func(value string) {
					currentClusterConfig := &cnao.NetworkAddonsConfigSpec{
						Multus:                &cnao.Multus{},
						MultusDynamicNetworks: &cnao.MultusDynamicNetworks{HostCRISocketPath: value},
					}
					errorList := validateMultusDynamicNetworks(currentClusterConfig, nil)
					Expect(currentClusterConfig.MultusDynamicNetworks.HostCRISocketPath).To(Equal(value))
					Expect(errorList).To(BeEmpty())
				},
				Entry("is set to \"/run/crio/crio.sock\"", "/run/crio/crio.sock"),
				Entry("is set to \"/run/containerd/containerd.sock\"", "/run/containerd/containerd.sock"),
				Entry("is set to \"/run/k3s/containerd/containerd.sock\"", "/run/k3s/containerd/containerd.sock"),
			)
		})

		When("`hostCRISocketPath` is configured with an invalid value", func() {
			DescribeTable(
				"should fails to be validated",
				func(value string) {
					currentClusterConfig := &cnao.NetworkAddonsConfigSpec{
						Multus:                &cnao.Multus{},
						MultusDynamicNetworks: &cnao.MultusDynamicNetworks{HostCRISocketPath: value},
					}
					errorList := validateMultusDynamicNetworks(currentClusterConfig, nil)
					Expect(errorList).To(ConsistOf(MatchError("`hostCRISocketPath` must be one of: /run/crio/crio.sock, /run/containerd/containerd.sock, /run/k3s/containerd/containerd.sock")))
				},
				Entry("is random string", "Lorem ipsum"),
				Entry("is random path", "/etc/somewhere/some.conf"),
				Entry("is untested socket path", "/run/unknown-cri/unknown-cri.sock"),
			)
		})
	})

	Describe("fill defaults function", func() {
		When("`multusDynamicNetworks` not configured", func() {
			DescribeTable(
				"should NOT return an error and `multusDynamicNetworks` should remain nil",
				func(previousClusterConfig *cnao.NetworkAddonsConfigSpec) {
					currentClusterConfig := &cnao.NetworkAddonsConfigSpec{}
					errorList := fillDefaultsMultusDynamicNetworks(currentClusterConfig, previousClusterConfig)
					Expect(currentClusterConfig.MultusDynamicNetworks).To(BeNil())
					Expect(errorList).To(BeEmpty())
				},
				Entry("with no previous multusDynamicNetworks exists", &cnao.NetworkAddonsConfigSpec{}),
				Entry("with a previous multusDynamicNetworks exists but not parameter configured", &cnao.NetworkAddonsConfigSpec{
					Multus:                &cnao.Multus{},
					MultusDynamicNetworks: &cnao.MultusDynamicNetworks{},
				}),
				Entry("with a previous multusDynamicNetworks exists with custom hostCRISocketPath", &cnao.NetworkAddonsConfigSpec{
					Multus:                &cnao.Multus{},
					MultusDynamicNetworks: &cnao.MultusDynamicNetworks{HostCRISocketPath: "/run/containerd/containerd.sock"},
				}),
			)
		})

		When("the user hasn't configured hostCRISocketPath", func() {
			DescribeTable(
				"should use the default value of hostCRISocketPath, and not return an error",
				func(previousClusterConfig *cnao.NetworkAddonsConfigSpec) {
					currentClusterConfig := &cnao.NetworkAddonsConfigSpec{
						Multus:                &cnao.Multus{},
						MultusDynamicNetworks: &cnao.MultusDynamicNetworks{},
					}
					errorList := fillDefaultsMultusDynamicNetworks(currentClusterConfig, previousClusterConfig)
					Expect(currentClusterConfig.MultusDynamicNetworks.HostCRISocketPath).To(Equal("/run/crio/crio.sock"))
					Expect(errorList).To(BeEmpty())
				},
				Entry("with no previous multusDynamicNetworks exists", &cnao.NetworkAddonsConfigSpec{}),
				Entry("with a previous multusDynamicNetworks exists", &cnao.NetworkAddonsConfigSpec{
					Multus:                &cnao.Multus{},
					MultusDynamicNetworks: &cnao.MultusDynamicNetworks{},
				}),
			)
		})

		When("the user hasn't configured hostCRISocketPath but previous config has one", func() {
			previousClusterConfig := &cnao.NetworkAddonsConfigSpec{
				Multus:                &cnao.Multus{},
				MultusDynamicNetworks: &cnao.MultusDynamicNetworks{HostCRISocketPath: "/run/containerd/containerd.sock"},
			}
			currentClusterConfig := &cnao.NetworkAddonsConfigSpec{
				Multus:                &cnao.Multus{},
				MultusDynamicNetworks: &cnao.MultusDynamicNetworks{},
			}
			It("should use the previous value of hostCRISocketPath, and not return an error", func() {
				errorList := fillDefaultsMultusDynamicNetworks(currentClusterConfig, previousClusterConfig)
				Expect(currentClusterConfig.MultusDynamicNetworks.HostCRISocketPath).To(Equal("/run/containerd/containerd.sock"))
				Expect(errorList).To(BeEmpty())
			})
		})

		When("the user has configured a valid hostCRISocketPath", func() {
			DescribeTable(
				"should use the value of current hostCRISocketPath, and not return an error",
				func(previousClusterConfig *cnao.NetworkAddonsConfigSpec) {
					currentClusterConfig := &cnao.NetworkAddonsConfigSpec{
						Multus:                &cnao.Multus{},
						MultusDynamicNetworks: &cnao.MultusDynamicNetworks{HostCRISocketPath: "/run/containerd/containerd.sock"},
					}
					errorList := fillDefaultsMultusDynamicNetworks(currentClusterConfig, previousClusterConfig)
					Expect(currentClusterConfig.MultusDynamicNetworks.HostCRISocketPath).To(Equal("/run/containerd/containerd.sock"))
					Expect(errorList).To(BeEmpty())
				},
				Entry("with no previous multusDynamicNetworks exists", &cnao.NetworkAddonsConfigSpec{}),
				Entry("with a previous multusDynamicNetworks exists", &cnao.NetworkAddonsConfigSpec{
					Multus:                &cnao.Multus{},
					MultusDynamicNetworks: &cnao.MultusDynamicNetworks{},
				}),
				Entry("with a different previous hostCRISocketPath configured", &cnao.NetworkAddonsConfigSpec{
					Multus:                &cnao.Multus{},
					MultusDynamicNetworks: &cnao.MultusDynamicNetworks{HostCRISocketPath: "/run/crio/crio.sock"},
				}),
			)
		})
	})
})
