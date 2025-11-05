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
	. "github.com/kubevirt/cluster-network-addons-operator/test/check"
	testenv "github.com/kubevirt/cluster-network-addons-operator/test/env"
	"github.com/kubevirt/cluster-network-addons-operator/test/kubectl"
	. "github.com/kubevirt/cluster-network-addons-operator/test/operations"
	"github.com/kubevirt/cluster-network-addons-operator/test/reporter"
)

var cnaoReporter *reporter.KubernetesCNAOReporter

var _ = BeforeSuite(func() {
	// Set up controller-runtime logger to avoid warnings
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	os.Chdir("../../../")

	testenv.Start()

	Expect(filterCnaoPromRules()).To(Succeed())
})

func TestE2E(t *testing.T) {
	cnaoReporter = reporter.New("_out/e2e/monitoring/", components.Namespace)
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
	PrintOperatorPodStability()
	By("Performing cleanup")
	gvk := GetCnaoV1GroupVersionKind()
	if GetConfig(gvk) != nil {
		DeleteConfig(gvk)
	}
	CheckComponentsRemoval(AllComponents)
})
