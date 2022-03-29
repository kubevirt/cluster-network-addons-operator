package network

import (
	ocpv1 "github.com/openshift/api/config/v1"
)

func SelectCipherSuitesAndMinTLSVersion(profile *ocpv1.TLSSecurityProfile) ([]string, ocpv1.TLSProtocolVersion) {
	if profile == nil {
		profile = &ocpv1.TLSSecurityProfile{
			Type:         ocpv1.TLSProfileIntermediateType,
			Intermediate: &ocpv1.IntermediateTLSProfile{},
		}
	}
	if profile.Custom != nil {
		return profile.Custom.TLSProfileSpec.Ciphers, profile.Custom.TLSProfileSpec.MinTLSVersion
	}
	return ocpv1.TLSProfiles[profile.Type].Ciphers, ocpv1.TLSProfiles[profile.Type].MinTLSVersion
}

func TLSVersionToHumanReadable(version ocpv1.TLSProtocolVersion) string {
	switch version {
	case ocpv1.VersionTLS10:
		return "1.0"
	case ocpv1.VersionTLS11:
		return "1.1"
	case ocpv1.VersionTLS12:
		return "1.2"
	case ocpv1.VersionTLS13:
		return "1.3"
	default:
		return ""
	}
}
