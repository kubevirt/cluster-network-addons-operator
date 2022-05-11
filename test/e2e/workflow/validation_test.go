package test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	v1 "k8s.io/api/core/v1"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
	. "github.com/kubevirt/cluster-network-addons-operator/test/check"
	. "github.com/kubevirt/cluster-network-addons-operator/test/operations"
)

var _ = Describe("NetworkAddonsConfig", func() {
	gvk := GetCnaoV1GroupVersionKind()
	Context("when there is no running config", func() {
		Context("and an invalid config is created", func() {
			BeforeEach(func() {
				configSpec := cnao.NetworkAddonsConfigSpec{
					ImagePullPolicy: v1.PullAlways,
					KubeMacPool: &cnao.KubeMacPool{
						RangeStart: "this:aint:right",
					},
					LinuxBridge: &cnao.LinuxBridge{},
					Multus:      &cnao.Multus{},
					Ovs:         &cnao.Ovs{},
					MacvtapCni:  &cnao.MacvtapCni{},
				}
				CreateConfig(gvk, configSpec)
			})

			It("should report Failing condition and Available must be set to False", func() {
				CheckConfigCondition(gvk, ConditionDegraded, ConditionTrue, time.Minute, CheckDoNotRepeat)
				CheckConfigCondition(gvk, ConditionAvailable, ConditionFalse, CheckImmediately, CheckDoNotRepeat)
			})
		})
	})

	Context("when a valid config is deployed", func() {
		BeforeEach(func() {
			configSpec := cnao.NetworkAddonsConfigSpec{
				LinuxBridge: &cnao.LinuxBridge{},
			}
			CreateConfig(gvk, configSpec)
			CheckConfigCondition(gvk, ConditionAvailable, ConditionTrue, 2*time.Minute, CheckDoNotRepeat)
		})

		Context("and a component which does support removal is removed from the Spec", func() {
			var resourceVersion string
			var updatedConfigSpec cnao.NetworkAddonsConfigSpec

			BeforeEach(func() {
				updatedConfigSpec = cnao.NetworkAddonsConfigSpec{
					LinuxBridge: &cnao.LinuxBridge{},
				}
				UpdateConfig(gvk, updatedConfigSpec)
				resourceVersion = GetConfig(gvk).GetResourceVersion()
			})

			It("should remain at Available condition", func() {
				By("Checking these were no Warning events during the component's removal")
				CheckNoWarningEvents(gvk, resourceVersion)

				By("Checking that spec has been deployed")
				currentConfigSpec := ConvertToConfigV1(GetConfig(gvk)).Spec
				Expect(currentConfigSpec).To(Equal(updatedConfigSpec))

				By("Checking that Available status turn to True")
				CheckConfigCondition(gvk, ConditionAvailable, ConditionTrue, CheckImmediately, time.Minute)
				By("Checking that Degraded status turn to False")
				CheckConfigCondition(gvk, ConditionDegraded, ConditionFalse, CheckImmediately, time.Minute)
			})
		})
	})
})
