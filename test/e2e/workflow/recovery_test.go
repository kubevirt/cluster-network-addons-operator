package test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
	. "github.com/kubevirt/cluster-network-addons-operator/test/check"
	. "github.com/kubevirt/cluster-network-addons-operator/test/operations"
)

//2297
var _ = Describe("NetworkAddonsConfig", func() {
	gvk := GetCnaoV1GroupVersionKind()
	Context("when invalid config is applied", func() {
		BeforeEach(func() {
			configSpec := cnao.NetworkAddonsConfigSpec{
				KubeMacPool: &cnao.KubeMacPool{
					RangeStart: "this:aint:right",
				},
			}
			CreateConfig(gvk, configSpec)
			CheckConfigCondition(gvk, ConditionDegraded, ConditionTrue, 5*time.Second, CheckDoNotRepeat)
			CheckFailedEvent(gvk, "FailedToValidate")
		})

		Context("and it is updated with a valid config", func() {
			BeforeEach(func() {
				configSpec := cnao.NetworkAddonsConfigSpec{}
				UpdateConfig(gvk, configSpec)
			})

			It("should turn from Failing to Available", func() {
				CheckAvailableEvent(gvk)
				CheckConfigCondition(gvk, ConditionAvailable, ConditionTrue, 5*time.Second, CheckDoNotRepeat)
				CheckConfigCondition(gvk, ConditionDegraded, ConditionFalse, CheckImmediately, CheckDoNotRepeat)
			})
		})
	})
})
