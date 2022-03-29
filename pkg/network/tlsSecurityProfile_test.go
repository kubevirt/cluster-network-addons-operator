package network

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
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
			expectedCiphers:       ocpv1.TLSProfiles[ocpv1.TLSProfileIntermediateType].Ciphers,
			expectedMinTLSVersion: ocpv1.TLSProfiles[ocpv1.TLSProfileIntermediateType].MinTLSVersion,
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
})
