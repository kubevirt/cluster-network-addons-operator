package test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	ginkgoreporters "kubevirt.io/qe-tools/pkg/ginkgo-reporters"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	f "github.com/operator-framework/operator-sdk/pkg/test"
	framework "github.com/operator-framework/operator-sdk/pkg/test"

	"github.com/kubevirt/cluster-network-addons-operator/pkg/apis"
	cnaov1 "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/v1"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/components"
	. "github.com/kubevirt/cluster-network-addons-operator/test/check"
	. "github.com/kubevirt/cluster-network-addons-operator/test/operations"
	. "github.com/kubevirt/cluster-network-addons-operator/test/releases"
	cnaoreporter "github.com/kubevirt/cluster-network-addons-operator/test/reporter"
)

func TestMain(m *testing.M) {
	f.MainEntry(m)
}

func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	reporters := make([]Reporter, 0)
	reporters = append(reporters, cnaoreporter.New("test_logs/e2e/lifecycle", components.Namespace))
	if ginkgoreporters.JunitOutput != "" {
		reporters = append(reporters, ginkgoreporters.NewJunitReporter())
	}
	RunSpecsWithDefaultAndCustomReporters(t, "Lifecycle E2E Test Suite", reporters)

}

var _ = BeforeSuite(func() {
	By("Adding custom resource scheme to framework")
	err := framework.AddToFrameworkScheme(apis.AddToScheme, &cnaov1.NetworkAddonsConfigList{})
	Expect(err).ToNot(HaveOccurred())

	Expect(framework.AddToFrameworkScheme(monitoringv1.AddToScheme, &monitoringv1.ServiceMonitorList{})).To(Succeed())
	Expect(framework.AddToFrameworkScheme(monitoringv1.AddToScheme, &monitoringv1.PrometheusRuleList{})).To(Succeed())
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
