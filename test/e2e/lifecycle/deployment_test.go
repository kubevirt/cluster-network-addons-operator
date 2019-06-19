package test

import (
	"time"

	. "github.com/onsi/ginkgo"

	. "github.com/kubevirt/cluster-network-addons-operator/test/check"
	. "github.com/kubevirt/cluster-network-addons-operator/test/operations"
	. "github.com/kubevirt/cluster-network-addons-operator/test/releases"
)

var _ = Context("Cluster Network Addons Operator", func() {
	Context("when installed from master release", func() {
		masterRelease := LatestRelease()
		BeforeEach(func() {
			InstallRelease(masterRelease)
		})

		It("successfully turns ready", func() {
			CheckOperatorIsReady(10 * time.Minute)
		})

		Context("and when NodeNetworkConfig with supported spec is created", func() {
			BeforeEach(func() {
				CreateConfig(masterRelease.SupportedSpec)
			})

			It("reaches Available condition with all containers using expected images", func() {
				CheckConfigCondition(ConditionAvailable, ConditionTrue, 10*time.Minute, CheckDoNotRepeat)
				CheckReleaseUsesExpectedContainerImages(masterRelease)
			})

			It("stays in Available condition while the operator is being removed and redeployed", func() {
				configIsAvailable := func() {
					CheckConfigCondition(ConditionAvailable, ConditionTrue, CheckImmediately, CheckDoNotRepeat)
				}

				reinstallingOperator := func() {
					// Give validator some time to verify original state
					time.Sleep(3 * time.Second)

					UninstallRelease(masterRelease)
					InstallRelease(masterRelease)
					CheckOperatorIsReady(10 * time.Minute)

					// Give validator some time to verify conditions while the new installation is operating
					time.Sleep(3 * time.Second)
				}

				// Wait until the configuration reaches Available state
				CheckConfigCondition(ConditionAvailable, ConditionTrue, 10*time.Minute, CheckDoNotRepeat)

				// KubeMacPool is known to restart shortly after first started, try to skip this initial restart
				// https://github.com/kubevirt/cluster-network-addons-operator/issues/141
				// TODO: This should be dropped once KubeMacPool fixes the issue
				By("Ignoring initial KubeMacPool restart")
				time.Sleep(10 * time.Second)

				// Make sure that configuration stays available during operator's reinstallation
				KeepCheckingWhile(configIsAvailable, reinstallingOperator)
			})
		})
	})
})
