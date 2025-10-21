package test

import (
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	testenv "github.com/kubevirt/cluster-network-addons-operator/test/env"
	"github.com/kubevirt/cluster-network-addons-operator/test/kubectl"
	"github.com/kubevirt/cluster-network-addons-operator/test/reporter"
)

var cnaoReporter *reporter.KubernetesCNAOReporter

var _ = BeforeSuite(func() {
	os.Chdir("../../../")

	testenv.Start()

	Expect(filterCnaoPromRules()).To(Succeed())
})

func TestE2E(t *testing.T) {
	cnaoReporter = reporter.New("_out/e2e/monitoring/", prometheusMonitoringNamespace)
	cnaoReporter.Cleanup()

	RegisterFailHandler(Fail)
	RunSpecs(t, "monitoring Test Suite")
}

func filterCnaoPromRules() error {
	_, _, err := kubectl.Kubectl("patch", "prometheus", "k8s", "-n", prometheusMonitoringNamespace, "--type=json", "-p",
		"[{'op': 'replace', 'path': '/spec/ruleSelector', 'value':{'matchLabels': {'prometheus.cnao.io': 'true'}}}]")
	return err
}

var _ = JustAfterEach(func() {
	if CurrentSpecReport().Failed() {
		failureCount := cnaoReporter.DumpLogs()
		By(fmt.Sprintf("Test failed, collected logs and artifacts, failure count %d", failureCount))
	}
})

var _ = AfterEach(func() {
	By("Performing cleanup")
})
