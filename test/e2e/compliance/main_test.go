package test

import (
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/kubevirt/cluster-network-addons-operator/pkg/components"
	testenv "github.com/kubevirt/cluster-network-addons-operator/test/env"
	"github.com/kubevirt/cluster-network-addons-operator/test/reporter"
)

var cnaoReporter *reporter.KubernetesCNAOReporter

func TestCompliance(t *testing.T) {
	cnaoReporter = reporter.New("_out/e2e/compliance/", components.Namespace)
	cnaoReporter.Cleanup()

	RegisterFailHandler(Fail)
	RunSpecs(t, "Compliance Test Suite")
}

var _ = BeforeSuite(func() {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	os.Chdir("../../../")

	testenv.Start()
})

var _ = JustAfterEach(func() {
	if CurrentSpecReport().Failed() {
		failureCount := cnaoReporter.DumpLogs()
		By(fmt.Sprintf("Test failed, collected logs and artifacts, failure count %d", failureCount))
	}

	By("Exporting TLSCompliance reports")
	Expect(cnaoReporter.DumpTLSComplianceReports()).To(Succeed())
})
