package network

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	v1 "k8s.io/api/core/v1"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
)

var _ = Describe("Testing image-pull-policy", func() {
	Describe("validateImagePullPolicy", func() {
		Context("when configuration uses invalid policy type", func() {
			spec := &cnao.NetworkAddonsConfigSpec{ImagePullPolicy: v1.PullPolicy("BAD")}

			It("should fail", func() {
				errorList := validateImagePullPolicy(spec)
				Expect(errorList).To(HaveLen(1), "validation failed due to an unexpected error: %v", errorList)
				Expect(errorList[0]).To(MatchError("requested imagePullPolicy 'BAD' is not valid"))
			})
		})

		Context("when configuration uses a valid policy type", func() {
			spec := &cnao.NetworkAddonsConfigSpec{ImagePullPolicy: v1.PullAlways}

			It("should pass", func() {
				errorList := validateImagePullPolicy(spec)
				Expect(errorList).To(BeEmpty())
			})
		})
	})

	Describe("fillDefaultsImagePullPolicy", func() {
		Context("when no policy specified", func() {
			Context("and there was a policy specified in the previous config", func() {
				new := &cnao.NetworkAddonsConfigSpec{}
				prev := &cnao.NetworkAddonsConfigSpec{ImagePullPolicy: v1.PullAlways}

				It("should successfully pass", func() {
					errorList := fillDefaultsImagePullPolicy(new, prev)
					Expect(errorList).To(BeEmpty())
				})

				It("and fill in the previously defined policy", func() {
					Expect(new.ImagePullPolicy).To(Equal(v1.PullAlways))
				})
			})

			Context("and there was no policy specified in the last config", func() {
				new := &cnao.NetworkAddonsConfigSpec{}
				prev := &cnao.NetworkAddonsConfigSpec{}

				It("should successfully pass", func() {
					errorList := fillDefaultsImagePullPolicy(new, prev)
					Expect(errorList).To(BeEmpty())
				})

				It("should fill in the default policy", func() {
					Expect(new.ImagePullPolicy).To(Equal(defaultImagePullPolicy))
					// Following is in an ideal case a duplicate of previous line.
					// The reason we have this check is, that some tests in this module
					// are expecting this to be the default and they would need some
					// changes in case the default changes.
					Expect(new.ImagePullPolicy).To(Equal(v1.PullIfNotPresent))
				})
			})
		})
	})

	Describe("changeSafeImagePullPolicy", func() {
		Context("when it is kept disabled", func() {
			prev := &cnao.NetworkAddonsConfigSpec{}
			new := &cnao.NetworkAddonsConfigSpec{}

			It("should pass", func() {
				errorList := changeSafeImagePullPolicy(prev, new)
				Expect(errorList).To(BeEmpty())
			})
		})

		Context("when there is no previous value", func() {
			prev := &cnao.NetworkAddonsConfigSpec{}
			new := &cnao.NetworkAddonsConfigSpec{LinuxBridge: &cnao.LinuxBridge{}}

			It("should accept any configuration", func() {
				errorList := changeSafeImagePullPolicy(prev, new)
				Expect(errorList).To(BeEmpty())
			})
		})

		Context("when the previous and new configuration match", func() {
			prev := &cnao.NetworkAddonsConfigSpec{LinuxBridge: &cnao.LinuxBridge{}}
			new := &cnao.NetworkAddonsConfigSpec{LinuxBridge: &cnao.LinuxBridge{}}

			It("should accept the configuration", func() {
				errorList := changeSafeImagePullPolicy(prev, new)
				Expect(errorList).To(BeEmpty())
			})
		})

		Context("when there is previous value, but the new one is empty (removing component)", func() {
			prev := &cnao.NetworkAddonsConfigSpec{LinuxBridge: &cnao.LinuxBridge{}}
			new := &cnao.NetworkAddonsConfigSpec{}

			// If ImagePullPolicy is omitted, default or previously applied will be used
			It("should pass", func() {
				errorList := changeSafeImagePullPolicy(prev, new)
				Expect(errorList).To(BeEmpty())
			})
		})
	})
})
