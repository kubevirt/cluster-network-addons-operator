package test

import (
	"time"

	. "github.com/onsi/ginkgo"

	opv1alpha1 "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/v1alpha1"

	. "github.com/kubevirt/cluster-network-addons-operator/test/check"
	. "github.com/kubevirt/cluster-network-addons-operator/test/operations"
)

var _ = Describe("NetworkAddonsConfig", func() {
	Context("when a config is created", func() {
		BeforeEach(func() {
			configSpec := opv1alpha1.NetworkAddonsConfigSpec{
				LinuxBridge: &opv1alpha1.LinuxBridge{},
			}
			CreateConfig(configSpec)
		})

		It("should set targetVersion and operatorVersion immediately after it turns Progressing", func() {
			CheckConfigCondition(ConditionProgressing, ConditionTrue, time.Minute, CheckDoNotRepeat)
			CheckConfigVersions(operatorVersion, "", operatorVersion, CheckImmediately, CheckDoNotRepeat)
		})

		It("should set observedVersion once turns it Available", func() {
			CheckConfigCondition(ConditionAvailable, ConditionTrue, 10*time.Minute, CheckDoNotRepeat)
			CheckConfigVersions(operatorVersion, operatorVersion, operatorVersion, CheckImmediately, CheckDoNotRepeat)
		})
	})

	Context("when there is an existing config", func() {
		BeforeEach(func() {
			configSpec := opv1alpha1.NetworkAddonsConfigSpec{
				Multus: &opv1alpha1.Multus{},
			}
			CreateConfig(configSpec)
			CheckConfigCondition(ConditionAvailable, ConditionTrue, 10*time.Minute, CheckDoNotRepeat)
		})

		It("should keep reported versions while being changed", func() {
			versionRemainsTheSame := func() {
				CheckConfigVersions(operatorVersion, operatorVersion, operatorVersion, CheckImmediately, CheckDoNotRepeat)
			}

			updatingConfig := func() {
				configSpec := opv1alpha1.NetworkAddonsConfigSpec{
					Multus:      &opv1alpha1.Multus{},
					LinuxBridge: &opv1alpha1.LinuxBridge{},
				}

				// Give validator some time to verify original state
				time.Sleep(3 * time.Second)

				UpdateConfig(configSpec)
				CheckConfigCondition(ConditionAvailable, ConditionTrue, 10*time.Minute, CheckDoNotRepeat)

				// Give validator some time to verify versions while we stay in updated config
				time.Sleep(3 * time.Second)
			}

			KeepCheckingWhile(versionRemainsTheSame, updatingConfig)
		})
	})
})
