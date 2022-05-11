package test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
	. "github.com/kubevirt/cluster-network-addons-operator/test/check"
	. "github.com/kubevirt/cluster-network-addons-operator/test/operations"
)

var _ = Describe("NetworkAddonsConfig", func() {
	gvk := GetCnaoV1GroupVersionKind()
	Context("when a config is created", func() {
		BeforeEach(func() {
			configSpec := cnao.NetworkAddonsConfigSpec{
				LinuxBridge: &cnao.LinuxBridge{},
			}
			CreateConfig(gvk, configSpec)
		})

		It("should set targetVersion and operatorVersion immediately after it turns Progressing", func() {
			CheckConfigCondition(gvk, ConditionProgressing, ConditionTrue, time.Minute, CheckDoNotRepeat)
			CheckConfigVersions(gvk, operatorVersion, CheckIgnoreVersion, operatorVersion, CheckImmediately, CheckDoNotRepeat)
		})

		It("should set observedVersion once turns it Available", func() {
			CheckConfigCondition(gvk, ConditionAvailable, ConditionTrue, 15*time.Minute, CheckDoNotRepeat)
			CheckConfigVersions(gvk, operatorVersion, operatorVersion, operatorVersion, CheckImmediately, CheckDoNotRepeat)
		})
	})

	Context("when there is an existing config", func() {
		BeforeEach(func() {
			configSpec := cnao.NetworkAddonsConfigSpec{
				Multus: &cnao.Multus{},
			}
			CreateConfig(gvk, configSpec)
			CheckConfigCondition(gvk, ConditionAvailable, ConditionTrue, 15*time.Minute, CheckDoNotRepeat)
		})

		It("should keep reported versions while being changed", func() {
			versionRemainsTheSame := func() {
				CheckConfigVersions(gvk, operatorVersion, operatorVersion, operatorVersion, CheckImmediately, CheckDoNotRepeat)
			}

			updatingConfig := func() {
				configSpec := cnao.NetworkAddonsConfigSpec{
					Multus:      &cnao.Multus{},
					LinuxBridge: &cnao.LinuxBridge{},
				}

				// Give validator some time to verify original state
				time.Sleep(3 * time.Second)

				UpdateConfig(gvk, configSpec)
				CheckConfigCondition(gvk, ConditionAvailable, ConditionTrue, 15*time.Minute, CheckDoNotRepeat)

				// Give validator some time to verify versions while we stay in updated config
				time.Sleep(3 * time.Second)
			}

			KeepCheckingWhile(versionRemainsTheSame, updatingConfig)
		})
	})
})
