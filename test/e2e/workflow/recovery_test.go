package test

import (
	"time"

	. "github.com/onsi/ginkgo"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
	. "github.com/kubevirt/cluster-network-addons-operator/test/check"
	. "github.com/kubevirt/cluster-network-addons-operator/test/operations"
)

//2297
var _ = Describe("NetworkAddonsConfig", func() {
	Context("when invalid config is applied", func() {
		BeforeEach(func() {
			configSpec := cnao.NetworkAddonsConfigSpec{
				KubeMacPool: &cnao.KubeMacPool{
					RangeStart: "this:aint:right",
				},
			}
			CreateConfig(configSpec)
			CheckConfigCondition(ConditionDegraded, ConditionTrue, 5*time.Second, CheckDoNotRepeat)
			CheckFailedEvent("FailedToValidate")
		})

		Context("and it is updated with a valid config", func() {
			BeforeEach(func() {
				configSpec := cnao.NetworkAddonsConfigSpec{}
				UpdateConfig(configSpec)
			})

			It("should turn from Failing to Available", func() {
				CheckConfigCondition(ConditionAvailable, ConditionTrue, 5*time.Second, CheckDoNotRepeat)
				CheckConfigCondition(ConditionDegraded, ConditionFalse, CheckImmediately, CheckDoNotRepeat)
				CheckAvailableEvent()
			})
		})
	})
})
