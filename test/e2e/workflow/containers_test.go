package test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
	. "github.com/kubevirt/cluster-network-addons-operator/test/check"
	. "github.com/kubevirt/cluster-network-addons-operator/test/operations"
)

var _ = Describe("NetworkAddonsConfig", func() {
	Context("when is the config already deployed", func() {
		configSpec := cnao.NetworkAddonsConfigSpec{
			LinuxBridge: &cnao.LinuxBridge{},
		}

		BeforeEach(func() {
			CreateConfig(configSpec)
			CheckConfigCondition(ConditionAvailable, ConditionTrue, 15*time.Minute, CheckDoNotRepeat)
		})

		It("should report non-empty list of deployed containers", func() {
			config := GetConfig()
			Expect(config.Status.Containers).NotTo(BeEmpty())
		})
	})
})
