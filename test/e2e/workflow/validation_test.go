package test

import (
	"time"

	. "github.com/onsi/ginkgo"
	v1 "k8s.io/api/core/v1"

	opv1alpha1 "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/v1alpha1"

	. "github.com/kubevirt/cluster-network-addons-operator/test/check"
	. "github.com/kubevirt/cluster-network-addons-operator/test/operations"
)

var _ = Describe("NetworkAddonsConfig", func() {
	Context("when there is no running config", func() {
		Context("and an invalid config is created", func() {
			BeforeEach(func() {
				configSpec := opv1alpha1.NetworkAddonsConfigSpec{
					ImagePullPolicy: v1.PullAlways,
					KubeMacPool: &opv1alpha1.KubeMacPool{
						RangeStart: "this:aint:right",
					},
					LinuxBridge: &opv1alpha1.LinuxBridge{},
					Multus:      &opv1alpha1.Multus{},
					Ovs:         &opv1alpha1.Ovs{},
					NMState:     &opv1alpha1.NMState{},
				}
				CreateConfig(configSpec)
			})

			It("should report Failing condition and Available must be set to False", func() {
				CheckConfigCondition(ConditionDegraded, ConditionTrue, time.Minute, CheckDoNotRepeat)
				CheckConfigCondition(ConditionAvailable, ConditionFalse, CheckImmediately, CheckDoNotRepeat)
			})
		})
	})

	Context("when a valid config is deployed", func() {
		BeforeEach(func() {
			configSpec := opv1alpha1.NetworkAddonsConfigSpec{
				LinuxBridge: &opv1alpha1.LinuxBridge{},
			}
			CreateConfig(configSpec)
			CheckConfigCondition(ConditionAvailable, ConditionTrue, time.Minute, CheckDoNotRepeat)
		})

		Context("and a component which does not support removal is removed from the Spec", func() {
			BeforeEach(func() {
				configSpec := opv1alpha1.NetworkAddonsConfigSpec{}
				UpdateConfig(configSpec)
			})

			It("should report Failing condition and remain stop being Available", func() {
				CheckConfigCondition(ConditionDegraded, ConditionTrue, time.Minute, CheckDoNotRepeat)
				CheckConfigCondition(ConditionAvailable, ConditionFalse, CheckImmediately, CheckDoNotRepeat)
			})
		})
	})
})
