package network

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
)

var _ = Describe("Testing macvtap", func() {
	var config cnao.NetworkAddonsConfigSpec

	Context("the device plugin configuration is not provided", func() {
		BeforeEach(func() {
			config = cnao.NetworkAddonsConfigSpec{
				MacvtapCni: &cnao.MacvtapCni{},
			}
		})

		It("defaults to the configuration", func() {
			fillMacvtapDefaults(&config)
			Expect(config).To(Equal(
				cnao.NetworkAddonsConfigSpec{
					MacvtapCni: &cnao.MacvtapCni{
						DevicePluginConfig: "macvtap-deviceplugin-config"},
				}))
		})
	})
})
