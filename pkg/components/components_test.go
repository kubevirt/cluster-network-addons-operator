package components

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const imageName = "the-image-name"

var _ = Describe("Components", func() {
	DescribeTable("When RelatedImage is called", func(fullImageName, expectedShortName string) {
		ri := NewRelatedImage(fullImageName)
		Expect(ri.Ref).To(Equal(fullImageName))
		Expect(ri.Name).To(Equal(expectedShortName))
	},
		Entry("Should extract a short image name from image name",
			"imageRegistry/organization/the-image-name:1.2.3",
			"the-image-name",
		),
		Entry("Should extract a short image name from image digest",
			"imageRegistry/organization/the-image-name@sha256:76cc13fb4a60943dca6038619599b6a49fe451852aba23ad3046658429a9af30",
			"the-image-name",
		),
		Entry("Should copy the image name if it not match to the standard",
			"the-image-name",
			"the-image-name",
		),
		Entry("Should find the short name, when the host name is with a port number",
			"registry:5000/the-image-name@sha256:76cc13fb4a60943dca6038619599b6a49fe451852aba23ad3046658429a9af30",
			"the-image-name",
		),
		Entry("Should find the short name, when the image name is just organization/image",
			"organization/the-image-name",
			"the-image-name",
		),
	)

	Context("When RelatedImages constructor is called", func() {
		It("Should create an empty list when there are no parameters", func() {
			ris := NewRelatedImages()
			Expect(ris).To(BeEmpty())
		})

		It("Should create a list of related images if getting image names as parameters", func() {
			ris := NewRelatedImages(
				"nfvpe/multus@sha256:167722b954355361bd69829466f27172b871dbdbf86b85a95816362885dc0aba",
				"quay.io/kubevirt/cni-default-plugins@sha256:680ac8fd5eeab39c9a3c01479da344bdcaa43aa065d07ae00513b7bafa22fccf",
				"quay.io/kubevirt/ovs-cni-marker@sha256:0f08d6b1550a90c9f10221f2bb07709d1090e7c675ee1a711981bd429074d620",
			)
			Expect(ris).To(HaveLen(3))

			Expect(ris[0].Name).To(Equal("multus"))
			Expect(ris[0].Ref).To(Equal("nfvpe/multus@sha256:167722b954355361bd69829466f27172b871dbdbf86b85a95816362885dc0aba"))
			Expect(ris[1].Name).To(Equal("cni-default-plugins"))
			Expect(ris[1].Ref).To(Equal("quay.io/kubevirt/cni-default-plugins@sha256:680ac8fd5eeab39c9a3c01479da344bdcaa43aa065d07ae00513b7bafa22fccf"))
			Expect(ris[2].Name).To(Equal("ovs-cni-marker"))
			Expect(ris[2].Ref).To(Equal("quay.io/kubevirt/ovs-cni-marker@sha256:0f08d6b1550a90c9f10221f2bb07709d1090e7c675ee1a711981bd429074d620"))
		})
	})

	Context("When adding image to RelatedImages", func() {
		It("should add image to an empty list", func() {
			var ris RelatedImages
			Expect(ris).To(BeEmpty())

			ris.Add("nfvpe/multus@sha256:167722b954355361bd69829466f27172b871dbdbf86b85a95816362885dc0aba")
			Expect(ris).To(HaveLen(1))
			Expect(ris[0].Name).To(Equal("multus"))
			Expect(ris[0].Ref).To(Equal("nfvpe/multus@sha256:167722b954355361bd69829466f27172b871dbdbf86b85a95816362885dc0aba"))
		})

		It("should add image to an non-empty list", func() {
			ris := NewRelatedImages(
				"nfvpe/multus@sha256:167722b954355361bd69829466f27172b871dbdbf86b85a95816362885dc0aba",
				"quay.io/kubevirt/cni-default-plugins@sha256:680ac8fd5eeab39c9a3c01479da344bdcaa43aa065d07ae00513b7bafa22fccf",
			)
			Expect(ris).To(HaveLen(2))

			Expect(ris[0].Name).To(Equal("multus"))
			Expect(ris[0].Ref).To(Equal("nfvpe/multus@sha256:167722b954355361bd69829466f27172b871dbdbf86b85a95816362885dc0aba"))
			Expect(ris[1].Name).To(Equal("cni-default-plugins"))
			Expect(ris[1].Ref).To(Equal("quay.io/kubevirt/cni-default-plugins@sha256:680ac8fd5eeab39c9a3c01479da344bdcaa43aa065d07ae00513b7bafa22fccf"))

			By("adding a new image to the list")
			ris.Add("quay.io/kubevirt/ovs-cni-marker@sha256:0f08d6b1550a90c9f10221f2bb07709d1090e7c675ee1a711981bd429074d620")

			Expect(ris).To(HaveLen(3))

			Expect(ris[0].Name).To(Equal("multus"))
			Expect(ris[0].Ref).To(Equal("nfvpe/multus@sha256:167722b954355361bd69829466f27172b871dbdbf86b85a95816362885dc0aba"))
			Expect(ris[1].Name).To(Equal("cni-default-plugins"))
			Expect(ris[1].Ref).To(Equal("quay.io/kubevirt/cni-default-plugins@sha256:680ac8fd5eeab39c9a3c01479da344bdcaa43aa065d07ae00513b7bafa22fccf"))
			Expect(ris[2].Name).To(Equal("ovs-cni-marker"))
			Expect(ris[2].Ref).To(Equal("quay.io/kubevirt/ovs-cni-marker@sha256:0f08d6b1550a90c9f10221f2bb07709d1090e7c675ee1a711981bd429074d620"))
		})
	})

	Context("When calculating ciphers", func() {
		It("should not generate duplicates", func() {
			var ciphers = GetCrd().Spec.Versions[0].Schema.OpenAPIV3Schema.Properties["spec"].Properties["tlsSecurityProfile"].Properties["custom"].Properties["ciphers"].Items.Schema.Enum
			var stringCiphers = make([]string, len(ciphers))
			for i, c := range ciphers {
				stringCiphers[i] = string(c.Raw[:])
			}
			for i, vi := range stringCiphers {
				for j := i + 1; j < len(stringCiphers); j++ {
					Expect(vi).ToNot(Equal(stringCiphers[j]))
				}
			}
		})
	})

	Context("When Deploying CNAO operator", func() {
		It("should have TerminationMessagePolicy", func() {
			var ciphers = GetCrd().Spec.Versions[0].Schema.OpenAPIV3Schema.Properties["spec"].Properties["tlsSecurityProfile"].Properties["custom"].Properties["ciphers"].Items.Schema.Enum
			var stringCiphers = make([]string, len(ciphers))
			for i, c := range ciphers {
				stringCiphers[i] = string(c.Raw[:])
			}
			for i, vi := range stringCiphers {
				for j := i + 1; j < len(stringCiphers); j++ {
					Expect(vi).ToNot(Equal(stringCiphers[j]))
				}
			}
		})
	})
})
