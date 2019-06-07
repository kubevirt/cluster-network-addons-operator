package statusmanager

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestStatusManager(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Status Manager Suite")
}
