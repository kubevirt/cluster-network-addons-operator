package test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kubevirt/cluster-network-addons-operator/test/check"
	testenv "github.com/kubevirt/cluster-network-addons-operator/test/env"
	. "github.com/kubevirt/cluster-network-addons-operator/test/operations"
	. "github.com/kubevirt/cluster-network-addons-operator/test/releases"
)

func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "lifecycle Test Suite")
}

var _ = BeforeSuite(func() {

	// Change to root directory some test expect that
	os.Chdir("../../../")

	testenv.Start()
})

var _ = AfterEach(func() {
	By("Performing cleanup")
	gvk := GetCnaoV1GroupVersionKind()
	if GetConfig(gvk) != nil {
		DeleteConfig(gvk)
	}
	CheckComponentsRemoval(AllComponents)
	for _, release := range Releases() {
		UninstallRelease(release)
	}
})
