package components

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const imageName = "the-image-name"

var _ = Describe("Components Tests", func() {
	Context("RelatedImage Test", func() {
		It("Should extruct a short image name from image name", func() {
			image := "imageRegistry/organization/" + imageName + ":1.2.3"
			ri := NewRelatedImage(image)
			Expect(ri.Ref).To(Equal(image))
			Expect(ri.Name).To(Equal(imageName))
		})

		It("Should extruct a short image name from image digest", func() {
			image := "imageRegistry/organization/" + imageName + "@sha256:76cc13fb4a60943dca6038619599b6a49fe451852aba23ad3046658429a9af30"
			ri := NewRelatedImage(image)
			Expect(ri.Ref).To(Equal(image))
			Expect(ri.Name).To(Equal(imageName))
		})

		It("Should copy the imaage name if not it not match to the standard", func() {
			image := "justAstring" + imageName + "anotherString"
			ri := NewRelatedImage(image)
			Expect(ri.Ref).To(Equal(image))
			Expect(ri.Name).To(Equal(image))
		})
	})

	Context("RelatedImages Tests", func() {
		It(`"constructor" should create an empty list when there are no parameters`, func() {
			ris := NewRelatedImages()
			Expect(len(ris)).To(Equal(0))
		})

		It(`"constructor" should create a list of related images if getting image names as parameters`, func() {
			ris := NewRelatedImages(
				"nfvpe/multus@sha256:167722b954355361bd69829466f27172b871dbdbf86b85a95816362885dc0aba",
				"quay.io/kubevirt/cni-default-plugins@sha256:680ac8fd5eeab39c9a3c01479da344bdcaa43aa065d07ae00513b7bafa22fccf",
				"quay.io/kubevirt/ovs-cni-marker@sha256:0f08d6b1550a90c9f10221f2bb07709d1090e7c675ee1a711981bd429074d620",
			)
			Expect(len(ris)).To(Equal(3))

			Expect(ris[0].Name).To(Equal("multus"))
			Expect(ris[0].Ref).To(Equal("nfvpe/multus@sha256:167722b954355361bd69829466f27172b871dbdbf86b85a95816362885dc0aba"))
			Expect(ris[1].Name).To(Equal("cni-default-plugins"))
			Expect(ris[1].Ref).To(Equal("quay.io/kubevirt/cni-default-plugins@sha256:680ac8fd5eeab39c9a3c01479da344bdcaa43aa065d07ae00513b7bafa22fccf"))
			Expect(ris[2].Name).To(Equal("ovs-cni-marker"))
			Expect(ris[2].Ref).To(Equal("quay.io/kubevirt/ovs-cni-marker@sha256:0f08d6b1550a90c9f10221f2bb07709d1090e7c675ee1a711981bd429074d620"))
		})

		It("should add image to an empty list", func() {
			var ris RelatedImages
			Expect(len(ris)).To(Equal(0))

			ris.Add("nfvpe/multus@sha256:167722b954355361bd69829466f27172b871dbdbf86b85a95816362885dc0aba")
			Expect(len(ris)).To(Equal(1))
			Expect(ris[0].Name).To(Equal("multus"))
			Expect(ris[0].Ref).To(Equal("nfvpe/multus@sha256:167722b954355361bd69829466f27172b871dbdbf86b85a95816362885dc0aba"))
		})

		It("should add image to an non-empty list", func() {
			ris := NewRelatedImages(
				"nfvpe/multus@sha256:167722b954355361bd69829466f27172b871dbdbf86b85a95816362885dc0aba",
				"quay.io/kubevirt/cni-default-plugins@sha256:680ac8fd5eeab39c9a3c01479da344bdcaa43aa065d07ae00513b7bafa22fccf",
			)
			Expect(len(ris)).To(Equal(2))

			Expect(ris[0].Name).To(Equal("multus"))
			Expect(ris[0].Ref).To(Equal("nfvpe/multus@sha256:167722b954355361bd69829466f27172b871dbdbf86b85a95816362885dc0aba"))
			Expect(ris[1].Name).To(Equal("cni-default-plugins"))
			Expect(ris[1].Ref).To(Equal("quay.io/kubevirt/cni-default-plugins@sha256:680ac8fd5eeab39c9a3c01479da344bdcaa43aa065d07ae00513b7bafa22fccf"))

			ris.Add("quay.io/kubevirt/ovs-cni-marker@sha256:0f08d6b1550a90c9f10221f2bb07709d1090e7c675ee1a711981bd429074d620")

			Expect(len(ris)).To(Equal(3))

			Expect(ris[0].Name).To(Equal("multus"))
			Expect(ris[0].Ref).To(Equal("nfvpe/multus@sha256:167722b954355361bd69829466f27172b871dbdbf86b85a95816362885dc0aba"))
			Expect(ris[1].Name).To(Equal("cni-default-plugins"))
			Expect(ris[1].Ref).To(Equal("quay.io/kubevirt/cni-default-plugins@sha256:680ac8fd5eeab39c9a3c01479da344bdcaa43aa065d07ae00513b7bafa22fccf"))
			Expect(ris[2].Name).To(Equal("ovs-cni-marker"))
			Expect(ris[2].Ref).To(Equal("quay.io/kubevirt/ovs-cni-marker@sha256:0f08d6b1550a90c9f10221f2bb07709d1090e7c675ee1a711981bd429074d620"))
		})
	})
})
