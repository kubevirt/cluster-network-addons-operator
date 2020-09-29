package test

import (
	"time"

	. "github.com/onsi/ginkgo"
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
					NMState:     &cnao.NMState{},
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
				NMState:     &cnao.NMState{},
			}
			CreateConfig(gvk, configSpec)
			CheckConfigCondition(gvk, ConditionAvailable, ConditionTrue, 2*time.Minute, CheckDoNotRepeat)
		})

		Context("and a component which does support removal is removed from the Spec", func() {
			BeforeEach(func() {
				configSpec := cnao.NetworkAddonsConfigSpec{
					LinuxBridge: &cnao.LinuxBridge{},
				}
				UpdateConfig(gvk, configSpec)
			})

			It("should remain at Available condition", func() {
				components := []Component{NMStateComponent}
				CheckComponentsRemoval(components)

				By("Checking that Available status turn to True after config update")
				CheckConfigCondition(gvk, ConditionAvailable, ConditionTrue, time.Minute, time.Minute)
				By("Checking that Degraded status turn to False after config update")
				CheckConfigCondition(gvk, ConditionDegraded, ConditionFalse, CheckImmediately, time.Minute)
			})
		})
	})
})
