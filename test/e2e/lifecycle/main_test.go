package test

import (
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kubevirt/cluster-network-addons-operator/pkg/components"
	. "github.com/kubevirt/cluster-network-addons-operator/test/check"
	testenv "github.com/kubevirt/cluster-network-addons-operator/test/env"
	. "github.com/kubevirt/cluster-network-addons-operator/test/operations"
	. "github.com/kubevirt/cluster-network-addons-operator/test/releases"
	"github.com/kubevirt/cluster-network-addons-operator/test/reporter"
)

var cnaoReporter *reporter.KubernetesCNAOReporter

func TestE2E(t *testing.T) {
	cnaoReporter = reporter.New("_out/e2e/lifecycle/", components.Namespace)
	cnaoReporter.Cleanup()

	RegisterFailHandler(Fail)
	RunSpecs(t, "lifecycle Test Suite")
}

var _ = BeforeSuite(func() {

	// Change to root directory some test expect that
	os.Chdir("../../../")

	testenv.Start()
})

var _ = JustAfterEach(func() {
	if CurrentSpecReport().Failed() {
		failureCount := cnaoReporter.DumpLogs()
		By(fmt.Sprintf("Test failed, collected logs and artifacts, failure count %d", failureCount))
	}
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
