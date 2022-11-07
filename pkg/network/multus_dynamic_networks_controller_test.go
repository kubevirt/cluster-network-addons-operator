package network

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	openshiftoperatorv1 "github.com/openshift/api/operator/v1"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
)

var _ = Describe("Testing multus-dynamic-networks", func() {
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
})
