package tlsconfig

import (
	"crypto/tls"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	ocpv1 "github.com/openshift/api/config/v1"

	"github.com/kubevirt/cluster-network-addons-operator/pkg/network"
)

var _ = Describe("MutateTLSConfig", func() {
	It("should install a GetConfigForClient callback", func() {
		cache := &Cache{}
		cfg := &tls.Config{}
		MutateTLSConfig(cache)(cfg)

		Expect(cfg.GetConfigForClient).ToNot(BeNil())
	})

	It("should apply Modern defaults when cache is empty", func() {
		cache := &Cache{}
		cfg := &tls.Config{}
		MutateTLSConfig(cache)(cfg)

		result, err := cfg.GetConfigForClient(&tls.ClientHelloInfo{})
		Expect(err).ToNot(HaveOccurred())
		Expect(result.MinVersion).To(Equal(uint16(tls.VersionTLS13)))
	})

	It("should reflect the Intermediate profile from the cache", func() {
		cache := &Cache{}
		cache.Store(&ocpv1.TLSSecurityProfile{
			Type:         ocpv1.TLSProfileIntermediateType,
			Intermediate: &ocpv1.IntermediateTLSProfile{},
		})

		cfg := &tls.Config{}
		MutateTLSConfig(cache)(cfg)

		result, err := cfg.GetConfigForClient(&tls.ClientHelloInfo{})
		Expect(err).ToNot(HaveOccurred())
		Expect(result.MinVersion).To(Equal(uint16(tls.VersionTLS12)))

		expectedCiphers, _ := network.SelectCipherSuitesAndMinTLSVersion(cache.Load())
		Expect(result.CipherSuites).To(Equal(network.CipherSuiteIDs(expectedCiphers)))
	})

	It("should pick up profile changes without restart", func() {
		cache := &Cache{}
		cache.Store(&ocpv1.TLSSecurityProfile{
			Type:         ocpv1.TLSProfileIntermediateType,
			Intermediate: &ocpv1.IntermediateTLSProfile{},
		})

		cfg := &tls.Config{}
		MutateTLSConfig(cache)(cfg)

		result, err := cfg.GetConfigForClient(&tls.ClientHelloInfo{})
		Expect(err).ToNot(HaveOccurred())
		Expect(result.MinVersion).To(Equal(uint16(tls.VersionTLS12)))

		cache.Store(&ocpv1.TLSSecurityProfile{
			Type:   ocpv1.TLSProfileModernType,
			Modern: &ocpv1.ModernTLSProfile{},
		})

		result, err = cfg.GetConfigForClient(&tls.ClientHelloInfo{})
		Expect(err).ToNot(HaveOccurred())
		Expect(result.MinVersion).To(Equal(uint16(tls.VersionTLS13)))
	})
})
