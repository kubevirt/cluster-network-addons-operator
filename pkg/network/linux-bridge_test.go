package network

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	opv1alpha1 "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/v1alpha1"
)

var _ = Describe("Testing linux-bridge", func() {
	Describe("changeSafeLinuxBridge", func() {
		Context("when it is kept disabled", func() {
			prev := &opv1alpha1.NetworkAddonsConfigSpec{}
			new := &opv1alpha1.NetworkAddonsConfigSpec{}
			It("should pass", func() {
				errorList := changeSafeLinuxBridge(prev, new)
				Expect(errorList).To(BeEmpty())
			})
		})

		Context("when there is no previous value", func() {
			prev := &opv1alpha1.NetworkAddonsConfigSpec{}
			new := &opv1alpha1.NetworkAddonsConfigSpec{LinuxBridge: &opv1alpha1.LinuxBridge{}}
			It("should accept any configuration", func() {
				errorList := changeSafeLinuxBridge(prev, new)
				Expect(errorList).To(BeEmpty())
			})
		})

		Context("when the previous and new configuration match", func() {
			prev := &opv1alpha1.NetworkAddonsConfigSpec{LinuxBridge: &opv1alpha1.LinuxBridge{}}
			new := &opv1alpha1.NetworkAddonsConfigSpec{LinuxBridge: &opv1alpha1.LinuxBridge{}}
			It("should accept the configuration", func() {
				errorList := changeSafeLinuxBridge(prev, new)
				Expect(errorList).To(BeEmpty())
			})
		})

		Context("when there is previous value, but the new one is empty (removing component)", func() {
			prev := &opv1alpha1.NetworkAddonsConfigSpec{LinuxBridge: &opv1alpha1.LinuxBridge{}}
			new := &opv1alpha1.NetworkAddonsConfigSpec{}
			It("should fail", func() {
				errorList := changeSafeLinuxBridge(prev, new)
				Expect(len(errorList)).To(Equal(1), "validation of safe change failed due to an unexpected error: %v", errorList)
				Expect(errorList[0].Error()).To(Equal("cannot modify Linux Bridge configuration once it is deployed"))
			})
		})
	})
})
