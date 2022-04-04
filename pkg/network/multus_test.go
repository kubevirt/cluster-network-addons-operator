package network

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	osv1 "github.com/openshift/api/operator/v1"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
)

var _ = Describe("Testing multus", func() {
	Describe("validateMultus", func() {
		_validateMultus := func(
			multusRequested bool,
			openshiftNetworkOperatorRunning bool,
			openshiftNetworkOperatorDisableMultiNetwork bool,
		) []error {
			conf := &cnao.NetworkAddonsConfigSpec{}
			if multusRequested {
				conf.Multus = &cnao.Multus{}
			}

			var openshiftNetworkConfig *osv1.Network
			if openshiftNetworkOperatorRunning {
				openshiftNetworkConfig = &osv1.Network{}
				if openshiftNetworkOperatorDisableMultiNetwork {
					newTrue := func() *bool {
						val := true
						return &val
					}
					openshiftNetworkConfig.Spec.DisableMultiNetwork = newTrue()
				}
			}

			return validateMultus(conf, openshiftNetworkConfig)
		}

		Context("when configuration is empty", func() {
			multusRequested := false
			openshiftNetworkOperatorRunning := false
			openshiftNetworkOperatorDisableMultiNetwork := false
			It("should pass", func() {
				errorList := _validateMultus(multusRequested, openshiftNetworkOperatorRunning, openshiftNetworkOperatorDisableMultiNetwork)
				Expect(errorList).To(BeEmpty())
			})
		})
		//2299
		Context("when openshift network operator is running", func() {
			openshiftNetworkOperatorRunning := true
			Context("and has multiNetwork disabled", func() {
				openshiftNetworkOperatorDisableMultiNetwork := true
				Context("and configuration is empty", func() {
					multusRequested := false
					It("should pass", func() {
						errorList := _validateMultus(multusRequested, openshiftNetworkOperatorRunning, openshiftNetworkOperatorDisableMultiNetwork)
						Expect(errorList).To(BeEmpty())
					})
				})
				Context("and configuration requests multus", func() {
					multusRequested := true
					It("should fail with a nice error message", func() {
						errorList := _validateMultus(multusRequested, openshiftNetworkOperatorRunning, openshiftNetworkOperatorDisableMultiNetwork)
						Expect(len(errorList)).To(Equal(1), "validation failed due to an unexpected error: %v", errorList)
						Expect(errorList[0].Error()).To(Equal("multus has been requested, but is disabled on OpenShift Cluster Network Operator"))
					})
				})
			})

			Context("and has multiNetwork enabled", func() {
				openshiftNetworkOperatorDisableMultiNetwork := false
				Context("and configuration is empty", func() {
					multusRequested := false
					It("should pass", func() {
						errorList := _validateMultus(multusRequested, openshiftNetworkOperatorRunning, openshiftNetworkOperatorDisableMultiNetwork)
						Expect(errorList).To(BeEmpty())
					})
				})
				Context("and configuration requests multus", func() {
					multusRequested := true
					It("should pass", func() {
						errorList := _validateMultus(multusRequested, openshiftNetworkOperatorRunning, openshiftNetworkOperatorDisableMultiNetwork)
						Expect(errorList).To(BeEmpty())
					})
				})
			})
		})

		Context("when openshift network operator is not running", func() {
			openshiftNetworkOperatorRunning := false
			openshiftNetworkOperatorDisableMultiNetwork := false
			Context("and configuration is empty", func() {
				multusRequested := false
				It("should pass", func() {
					errorList := _validateMultus(multusRequested, openshiftNetworkOperatorRunning, openshiftNetworkOperatorDisableMultiNetwork)
					Expect(errorList).To(BeEmpty())
				})
			})
			Context("and configuration requests multus", func() {
				multusRequested := true
				It("should pass", func() {
					errorList := _validateMultus(multusRequested, openshiftNetworkOperatorRunning, openshiftNetworkOperatorDisableMultiNetwork)
					Expect(errorList).To(BeEmpty())
				})
			})
		})
	})
})
