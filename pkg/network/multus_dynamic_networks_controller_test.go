package network

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
)

var _ = Describe("Testing multus-dynamic-networks", func() {
	Context("the correct configuration is passed", func() {
		var config cnao.NetworkAddonsConfigSpec

		BeforeEach(func() {
			config = cnao.NetworkAddonsConfigSpec{
				Multus:                &cnao.Multus{},
				MultusDynamicNetworks: &cnao.MultusDynamicNetworks{},
			}
		})

		It("is successfully validated", func() {
			Expect(validateMultusDynamicNetworks(&config)).To(BeEmpty())
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
				validateMultusDynamicNetworks(&config),
			).To(ConsistOf(MatchError("the `multus` configuration is required")))
		})
	})
})
