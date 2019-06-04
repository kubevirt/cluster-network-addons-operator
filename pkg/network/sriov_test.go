package network

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	osv1 "github.com/openshift/api/operator/v1"

	opv1alpha1 "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/v1alpha1"
)

var _ = Describe("Testing sriov", func() {
	Describe("validateSriov", func() {
		_validateSriov := func(
			sriovRequested bool,
			openshiftNetworkOperatorRunning bool,
			openshiftSriovEnabled bool,
		) []error {
			conf := &opv1alpha1.NetworkAddonsConfigSpec{}
			if sriovRequested {
				conf.Sriov = &opv1alpha1.Sriov{}
			}

			var openshiftNetworkConfig *osv1.Network
			if openshiftNetworkOperatorRunning {
				openshiftNetworkConfig = &osv1.Network{}
				if openshiftSriovEnabled {
					openshiftNetworkConfig.Spec.AdditionalNetworks = append(openshiftNetworkConfig.Spec.AdditionalNetworks, osv1.AdditionalNetworkDefinition{
						Type:         osv1.NetworkTypeRaw,
						Name:         "sriov-network-name",
						RawCNIConfig: "{\"type\": \"sriov\"}",
					})
				}
			}

			return validateSriov(conf, openshiftNetworkConfig)
		}

		Context("when configuration is empty", func() {
			sriovRequested := false
			openshiftNetworkOperatorRunning := false
			openshiftSriovEnabled := false
			It("should pass", func() {
				errorList := _validateSriov(sriovRequested, openshiftNetworkOperatorRunning, openshiftSriovEnabled)
				Expect(errorList).To(BeEmpty())
			})
		})

		Context("when openshift network operator is running", func() {
			openshiftNetworkOperatorRunning := true
			Context("and has sriov enabled", func() {
				openshiftSriovEnabled := true
				Context("and configuration is empty", func() {
					sriovRequested := false
					It("should pass", func() {
						errorList := _validateSriov(sriovRequested, openshiftNetworkOperatorRunning, openshiftSriovEnabled)
						Expect(errorList).To(BeEmpty())
					})
				})
				Context("and configuration requests sriov", func() {
					sriovRequested := true
					It("should fail with an error message", func() {
						errorList := _validateSriov(sriovRequested, openshiftNetworkOperatorRunning, openshiftSriovEnabled)
						Expect(len(errorList)).To(Equal(1), "validation failed due to an unexpected error: %v", errorList)
						Expect(errorList[0].Error()).To(Equal("SR-IOV has been requested, but it's not compatible with OpenShift Cluster Network Operator SR-IOV support"))
					})
				})
			})

			Context("and has sriov disabled", func() {
				openshiftSriovEnabled := false
				Context("and configuration is empty", func() {
					sriovRequested := false
					It("should pass", func() {
						errorList := _validateSriov(sriovRequested, openshiftNetworkOperatorRunning, openshiftSriovEnabled)
						Expect(errorList).To(BeEmpty())
					})
				})
				Context("and configuration requests sriov", func() {
					sriovRequested := true
					It("should pass", func() {
						errorList := _validateSriov(sriovRequested, openshiftNetworkOperatorRunning, openshiftSriovEnabled)
						Expect(errorList).To(BeEmpty())
					})
				})
			})
		})

		Context("when openshift network operator is not running", func() {
			openshiftNetworkOperatorRunning := false
			Context("and configuration is empty", func() {
				sriovRequested := false
				It("should pass", func() {
					errorList := _validateSriov(sriovRequested, openshiftNetworkOperatorRunning, false)
					Expect(errorList).To(BeEmpty())
				})
			})
			Context("and configuration requests sriov", func() {
				sriovRequested := true
				It("should pass", func() {
					errorList := _validateSriov(sriovRequested, openshiftNetworkOperatorRunning, false)
					Expect(errorList).To(BeEmpty())
				})
			})
		})
	})

	Describe("changeSafeSriov", func() {
		Context("when it is kept disabled", func() {
			prev := &opv1alpha1.NetworkAddonsConfigSpec{}
			new := &opv1alpha1.NetworkAddonsConfigSpec{}
			It("should pass", func() {
				errorList := changeSafeSriov(prev, new)
				Expect(errorList).To(BeEmpty())
			})
		})

		Context("when there is no previous value", func() {
			prev := &opv1alpha1.NetworkAddonsConfigSpec{}
			new := &opv1alpha1.NetworkAddonsConfigSpec{Sriov: &opv1alpha1.Sriov{}}
			It("should accept any configuration", func() {
				errorList := changeSafeSriov(prev, new)
				Expect(errorList).To(BeEmpty())
			})
		})

		Context("when the previous and new configuration matches", func() {
			prev := &opv1alpha1.NetworkAddonsConfigSpec{Sriov: &opv1alpha1.Sriov{}}
			new := &opv1alpha1.NetworkAddonsConfigSpec{Sriov: &opv1alpha1.Sriov{}}
			It("should accept the configuration", func() {
				errorList := changeSafeSriov(prev, new)
				Expect(errorList).To(BeEmpty())
			})
		})

		Context("when there is previous value, but the new one is empty (removing component)", func() {
			prev := &opv1alpha1.NetworkAddonsConfigSpec{Sriov: &opv1alpha1.Sriov{}}
			new := &opv1alpha1.NetworkAddonsConfigSpec{}
			It("should fail", func() {
				errorList := changeSafeSriov(prev, new)
				Expect(len(errorList)).To(Equal(1), "validation of safe change failed due to an unexpected error: %v", errorList)
				Expect(errorList[0].Error()).To(Equal("cannot modify Sriov configuration once it is deployed"))
			})
		})
	})

	Describe("getRootDevicesConfigString", func() {
		Context("when given a list of addresses separated by comma", func() {
			It("should return the same list, where each address is enclosed in escaped quotes", func() {
				escapedList := getRootDevicesConfigString("0000:03:02.1,0000:03:04.3")
				Expect(escapedList).To(Equal("\"0000:03:02.1\",\"0000:03:04.3\""))
			})
		})
	})
})
