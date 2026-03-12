package network

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	ocpv1 "github.com/openshift/api/config/v1"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
)

var _ = Describe("Testing TLS Security Profile", func() {
	type loadSecurityProfileCase struct {
		config                *cnao.NetworkAddonsConfigSpec
		expectedCiphers       []string
		expectedMinTLSVersion ocpv1.TLSProtocolVersion
	}
	testCustomTLSProfileSpec := ocpv1.TLSProfileSpec{
		Ciphers:       []string{"foo,bar"},
		MinTLSVersion: "foobar",
	}
	DescribeTable("SecurityProfileSpec function",
		func(c loadSecurityProfileCase) {
			ciphers, minTLSVersion := SelectCipherSuitesAndMinTLSVersion(c.config.TLSSecurityProfile)
			Expect(ciphers).To(Equal(c.expectedCiphers))
			Expect(minTLSVersion).To(Equal(c.expectedMinTLSVersion))
		},
		Entry("when TLSSecurityProfile is nil", loadSecurityProfileCase{
			config:                &cnao.NetworkAddonsConfigSpec{},
			expectedCiphers:       ocpv1.TLSProfiles[ocpv1.TLSProfileModernType].Ciphers,
			expectedMinTLSVersion: ocpv1.TLSProfiles[ocpv1.TLSProfileModernType].MinTLSVersion,
		}),
		Entry("when Old Security Profile is selected", loadSecurityProfileCase{
			config: &cnao.NetworkAddonsConfigSpec{
				TLSSecurityProfile: &ocpv1.TLSSecurityProfile{
					Type: ocpv1.TLSProfileOldType,
					Old:  &ocpv1.OldTLSProfile{},
				},
			},
			expectedCiphers:       ocpv1.TLSProfiles[ocpv1.TLSProfileOldType].Ciphers,
			expectedMinTLSVersion: ocpv1.TLSProfiles[ocpv1.TLSProfileOldType].MinTLSVersion,
		}),
		Entry("when Intermediate Security Profile is selected", loadSecurityProfileCase{
			config: &cnao.NetworkAddonsConfigSpec{
				TLSSecurityProfile: &ocpv1.TLSSecurityProfile{
					Type:         ocpv1.TLSProfileIntermediateType,
					Intermediate: &ocpv1.IntermediateTLSProfile{},
				},
			},
			expectedCiphers:       ocpv1.TLSProfiles[ocpv1.TLSProfileIntermediateType].Ciphers,
			expectedMinTLSVersion: ocpv1.TLSProfiles[ocpv1.TLSProfileIntermediateType].MinTLSVersion,
		}),
		Entry("when Custom Security Profile is selected", loadSecurityProfileCase{
			config: &cnao.NetworkAddonsConfigSpec{
				TLSSecurityProfile: &ocpv1.TLSSecurityProfile{
					Type: ocpv1.TLSProfileCustomType,
					Custom: &ocpv1.CustomTLSProfile{
						TLSProfileSpec: testCustomTLSProfileSpec,
					},
				},
			},
			expectedCiphers:       testCustomTLSProfileSpec.Ciphers,
			expectedMinTLSVersion: testCustomTLSProfileSpec.MinTLSVersion,
		}),
	)

	Context("When selecting ciphers", func() {
		It("should not generate duplicates", func() {
			var profile = &ocpv1.TLSSecurityProfile{
				Type: ocpv1.TLSProfileCustomType,
				Custom: &ocpv1.CustomTLSProfile{
					TLSProfileSpec: ocpv1.TLSProfileSpec{
						Ciphers: []string{"foo", "foo", "bar"},
					},
				},
				Intermediate: &ocpv1.IntermediateTLSProfile{},
			}
			var ciphers, _ = SelectCipherSuitesAndMinTLSVersion(profile)
			for i, vi := range ciphers {
				for j := i + 1; j < len(ciphers); j++ {
					Expect(vi).ToNot(Equal(ciphers[j]))
				}
			}
		})
	})

	Context("GoTLSCipherSuiteNames", func() {
		It("should convert Intermediate profile ciphers to Go crypto/tls names", func() {
			ciphers, _ := SelectCipherSuitesAndMinTLSVersion(&ocpv1.TLSSecurityProfile{
				Type:         ocpv1.TLSProfileIntermediateType,
				Intermediate: &ocpv1.IntermediateTLSProfile{},
			})
			goNames := OCPTLSProfileCiphersToGoCipherNames(ciphers)

			Expect(goNames).To(ConsistOf(
				"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
				"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
				"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384",
				"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
				"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256",
				"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256",
			))
		})

		It("should return empty for Modern profile (TLS 1.3 only ciphers)", func() {
			ciphers, _ := SelectCipherSuitesAndMinTLSVersion(&ocpv1.TLSSecurityProfile{
				Type:   ocpv1.TLSProfileModernType,
				Modern: &ocpv1.ModernTLSProfile{},
			})
			goNames := OCPTLSProfileCiphersToGoCipherNames(ciphers)

			Expect(goNames).To(BeEmpty(),
				"TLS 1.3 ciphers are not configurable in Go and should be excluded")
		})

		It("should convert Old profile ciphers and exclude TLS 1.3 entries", func() {
			ciphers, _ := SelectCipherSuitesAndMinTLSVersion(&ocpv1.TLSSecurityProfile{
				Type: ocpv1.TLSProfileOldType,
				Old:  &ocpv1.OldTLSProfile{},
			})
			goNames := OCPTLSProfileCiphersToGoCipherNames(ciphers)

			Expect(goNames).ToNot(BeEmpty())
			Expect(goNames).To(ContainElement("TLS_RSA_WITH_3DES_EDE_CBC_SHA"),
				"Old profile includes DES-CBC3-SHA")
			for _, name := range goNames {
				Expect(name).ToNot(Equal("TLS_AES_128_GCM_SHA256"),
					"TLS 1.3 ciphers should not appear")
			}
		})

		It("should skip unknown cipher names", func() {
			goNames := OCPTLSProfileCiphersToGoCipherNames([]string{
				"ECDHE-ECDSA-AES128-GCM-SHA256",
				"TOTALLY-MADE-UP-CIPHER",
			})
			Expect(goNames).To(Equal([]string{"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256"}))
		})
	})
})
