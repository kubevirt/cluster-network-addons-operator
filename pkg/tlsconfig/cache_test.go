package tlsconfig

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	ocpv1 "github.com/openshift/api/config/v1"
)

var _ = Describe("Cache", func() {
	var cache *Cache

	BeforeEach(func() {
		cache = &Cache{}
	})

	It("should return nil before any Store", func() {
		Expect(cache.Load()).To(BeNil())
	})

	It("should return the stored profile", func() {
		profile := &ocpv1.TLSSecurityProfile{
			Type:         ocpv1.TLSProfileIntermediateType,
			Intermediate: &ocpv1.IntermediateTLSProfile{},
		}
		cache.Store(profile)
		Expect(cache.Load()).To(Equal(profile))
	})

	It("should reflect the latest Store", func() {
		intermediate := &ocpv1.TLSSecurityProfile{
			Type:         ocpv1.TLSProfileIntermediateType,
			Intermediate: &ocpv1.IntermediateTLSProfile{},
		}
		modern := &ocpv1.TLSSecurityProfile{
			Type:   ocpv1.TLSProfileModernType,
			Modern: &ocpv1.ModernTLSProfile{},
		}

		cache.Store(intermediate)
		Expect(cache.Load()).To(Equal(intermediate))

		cache.Store(modern)
		Expect(cache.Load()).To(Equal(modern))
	})

	It("should allow storing nil to reset", func() {
		profile := &ocpv1.TLSSecurityProfile{
			Type:   ocpv1.TLSProfileModernType,
			Modern: &ocpv1.ModernTLSProfile{},
		}
		cache.Store(profile)
		cache.Store(nil)
		Expect(cache.Load()).To(BeNil())
	})
})
