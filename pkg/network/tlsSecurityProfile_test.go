package network

import (
	"crypto/tls"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	ocpv1 "github.com/openshift/api/config/v1"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
)

var _ = Describe("TLS Security Profile", func() {
	type loadSecurityProfileCase struct {
		config                *cnao.NetworkAddonsConfigSpec
		expectedCiphers       []string
		expectedMinTLSVersion ocpv1.TLSProtocolVersion
		expectedGroups        []ocpv1.TLSGroup
		expectedCurveIDs      []tls.CurveID
	}
	testCustomTLSProfileSpec := ocpv1.TLSProfileSpec{
		Ciphers:       []string{"foo,bar"},
		MinTLSVersion: "foobar",
		Groups:        []ocpv1.TLSGroup{"mygroup"},
	}
	testCurveIDs := []tls.CurveID{tls.X25519MLKEM768, tls.X25519, tls.CurveP256, tls.CurveP384}

	DescribeTable("SelectTLSSettings should aggregate valid selected TLS settings",
		func(c loadSecurityProfileCase) {
			tlsCfg := SelectTLSSettings(c.config.TLSSecurityProfile)
			Expect(tlsCfg).To(Equal(tlsConfig{
				MinVersion:   c.expectedMinTLSVersion,
				CipherSuites: c.expectedCiphers,
				Groups:       c.expectedGroups,
			}))

			Expect(CurveIDs(tlsCfg.Groups)).To(Equal(c.expectedCurveIDs))
		},
		Entry("when TLSSecurityProfile is nil", loadSecurityProfileCase{
			config:                &cnao.NetworkAddonsConfigSpec{},
			expectedCiphers:       ocpv1.TLSProfiles[ocpv1.TLSProfileModernType].Ciphers,
			expectedMinTLSVersion: ocpv1.TLSProfiles[ocpv1.TLSProfileModernType].MinTLSVersion,
			expectedGroups:        ocpv1.TLSProfiles[ocpv1.TLSProfileModernType].Groups,
			expectedCurveIDs:      testCurveIDs,
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
			expectedGroups:        ocpv1.TLSProfiles[ocpv1.TLSProfileOldType].Groups,
			expectedCurveIDs:      testCurveIDs,
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
			expectedGroups:        ocpv1.TLSProfiles[ocpv1.TLSProfileIntermediateType].Groups,
			expectedCurveIDs:      testCurveIDs,
		}),
		Entry("when Modern Security Profile is selected", loadSecurityProfileCase{
			config: &cnao.NetworkAddonsConfigSpec{
				TLSSecurityProfile: &ocpv1.TLSSecurityProfile{
					Type:   ocpv1.TLSProfileModernType,
					Modern: &ocpv1.ModernTLSProfile{},
				},
			},
			expectedCiphers:       ocpv1.TLSProfiles[ocpv1.TLSProfileModernType].Ciphers,
			expectedMinTLSVersion: ocpv1.TLSProfiles[ocpv1.TLSProfileModernType].MinTLSVersion,
			expectedGroups:        ocpv1.TLSProfiles[ocpv1.TLSProfileModernType].Groups,
			expectedCurveIDs:      testCurveIDs,
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
			expectedGroups:        testCustomTLSProfileSpec.Groups,
		}),
		Entry("when Custom Security Profile, explicit no TLS groups set", loadSecurityProfileCase{
			config: &cnao.NetworkAddonsConfigSpec{
				TLSSecurityProfile: &ocpv1.TLSSecurityProfile{
					Type: ocpv1.TLSProfileCustomType,
					Custom: &ocpv1.CustomTLSProfile{TLSProfileSpec: ocpv1.TLSProfileSpec{
						MinTLSVersion: "myversion",
						Ciphers:       []string{"myciphers"},
					}},
				},
			},
			expectedMinTLSVersion: "myversion",
			expectedCiphers:       []string{"myciphers"},
			expectedGroups:        nil,
			expectedCurveIDs:      nil,
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
			var tlsCfg = SelectTLSSettings(profile)
			for i, vi := range tlsCfg.CipherSuites {
				for j := i + 1; j < len(tlsCfg.CipherSuites); j++ {
					Expect(vi).ToNot(Equal(tlsCfg.CipherSuites[j]))
				}
			}
		})
	})

	It("when duplicate TLS groups are selected should aggregate unique ones", func() {
		profile := &ocpv1.TLSSecurityProfile{
			Type: ocpv1.TLSProfileCustomType,
			Custom: &ocpv1.CustomTLSProfile{TLSProfileSpec: ocpv1.TLSProfileSpec{
				Groups: []ocpv1.TLSGroup{ocpv1.TLSGroupX25519, ocpv1.TLSGroupX25519, ocpv1.TLSGroupX25519},
			}},
		}
		tlsCfg := SelectTLSSettings(profile)
		Expect(tlsCfg.Groups).To(Equal([]ocpv1.TLSGroup{ocpv1.TLSGroupX25519}))
	})

	Context("GoTLSCipherSuiteNames", func() {
		It("should convert Intermediate profile ciphers to Go crypto/tls names", func() {
			tlsCfg := SelectTLSSettings(&ocpv1.TLSSecurityProfile{
				Type:         ocpv1.TLSProfileIntermediateType,
				Intermediate: &ocpv1.IntermediateTLSProfile{},
			})
			goNames := OCPTLSProfileCiphersToGoCipherNames(tlsCfg.CipherSuites)

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
			tlsCfg := SelectTLSSettings(&ocpv1.TLSSecurityProfile{
				Type:   ocpv1.TLSProfileModernType,
				Modern: &ocpv1.ModernTLSProfile{},
			})
			goNames := OCPTLSProfileCiphersToGoCipherNames(tlsCfg.CipherSuites)

			Expect(goNames).To(BeEmpty(),
				"TLS 1.3 ciphers are not configurable in Go and should be excluded")
		})

		It("should convert Old profile ciphers and exclude TLS 1.3 entries", func() {
			tlsCfg := SelectTLSSettings(&ocpv1.TLSSecurityProfile{
				Type: ocpv1.TLSProfileOldType,
				Old:  &ocpv1.OldTLSProfile{},
			})
			goNames := OCPTLSProfileCiphersToGoCipherNames(tlsCfg.CipherSuites)

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

	Context("CipherSuiteIDs", func() {
		It("should skip unknown ciphers", func() {
			ids := CipherSuiteIDs([]string{
				"ECDHE-ECDSA-AES128-GCM-SHA256",
				"TOTALLY-MADE-UP-CIPHER",
			})
			Expect(ids).To(Equal([]uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			}))
		})

		It("should return nil for empty input", func() {
			Expect(CipherSuiteIDs(nil)).To(BeNil())
			Expect(CipherSuiteIDs([]string{})).To(BeNil())
		})

		type cipherSuiteIDsCase struct {
			profile     *ocpv1.TLSSecurityProfile
			expectedIDs []uint16
		}
		DescribeTable("should return expected IDs for each TLS profile",
			func(c cipherSuiteIDsCase) {
				tlsCfg := SelectTLSSettings(c.profile)
				Expect(CipherSuiteIDs(tlsCfg.CipherSuites)).To(Equal(c.expectedIDs))
			},
			Entry("Intermediate profile", cipherSuiteIDsCase{
				profile: &ocpv1.TLSSecurityProfile{
					Type:         ocpv1.TLSProfileIntermediateType,
					Intermediate: &ocpv1.IntermediateTLSProfile{},
				},
				expectedIDs: []uint16{
					tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
					tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
				},
			}),
			Entry("Old profile", cipherSuiteIDsCase{
				profile: &ocpv1.TLSSecurityProfile{
					Type: ocpv1.TLSProfileOldType,
					Old:  &ocpv1.OldTLSProfile{},
				},
				expectedIDs: []uint16{
					tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
					tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
					tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,
					tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
					tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
					tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
					tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
					tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
					tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_RSA_WITH_AES_128_CBC_SHA256,
					tls.TLS_RSA_WITH_AES_128_CBC_SHA,
					tls.TLS_RSA_WITH_AES_256_CBC_SHA,
					tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,
				},
			}),
			Entry("Modern profile (TLS 1.3 only, no configurable cipher IDs)", cipherSuiteIDsCase{
				profile: &ocpv1.TLSSecurityProfile{
					Type:   ocpv1.TLSProfileModernType,
					Modern: &ocpv1.ModernTLSProfile{},
				},
				expectedIDs: nil,
			}),
		)
	})

	Context("TLSMinVersionID", func() {
		DescribeTable("should map known versions correctly",
			func(version ocpv1.TLSProtocolVersion, expected uint16) {
				Expect(TLSMinVersionID(version)).To(Equal(expected))
			},
			Entry("TLS 1.0", ocpv1.VersionTLS10, uint16(tls.VersionTLS10)),
			Entry("TLS 1.1", ocpv1.VersionTLS11, uint16(tls.VersionTLS11)),
			Entry("TLS 1.2", ocpv1.VersionTLS12, uint16(tls.VersionTLS12)),
			Entry("TLS 1.3", ocpv1.VersionTLS13, uint16(tls.VersionTLS13)),
		)

		It("should return 0 for unrecognized values", func() {
			Expect(TLSMinVersionID("")).To(Equal(uint16(0)))
			Expect(TLSMinVersionID("VersionTLS99")).To(Equal(uint16(0)))
		})
	})

	Context("CurveIDs", func() {
		It("should return nil for empty input", func() {
			Expect(CurveIDs(nil)).To(BeNil())
			Expect(CurveIDs([]ocpv1.TLSGroup{})).To(BeNil())
		})
		It("should convert OCP TLSGroup to IDs", func() {
			ids := CurveIDs([]ocpv1.TLSGroup{
				ocpv1.TLSGroupX25519MLKEM768,
				ocpv1.TLSGroupX25519,
				ocpv1.TLSGroupSecP256r1,
				ocpv1.TLSGroupSecP384r1,
				ocpv1.TLSGroupSecP521r1,
				ocpv1.TLSGroupSecP256r1MLKEM768,
				ocpv1.TLSGroupSecP384r1MLKEM1024,
			})
			Expect(ids).To(Equal([]tls.CurveID{
				tls.X25519MLKEM768,
				tls.X25519,
				tls.CurveP256,
				tls.CurveP384,
				tls.CurveP521,
				tls.SecP256r1MLKEM768,
				tls.SecP384r1MLKEM1024,
			}))
		})
		It("should skip groups with no known crypto/tls.CurveID", func() {
			ids := CurveIDs([]ocpv1.TLSGroup{
				ocpv1.TLSGroupX25519,
				"my-non-exist-tls-group",
			})
			Expect(ids).To(Equal([]tls.CurveID{tls.X25519}))
		})
	})

	Context("OCPTLSProfileTLSGroupToGoTLSGroupNames", func() {
		It("should return nil for empty input", func() {
			Expect(OCPTLSProfileTLSGroupToGoTLSGroupNames(nil)).To(BeNil())
			Expect(OCPTLSProfileTLSGroupToGoTLSGroupNames([]ocpv1.TLSGroup{})).To(BeNil())
		})
		It("should convert OCP TLSGroup to Go crypto/tls constant names", func() {
			names := OCPTLSProfileTLSGroupToGoTLSGroupNames([]ocpv1.TLSGroup{
				ocpv1.TLSGroupX25519MLKEM768,
				ocpv1.TLSGroupX25519,
				ocpv1.TLSGroupSecP256r1,
				ocpv1.TLSGroupSecP384r1,
				ocpv1.TLSGroupSecP521r1,
				ocpv1.TLSGroupSecP256r1MLKEM768,
				ocpv1.TLSGroupSecP384r1MLKEM1024,
			})
			Expect(names).To(Equal([]string{
				"X25519MLKEM768",
				"X25519",
				"CurveP256",
				"CurveP384",
				"CurveP521",
				"SecP256r1MLKEM768",
				"SecP384r1MLKEM1024",
			}))
		})
		It("should skip groups with no known crypto/tls.CurveID", func() {
			names := OCPTLSProfileTLSGroupToGoTLSGroupNames([]ocpv1.TLSGroup{
				ocpv1.TLSGroupX25519,
				"my-non-exist-tls-group",
			})
			Expect(names).To(Equal([]string{"X25519"}))
		})
	})
})
