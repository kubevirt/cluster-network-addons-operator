package tlsconfig

import (
	"crypto/tls"

	"github.com/kubevirt/cluster-network-addons-operator/pkg/network"
)

// MutateTLSConfig returns a function suitable for controller-runtime's
// metricsserver.Options.TLSOpts. It installs a GetConfigForClient callback
// that reads the latest TLS profile from the cache on every new TLS
// handshake, so cipher suites and minimum version stay in sync with the
// NetworkAddonsConfig CR without requiring a pod restart.
func MutateTLSConfig(cache *Cache) func(*tls.Config) {
	return func(cfg *tls.Config) {
		cfg.GetConfigForClient = func(_ *tls.ClientHelloInfo) (*tls.Config, error) {
			ciphers, minVersion := network.SelectCipherSuitesAndMinTLSVersion(cache.Load())
			cfg.CipherSuites = network.CipherSuiteIDs(ciphers)
			cfg.MinVersion = network.TLSMinVersionID(minVersion)
			return cfg, nil
		}
	}
}
