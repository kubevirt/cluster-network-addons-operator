package network

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	opv1alpha1 "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/v1alpha1"
)

var _ = Describe("kubemacpool_test", func() {
	It("should NOT return an error because kube mac pool can be nil", func() {
		clusterConfig := &opv1alpha1.NetworkAddonsConfigSpec{KubeMacPool: &opv1alpha1.KubeMacPool{}}
		errorList := validateKubeMacPool(clusterConfig)
		Expect(errorList).To(BeEmpty())
	})

	It("should NOT return an error because both or none of the KubeMacPool can be configured", func() {
		clusterConfig := &opv1alpha1.NetworkAddonsConfigSpec{KubeMacPool: &opv1alpha1.KubeMacPool{RangeStart: "", RangeEnd: ""}}
		errorList := validateKubeMacPool(clusterConfig)
		Expect(errorList).To(BeEmpty())
	})

	It("should return an error because only the start range is configured", func() {
		clusterConfig := &opv1alpha1.NetworkAddonsConfigSpec{
			KubeMacPool: &opv1alpha1.KubeMacPool{RangeStart: "02:00:00:00:00:00", RangeEnd: ""}}
		errorList := validateKubeMacPool(clusterConfig)
		Expect(len(errorList)).To(Equal(1), "validation failed due to an unexpected error: %s", errorList[0].Error())
		Expect(errorList[0].Error()).To(Equal("both or none of the KubeMacPool ranges needs to be configured"))
	})

	It("should return an error because only the end range is configured", func() {
		clusterConfig := &opv1alpha1.NetworkAddonsConfigSpec{
			KubeMacPool: &opv1alpha1.KubeMacPool{RangeStart: "", RangeEnd: "02:00:00:FF:FF:FF"}}
		errorList := validateKubeMacPool(clusterConfig)
		Expect(len(errorList)).To(Equal(1), "validation failed due to an unexpected error: %s", errorList[0].Error())
		Expect(errorList[0].Error()).To(Equal("both or none of the KubeMacPool ranges needs to be configured"))
	})

	It("should return an error because the start range mac address contains 7 octets and not 6", func() {
		clusterConfig := &opv1alpha1.NetworkAddonsConfigSpec{
			KubeMacPool: &opv1alpha1.KubeMacPool{RangeStart: "00:00:00:00:00:00:00", RangeEnd: "02:FF:FF:FF:FF:FF"}}
		errorList := validateKubeMacPool(clusterConfig)
		Expect(len(errorList)).To(Equal(1), "validation failed due to an unexpected error: %s", errorList[0].Error())
		Expect(errorList[0].Error()).To(Equal("failed to parse rangeStart because the mac address is invalid"))

	})

	It("should return an error because the end range mac contains 7 octets and not 6", func() {
		clusterConfig := &opv1alpha1.NetworkAddonsConfigSpec{
			KubeMacPool: &opv1alpha1.KubeMacPool{RangeStart: "02:00:00:00:00:00", RangeEnd: "FF:FF:FF:FF:FF:FF:FF"}}
		errorList := validateKubeMacPool(clusterConfig)
		Expect(len(errorList)).To(Equal(1), "validation failed due to an unexpected error: %s", errorList[0].Error())
		Expect(errorList[0].Error()).To(Equal("failed to parse rangeEnd because the mac address is invalid"))
	})

	It("should return an error because Range end is lesser than its Start ", func() {
		clusterConfig := &opv1alpha1.NetworkAddonsConfigSpec{
			KubeMacPool: &opv1alpha1.KubeMacPool{RangeStart: "02:00:00:FF:00:00", RangeEnd: "02:00:00:00:00:00"}}
		errorList := validateKubeMacPool(clusterConfig)
		Expect(len(errorList)).To(Equal(1), "validation failed due to an unexpected error: %s", errorList[0].Error())
		Expect(errorList[0].Error()).To(Equal("failed to set mac address range. Range end is lesser than its Start"), "validation failed: %s", errorList[0])
	})

	It("should return an error because the multicast-bit is on in RangeStart", func() {
		clusterConfig := &opv1alpha1.NetworkAddonsConfigSpec{
			KubeMacPool: &opv1alpha1.KubeMacPool{RangeStart: "01:00:00:00:00:00", RangeEnd: "06:FF:FF:FF:FF:FF"}}
		errorList := validateKubeMacPool(clusterConfig)
		Expect(len(errorList)).To(Equal(1), "validation failed due to an unexpected error: %s", errorList[0].Error())
		Expect(errorList[0].Error()).To(Equal("failed to set RangeStart. invalid mac address. Multicast addressing is not supported. The first octet is 0X1"))
	})

	It("should return an error because the multicast-bit is on in RangeEnd", func() {
		clusterConfig := &opv1alpha1.NetworkAddonsConfigSpec{
			KubeMacPool: &opv1alpha1.KubeMacPool{RangeStart: "02:00:00:00:00:00", RangeEnd: "03:FF:FF:FF:FF:FF"}}
		errorList := validateKubeMacPool(clusterConfig)
		Expect(len(errorList)).To(Equal(1), "validation failed due to an unexpected error: %s", errorList[0].Error())
		Expect(errorList[0].Error()).To(Equal("failed to set RangeEnd. invalid mac address. Multicast addressing is not supported. The first octet is 0X3"))
	})

	It("should return an error because the locally-administred bit is off in RangeStart", func() {
		clusterConfig := &opv1alpha1.NetworkAddonsConfigSpec{
			KubeMacPool: &opv1alpha1.KubeMacPool{RangeStart: "04:00:00:00:00:00", RangeEnd: "02:FF:FF:FF:FF:FF"}}
		errorList := validateKubeMacPool(clusterConfig)
		Expect(len(errorList)).To(Equal(1), "validation failed due to an unexpected error: %s", errorList[0].Error())
		Expect(errorList[0].Error()).To(Equal("failed to set RangeStart. invalid mac address. Universally administered addresses are not supported. The first octet is 0X4"))
	})

	It("should return an error because the locally-administred bit is off in RangeEnd", func() {
		clusterConfig := &opv1alpha1.NetworkAddonsConfigSpec{
			KubeMacPool: &opv1alpha1.KubeMacPool{RangeStart: "02:00:00:00:00:00", RangeEnd: "08:FF:FF:FF:FF:FF"}}
		errorList := validateKubeMacPool(clusterConfig)
		Expect(len(errorList)).To(Equal(1), "validation failed due to an unexpected error: %s", errorList[0].Error())
		Expect(errorList[0].Error()).To(Equal("failed to set RangeEnd. invalid mac address. Universally administered addresses are not supported. The first octet is 0X8"))
	})

	It("should NOT return an error because the mac address is valid", func() {
		clusterConfig := &opv1alpha1.NetworkAddonsConfigSpec{
			KubeMacPool: &opv1alpha1.KubeMacPool{RangeStart: "02:00:00:00:00:00", RangeEnd: "0A:FF:FF:FF:FF:FF"}}
		errorList := validateKubeMacPool(clusterConfig)
		Expect(errorList).To(BeEmpty())
	})
})
