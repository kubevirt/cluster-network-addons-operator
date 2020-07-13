package k8s_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/kubevirt/cluster-network-addons-operator/pkg/util/k8s"
)

var _ = Describe("StringToLabel", func() {
	Context("when given a very long string", func() {
		s := "Avalanches of water in the midst of a calm/And the distances cataracting toward the abyss!"

		It("should trim the text to supported 63 characters", func() {
			label := k8s.StringToLabel(s)
			Expect(label).To(HaveLen(63), "The string should be trimmed to 63 characters")
		})
	})

	Context("when given a string with special characters", func() {
		s := "Avalanches&of:water"

		It("should replace special characters with underscore", func() {
			label := k8s.StringToLabel(s)
			Expect(label).To(Equal("Avalanches_of_water"), "The string should have special characters replaced")
		})
	})

	Context("when given a short string without special characters", func() {
		s := "Avalanches_of-water."

		It("should leave it intact", func() {
			label := k8s.StringToLabel(s)
			Expect(label).To(Equal(s), "The string should be left intact")
		})
	})

	Context("when given an empty string", func() {
		s := ""

		It("should leave it intact", func() {
			label := k8s.StringToLabel(s)
			Expect(label).To(Equal(s), "The string should be left intact")
		})
	})
})
