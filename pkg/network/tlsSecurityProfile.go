package network

import (
	"crypto/tls"

	ocpv1 "github.com/openshift/api/config/v1"
)

var possibleCipherSuites = map[uint16]string{}

func init() {
	for _, c := range tls.CipherSuites() {
		possibleCipherSuites[c.ID] = c.Name
	}
	for _, c := range tls.InsecureCipherSuites() {
		possibleCipherSuites[c.ID] = c.Name
	}
}

// ocpTLSProfileOpenSSLCipherNames maps OpenSSL cipher suite names
// used in OpenShift TLS profiles to Go crypto/tls cipher suite IDs.
// Ref: https://www.iana.org/assignments/tls-parameters/tls-parameters.xml
var ocpTLSProfileOpenSSLCipherNames = map[string]uint16{
	"ECDHE-ECDSA-AES128-GCM-SHA256": tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	"ECDHE-RSA-AES128-GCM-SHA256":   tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	"ECDHE-ECDSA-AES256-GCM-SHA384": tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
	"ECDHE-RSA-AES256-GCM-SHA384":   tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	"ECDHE-ECDSA-CHACHA20-POLY1305": tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
	"ECDHE-RSA-CHACHA20-POLY1305":   tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
	"ECDHE-ECDSA-AES128-SHA256":     tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,
	"ECDHE-RSA-AES128-SHA256":       tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
	"AES128-GCM-SHA256":             tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
	"AES256-GCM-SHA384":             tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
	"AES128-SHA256":                 tls.TLS_RSA_WITH_AES_128_CBC_SHA256,
	"ECDHE-ECDSA-AES128-SHA":        tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
	"ECDHE-RSA-AES128-SHA":          tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
	"ECDHE-ECDSA-AES256-SHA":        tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
	"ECDHE-RSA-AES256-SHA":          tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
	"AES128-SHA":                    tls.TLS_RSA_WITH_AES_128_CBC_SHA,
	"AES256-SHA":                    tls.TLS_RSA_WITH_AES_256_CBC_SHA,
	"DES-CBC3-SHA":                  tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,
}

func SelectCipherSuitesAndMinTLSVersion(profile *ocpv1.TLSSecurityProfile) ([]string, ocpv1.TLSProtocolVersion) {
	if profile == nil {
		profile = &ocpv1.TLSSecurityProfile{
			Type:   ocpv1.TLSProfileModernType,
			Modern: &ocpv1.ModernTLSProfile{},
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

// OCPTLSProfileCiphersToGoCipherNames converts OpenSSL-format cipher names used in OpenShift TLS profiles
// to Go crypto/tls constant names.
// TLs 1.3 cipher names are not included in the result.
// Go's cipher suite names are resolved from the runtime via tls.CipherSuites and tls.InsecureCipherSuites.
func OCPTLSProfileCiphersToGoCipherNames(openSSLCiphers []string) []string {
	var result []string
	for _, c := range openSSLCiphers {
		if id, ok := ocpTLSProfileOpenSSLCipherNames[c]; ok {
			if name, ok := possibleCipherSuites[id]; ok {
				result = append(result, name)
			}
		}
	}
	return result
}

// CipherSuiteIDs converts OpenSSL cipher names to crypto/tls uint16 IDs
// suitable for tls.Config.CipherSuites. Unknown names are silently skipped.
func CipherSuiteIDs(openSSLCiphers []string) []uint16 {
	var ids []uint16
	for _, c := range openSSLCiphers {
		if id, ok := ocpTLSProfileOpenSSLCipherNames[c]; ok {
			ids = append(ids, id)
		}
	}
	return ids
}

var tlsVersionID = map[ocpv1.TLSProtocolVersion]uint16{
	ocpv1.VersionTLS10: tls.VersionTLS10,
	ocpv1.VersionTLS11: tls.VersionTLS11,
	ocpv1.VersionTLS12: tls.VersionTLS12,
	ocpv1.VersionTLS13: tls.VersionTLS13,
}

// TLSMinVersionID converts an OpenShift TLSProtocolVersion to the crypto/tls
// uint16 constant suitable for tls.Config.MinVersion.
// An unrecognized version returns 0, which causes crypto/tls to use its
// default minimum (currently TLS 1.2).
func TLSMinVersionID(version ocpv1.TLSProtocolVersion) uint16 {
	return tlsVersionID[version]
}
