package test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	testenv "github.com/kubevirt/cluster-network-addons-operator/test/env"
	"github.com/kubevirt/cluster-network-addons-operator/test/kubectl"
)

var _ = BeforeSuite(func() {
	testenv.Start()

	Expect(filterCnaoPromRules()).To(Succeed())
})

func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "monitoring Test Suite")
}

func filterCnaoPromRules() error {
	_, _, err := kubectl.Kubectl("patch", "prometheus", "k8s", "-n", prometheusMonitoringNamespace, "--type=json", "-p",
		"[{'op': 'replace', 'path': '/spec/ruleSelector', 'value':{'matchLabels': {'prometheus.cnao.io': 'true'}}}]")
	return err
}

var _ = AfterEach(func() {
	By("Performing cleanup")
})
