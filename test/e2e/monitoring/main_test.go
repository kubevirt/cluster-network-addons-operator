package test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	ginkgoreporters "kubevirt.io/qe-tools/pkg/ginkgo-reporters"

	"github.com/kubevirt/cluster-network-addons-operator/pkg/components"
	testenv "github.com/kubevirt/cluster-network-addons-operator/test/env"
	"github.com/kubevirt/cluster-network-addons-operator/test/kubectl"
	cnaoreporter "github.com/kubevirt/cluster-network-addons-operator/test/reporter"
)

var _ = BeforeSuite(func() {
	testenv.Start()

	Expect(filterCnaoPromRules()).To(Succeed())
})

func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	reporters := make([]Reporter, 0)
	reporters = append(reporters, cnaoreporter.New("test_logs/e2e/monitoring", components.Namespace))
	if ginkgoreporters.JunitOutput != "" {
		reporters = append(reporters, ginkgoreporters.NewJunitReporter())
	}
	RunSpecsWithDefaultAndCustomReporters(t, "Monitoring E2E Test Suite", reporters)
}

func filterCnaoPromRules() error {
	_, _, err := kubectl.Kubectl("patch", "prometheus", "k8s", "-n", prometheusMonitoringNamespace, "--type=json", "-p",
		"[{'op': 'replace', 'path': '/spec/ruleSelector', 'value':{'matchLabels': {'prometheus.cnao.io': 'true'}}}]")
	return err
}

var _ = AfterEach(func() {
	By("Performing cleanup")
})
