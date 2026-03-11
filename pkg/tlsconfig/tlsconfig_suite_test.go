package tlsconfig

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTLSConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TLSConfig Suite")
}
