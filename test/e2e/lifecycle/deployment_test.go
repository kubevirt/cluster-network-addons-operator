package test

import (
	"time"

	. "github.com/onsi/ginkgo"

	. "github.com/kubevirt/cluster-network-addons-operator/test/check"
	. "github.com/kubevirt/cluster-network-addons-operator/test/operations"
	. "github.com/kubevirt/cluster-network-addons-operator/test/releases"
)

var _ = Context("Cluster Network Addons Operator", func() {
	Context("when installed from main release", func() {
		mainRelease := LatestRelease()
		BeforeEach(func() {
			InstallRelease(mainRelease)
			CheckOperatorIsReady(15 * time.Minute)
		})
		Context("and when NodeNetworkConfig with supported spec is created", func() {
			gvk := GetCnaoV1GroupVersionKind()
			BeforeEach(func() {
				CreateConfig(gvk, mainRelease.SupportedSpec)
			})

			It("reaches Available condition with all containers using expected images", func() {
				CheckConfigCondition(gvk, ConditionAvailable, ConditionTrue, 15*time.Minute, CheckDoNotRepeat)
				CheckReleaseUsesExpectedContainerImages(gvk, mainRelease)
			})

			It("stays in Available condition while the operator is being removed and redeployed", func() {
				configIsAvailable := func() {
					CheckConfigCondition(gvk, ConditionAvailable, ConditionTrue, CheckImmediately, CheckDoNotRepeat)
				}

				reinstallingOperator := func() {
					// Give validator some time to verify original state
					time.Sleep(3 * time.Second)

					UninstallOperator(mainRelease)
					InstallOperator(mainRelease)
					CheckOperatorIsReady(15 * time.Minute)

					// Give validator some time to verify conditions while the new installation is operating
					time.Sleep(3 * time.Second)
				}

				// Wait until the configuration reaches Available state
				CheckConfigCondition(gvk, ConditionAvailable, ConditionTrue, 15*time.Minute, CheckDoNotRepeat)

				// Make sure that configuration stays available during operator's reinstallation
				KeepCheckingWhile(configIsAvailable, reinstallingOperator)
			})
		})
	})
})
