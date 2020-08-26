package network

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
)

var _ = Describe("Testing kubeMacPool", func() {
	Describe("validation function", func() {
		Context("When kubeMacPool is nil ", func() {
			It("should NOT return an error because there is nothing to validate", func() {
				clusterConfig := &cnao.NetworkAddonsConfigSpec{}
				errorList := validateKubeMacPool(clusterConfig)
				Expect(errorList).To(BeEmpty())
			})
		})

		Context("When both ranges are configured to be empty", func() {
			It("should NOT return an error because both or none of the ranges can be empty", func() {
				clusterConfig := &cnao.NetworkAddonsConfigSpec{KubeMacPool: &cnao.KubeMacPool{RangeStart: "", RangeEnd: ""}}
				errorList := validateKubeMacPool(clusterConfig)
				Expect(errorList).To(BeEmpty())
			})
		})

		Context("When both ranges are not configured", func() {
			It("should NOT return an error because both or none of the ranges can be configured", func() {
				clusterConfig := &cnao.NetworkAddonsConfigSpec{KubeMacPool: &cnao.KubeMacPool{}}
				errorList := validateKubeMacPool(clusterConfig)
				Expect(errorList).To(BeEmpty())
			})
		})

		Context("When only RangeStart is configured", func() {
			It("should return an error because both or none of the ranges must be configured", func() {
				clusterConfig := &cnao.NetworkAddonsConfigSpec{
					KubeMacPool: &cnao.KubeMacPool{RangeStart: "02:00:00:00:00:00", RangeEnd: ""}}
				errorList := validateKubeMacPool(clusterConfig)
				Expect(len(errorList)).To(Equal(1), "validation failed due to an unexpected error: %v", errorList)
				Expect(errorList[0].Error()).To(Equal("both or none of the KubeMacPool ranges needs to be configured"))
			})
		})

		Context("When only RangeEnd is configured", func() {
			It("should return an error because both or none of the ranges must be configured", func() {
				clusterConfig := &cnao.NetworkAddonsConfigSpec{
					KubeMacPool: &cnao.KubeMacPool{RangeStart: "", RangeEnd: "02:00:00:FF:FF:FF"}}
				errorList := validateKubeMacPool(clusterConfig)
				Expect(len(errorList)).To(Equal(1), "validation failed due to an unexpected error: %v", errorList)
				Expect(errorList[0].Error()).To(Equal("both or none of the KubeMacPool ranges needs to be configured"))
			})
		})

		Context("When RangeStart contains 7 octets instead of 6", func() {
			It("should return an error because the mac address set for RangeStart is invalid", func() {
				clusterConfig := &cnao.NetworkAddonsConfigSpec{
					KubeMacPool: &cnao.KubeMacPool{RangeStart: "00:00:00:00:00:00:00", RangeEnd: "02:FF:FF:FF:FF:FF"}}
				errorList := validateKubeMacPool(clusterConfig)
				Expect(len(errorList)).To(Equal(1), "validation failed due to an unexpected error: %v", errorList)
				Expect(errorList[0].Error()).To(Equal("failed to parse rangeStart because the mac address is invalid"))

			})
		})

		Context("When RangeEnd contains 7 octets instead of 6", func() {
			It("should return an error because the mac address set for RangeEnd is invalid", func() {
				clusterConfig := &cnao.NetworkAddonsConfigSpec{
					KubeMacPool: &cnao.KubeMacPool{RangeStart: "02:00:00:00:00:00", RangeEnd: "FF:FF:FF:FF:FF:FF:FF"}}
				errorList := validateKubeMacPool(clusterConfig)
				Expect(len(errorList)).To(Equal(1), "validation failed due to an unexpected errors: %v", errorList)
				Expect(errorList[0].Error()).To(Equal("failed to parse rangeEnd because the mac address is invalid"))
			})
		})

		Context("When range end is lesser than its start", func() {
			It("should return an error", func() {
				clusterConfig := &cnao.NetworkAddonsConfigSpec{
					KubeMacPool: &cnao.KubeMacPool{RangeStart: "02:00:00:ff:00:00", RangeEnd: "02:00:00:00:00:00"}}
				errorList := validateKubeMacPool(clusterConfig)
				Expect(len(errorList)).To(Equal(1), "validation failed due to an unexpected error: %v", errorList)
				Expect(errorList[0].Error()).To(Equal("failed to set mac address range: invalid range. Range end is lesser than or equal to its start. start: 02:00:00:ff:00:00 end: 02:00:00:00:00:00"))
			})
		})

		Context("When range end is the same as its start", func() {
			It("should return an error", func() {
				clusterConfig := &cnao.NetworkAddonsConfigSpec{
					KubeMacPool: &cnao.KubeMacPool{RangeStart: "02:00:00:ff:00:00", RangeEnd: "02:00:00:ff:00:00"}}
				errorList := validateKubeMacPool(clusterConfig)
				Expect(len(errorList)).To(Equal(1), "validation failed due to an unexpected error: %v", errorList)
				Expect(errorList[0].Error()).To(Equal("failed to set mac address range: invalid range. Range end is lesser than or equal to its start. start: 02:00:00:ff:00:00 end: 02:00:00:ff:00:00"))
			})
		})

		Context("when range end is greater than its start only by 2", func() {
			It("should NOT return an error", func() {
				clusterConfig := &cnao.NetworkAddonsConfigSpec{
					KubeMacPool: &cnao.KubeMacPool{RangeStart: "02:00:00:ff:00:00", RangeEnd: "02:00:00:ff:00:02"}}
				errorList := validateKubeMacPool(clusterConfig)
				Expect(errorList).To(BeEmpty())
			})
		})

		Context("When the multicast bit is on in rangeStart", func() {
			It("should return an error because unicast addressing must be used", func() {
				clusterConfig := &cnao.NetworkAddonsConfigSpec{
					KubeMacPool: &cnao.KubeMacPool{RangeStart: "01:00:00:00:00:00", RangeEnd: "06:FF:FF:FF:FF:FF"}}
				errorList := validateKubeMacPool(clusterConfig)
				Expect(len(errorList)).To(Equal(1), "validation failed due to an unexpected error: %v", errorList)
				Expect(errorList[0].Error()).To(Equal("failed to set RangeStart: invalid mac address. Multicast addressing is not supported. Unicast addressing must be used. The first octet is 0X1"))
			})
		})

		Context("When the multicast bit is on in RangeEnd", func() {
			It("should return an error because unicast addressing must be used", func() {
				clusterConfig := &cnao.NetworkAddonsConfigSpec{
					KubeMacPool: &cnao.KubeMacPool{RangeStart: "02:00:00:00:00:00", RangeEnd: "03:FF:FF:FF:FF:FF"}}
				errorList := validateKubeMacPool(clusterConfig)
				Expect(len(errorList)).To(Equal(1), "validation failed due to an unexpected error: %v", errorList)
				Expect(errorList[0].Error()).To(Equal("failed to set RangeEnd: invalid mac address. Multicast addressing is not supported. Unicast addressing must be used. The first octet is 0X3"))
			})
		})

		Context("When the mac address is valid and multicast bit is off", func() {
			It("should NOT return an error", func() {
				clusterConfig := &cnao.NetworkAddonsConfigSpec{
					KubeMacPool: &cnao.KubeMacPool{RangeStart: "02:00:00:00:00:00", RangeEnd: "0A:FF:FF:FF:FF:FF"}}
				errorList := validateKubeMacPool(clusterConfig)
				Expect(errorList).To(BeEmpty())
			})
		})
	})

	Describe("fill defaults function", func() {
		Context("When kubeMacPool is nil", func() {
			It("should NOT return an error", func() {
				currentClusterConfig := &cnao.NetworkAddonsConfigSpec{}
				previousClusterConfig := &cnao.NetworkAddonsConfigSpec{
					KubeMacPool: &cnao.KubeMacPool{RangeStart: "02:00:00:00:00:00", RangeEnd: "0A:FF:FF:FF:FF:FF"}}
				errorList := fillDefaultsKubeMacPool(currentClusterConfig, previousClusterConfig)
				Expect(currentClusterConfig.KubeMacPool).To(BeNil())
				Expect(errorList).To(BeEmpty())
			})
		})

		Context("When the user hasn't explicitly requested a range", func() {
			Context("When a previous kubeMacPool exits", func() {
				It("should use the previous range for the current one, and not return an error", func() {
					previousClusterConfig := &cnao.NetworkAddonsConfigSpec{
						KubeMacPool: &cnao.KubeMacPool{RangeStart: "02:00:00:00:00:00", RangeEnd: "0A:FF:FF:FF:FF:FF"}}
					currentClusterConfig := &cnao.NetworkAddonsConfigSpec{
						KubeMacPool: &cnao.KubeMacPool{RangeStart: "", RangeEnd: ""}}
					errorList := fillDefaultsKubeMacPool(currentClusterConfig, previousClusterConfig)
					Expect(errorList).To(BeEmpty())
					Expect(currentClusterConfig.KubeMacPool.RangeStart).To(Equal(previousClusterConfig.KubeMacPool.RangeStart))
					Expect(currentClusterConfig.KubeMacPool.RangeEnd).To(Equal(previousClusterConfig.KubeMacPool.RangeEnd))

				})
			})

			Context("When a previous kubeMacPool doesn't exits", func() {
				It("should generate a new range for the current kubeMacPool, and not return an error", func() {
					previousClusterConfig := &cnao.NetworkAddonsConfigSpec{}
					currentClusterConfig := &cnao.NetworkAddonsConfigSpec{
						KubeMacPool: &cnao.KubeMacPool{}}
					errorList := fillDefaultsKubeMacPool(currentClusterConfig, previousClusterConfig)
					Expect(errorList).To(BeEmpty())
					Expect(currentClusterConfig.KubeMacPool.RangeStart).To(Not(Equal("")), "RangeStart should not be empty:")
					Expect(currentClusterConfig.KubeMacPool.RangeEnd).To(Not(Equal("")), "RangeEnd should not be empty")

				})
			})

		})

		Context("When the user has explicitly requested a range", func() {
			It("should leave the range as it is, and not return an error", func() {
				currentRangeStart := "02:00:00:ff:00:00"
				currentRangeEnd := "02:00:00:ff:00:02"

				previousClusterConfig := &cnao.NetworkAddonsConfigSpec{
					KubeMacPool: &cnao.KubeMacPool{RangeStart: "02:00:00:00:00:00", RangeEnd: "0A:FF:FF:FF:FF:FF"}}
				currentClusterConfig := &cnao.NetworkAddonsConfigSpec{
					KubeMacPool: &cnao.KubeMacPool{RangeStart: currentRangeStart, RangeEnd: currentRangeEnd}}
				errorList := fillDefaultsKubeMacPool(currentClusterConfig, previousClusterConfig)
				Expect(errorList).To(BeEmpty())
				Expect(currentClusterConfig.KubeMacPool.RangeStart).To(Equal(currentRangeStart), "RangeStart should be as the user explicitly requested")
				Expect(currentClusterConfig.KubeMacPool.RangeEnd).To(Equal(currentRangeEnd), "RangeEnd should be as the user explicitly requested")

			})
		})

	})

	Describe("change safe function", func() {
		Context("When they are equal", func() {
			It("should NOT return an error", func() {
				previousClusterConfig := &cnao.NetworkAddonsConfigSpec{
					KubeMacPool: &cnao.KubeMacPool{RangeStart: "02:00:00:00:00:00", RangeEnd: "0A:FF:FF:FF:FF:FF"}}
				currentClusterConfig := &cnao.NetworkAddonsConfigSpec{
					KubeMacPool: &cnao.KubeMacPool{RangeStart: "02:00:00:00:00:00", RangeEnd: "0A:FF:FF:FF:FF:FF"}}

				errorList := changeSafeKubeMacPool(previousClusterConfig, currentClusterConfig)
				Expect(errorList).To(BeEmpty())
			})
		})

		Context("When they are not equal", func() {
			It("should return an error", func() {
				previousClusterConfig := &cnao.NetworkAddonsConfigSpec{
					KubeMacPool: &cnao.KubeMacPool{RangeStart: "02:00:00:00:00:00", RangeEnd: "0A:FF:FF:FF:FF:FF"}}
				currentClusterConfig := &cnao.NetworkAddonsConfigSpec{
					KubeMacPool: &cnao.KubeMacPool{RangeStart: "02:00:00:00:00:00", RangeEnd: "0A:FF:FF:FF:FF:F1"}}

				errorList := changeSafeKubeMacPool(previousClusterConfig, currentClusterConfig)
				Expect(len(errorList)).To(Equal(1), "validation failed due to an unexpected error: %v", errorList)
				Expect(errorList[0].Error()).To(Equal("cannot modify KubeMacPool configuration once it is deployed"))
			})
		})

		Context("When trying to remove kubeMacPool", func() {
			It("should NOT return an error", func() {
				previousClusterConfig := &cnao.NetworkAddonsConfigSpec{
					KubeMacPool: &cnao.KubeMacPool{RangeStart: "02:00:00:00:00:00", RangeEnd: "0A:FF:FF:FF:FF:FF"}}
				currentClusterConfig := &cnao.NetworkAddonsConfigSpec{}

				errorList := changeSafeKubeMacPool(previousClusterConfig, currentClusterConfig)
				Expect(errorList).To(BeEmpty())
			})
		})
	})
})
