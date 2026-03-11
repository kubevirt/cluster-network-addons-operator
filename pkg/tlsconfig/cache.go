package tlsconfig

import (
	"sync/atomic"

	ocpv1 "github.com/openshift/api/config/v1"
)

// Cache provides thread-safe access to the active TLS security profile.
// The reconciler calls Store after reading the CR; the TLS handshake
// callback calls Load to build a fresh tls.Config on every connection.
type Cache struct {
	profile atomic.Pointer[ocpv1.TLSSecurityProfile]
}

func (c *Cache) Store(profile *ocpv1.TLSSecurityProfile) {
	c.profile.Store(profile)
}

// Load returns the current profile, or nil if none has been stored yet.
// Callers should treat nil as "use default" (Modern / TLS 1.3).
func (c *Cache) Load() *ocpv1.TLSSecurityProfile {
	return c.profile.Load()
}
