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

		It("uses a default config map name", func() {
			fillMacvtapDefaults(&config, nil)
			Expect(config).To(Equal(
				cnao.NetworkAddonsConfigSpec{
					MacvtapCni: &cnao.MacvtapCni{
						DevicePluginConfig: "macvtap-deviceplugin-config"},
				}))
		})

		It("uses the previous configuration config map name when available", func() {
			const configMapNameFromOldDays = "ldfj239"
			oldConfig := cnao.NetworkAddonsConfigSpec{
				MacvtapCni: &cnao.MacvtapCni{DevicePluginConfig: configMapNameFromOldDays},
			}
			fillMacvtapDefaults(&config, &oldConfig)
			Expect(config).To(Equal(
				cnao.NetworkAddonsConfigSpec{
					MacvtapCni: &cnao.MacvtapCni{
						DevicePluginConfig: configMapNameFromOldDays},
				}))
		})
	})

	It("uses the device plugin configuration defined in the macvtap config", func() {
		const (
			configMapNameFromOldDays = "ldfj239"
			newConfigMapName         = "jasfjaofj"
		)
		oldConfig := cnao.NetworkAddonsConfigSpec{
			MacvtapCni: &cnao.MacvtapCni{DevicePluginConfig: configMapNameFromOldDays},
		}
		newConfig := cnao.NetworkAddonsConfigSpec{
			MacvtapCni: &cnao.MacvtapCni{DevicePluginConfig: newConfigMapName},
		}
		fillMacvtapDefaults(&newConfig, &oldConfig)
		Expect(newConfig).To(Equal(
			cnao.NetworkAddonsConfigSpec{
				MacvtapCni: &cnao.MacvtapCni{
					DevicePluginConfig: newConfigMapName},
			}))
	})
})
