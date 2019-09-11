package test

import (
	"time"

	. "github.com/onsi/ginkgo"

	opv1alpha1 "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/v1alpha1"

	. "github.com/kubevirt/cluster-network-addons-operator/test/check"
	. "github.com/kubevirt/cluster-network-addons-operator/test/operations"
)

//2297
var _ = Describe("NetworkAddonsConfig", func() {
	Context("when invalid config is applied", func() {
		BeforeEach(func() {
			configSpec := opv1alpha1.NetworkAddonsConfigSpec{
				KubeMacPool: &opv1alpha1.KubeMacPool{
					RangeStart: "this:aint:right",
				},
			}
			CreateConfig(configSpec)
			CheckConfigCondition(ConditionDegraded, ConditionTrue, 5*time.Second, CheckDoNotRepeat)
		})

		Context("and it is updated with a valid config", func() {
			BeforeEach(func() {
				configSpec := opv1alpha1.NetworkAddonsConfigSpec{}
				UpdateConfig(configSpec)
			})

			It("should turn from Failing to Available", func() {
				CheckConfigCondition(ConditionAvailable, ConditionTrue, 5*time.Second, CheckDoNotRepeat)
				CheckConfigCondition(ConditionDegraded, ConditionFalse, CheckImmediately, CheckDoNotRepeat)
			})
		})
	})
})
