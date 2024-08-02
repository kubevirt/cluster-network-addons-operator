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
	var ciphers []string
	var minTlsVersion ocpv1.TLSProtocolVersion
	if profile.Custom != nil {
		ciphers = profile.Custom.TLSProfileSpec.Ciphers
		minTlsVersion = profile.Custom.TLSProfileSpec.MinTLSVersion
	} else {
		ciphers = ocpv1.TLSProfiles[profile.Type].Ciphers
		minTlsVersion = ocpv1.TLSProfiles[profile.Type].MinTLSVersion
	}
	m := make(map[string]bool)
	var result []string
	for _, c := range ciphers {
		if m[c] {
			continue
		}
		m[c] = true
		result = append(result, c)
	}

	return result, minTlsVersion
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
