package test

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	ginkgoreporters "kubevirt.io/qe-tools/pkg/ginkgo-reporters"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	f "github.com/operator-framework/operator-sdk/pkg/test"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/kubevirt/cluster-network-addons-operator/pkg/apis"
	cnaov1 "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/v1"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/components"
	. "github.com/kubevirt/cluster-network-addons-operator/test/check"
	. "github.com/kubevirt/cluster-network-addons-operator/test/operations"
	cnaoreporter "github.com/kubevirt/cluster-network-addons-operator/test/reporter"
)

var operatorVersion string

func TestMain(m *testing.M) {
	f.MainEntry(m)
}

func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	reporters := make([]Reporter, 0)
	reporters = append(reporters, cnaoreporter.New("test_logs/e2e/workflow", components.Namespace))
	if ginkgoreporters.JunitOutput != "" {
		reporters = append(reporters, ginkgoreporters.NewJunitReporter())
	}
	RunSpecsWithDefaultAndCustomReporters(t, "Workflow E2E Test Suite", reporters)

}

var _ = BeforeSuite(func() {
	By("Adding custom resource scheme to framework")
	err := framework.AddToFrameworkScheme(apis.AddToScheme, &cnaov1.NetworkAddonsConfigList{})
	Expect(err).ToNot(HaveOccurred())

	Expect(framework.AddToFrameworkScheme(monitoringv1.AddToScheme, &monitoringv1.ServiceMonitorList{})).To(Succeed())
	Expect(framework.AddToFrameworkScheme(monitoringv1.AddToScheme, &monitoringv1.PrometheusRuleList{})).To(Succeed())

	By("Detecting operator version")
	operatorVersion, err = getRunningOperatorVersion()
	Expect(err).ToNot(HaveOccurred())
})

var _ = AfterSuite(func() {
	CheckOperatorPodStability(time.Minute)
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

func getRunningOperatorVersion() (string, error) {
	operatorDep := &appsv1.Deployment{}

	err := framework.Global.Client.Get(context.Background(), types.NamespacedName{Name: components.Name, Namespace: components.Namespace}, operatorDep)
	if err != nil {
		return "", err
	}

	for _, env := range operatorDep.Spec.Template.Spec.Containers[0].Env {
		if env.Name == "OPERATOR_VERSION" {
			return env.Value, nil
		}
	}

	return "", nil
}
