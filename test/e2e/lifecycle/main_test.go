package test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	ginkgoreporters "kubevirt.io/qe-tools/pkg/ginkgo-reporters"

	f "github.com/operator-framework/operator-sdk/pkg/test"
	framework "github.com/operator-framework/operator-sdk/pkg/test"

	"github.com/kubevirt/cluster-network-addons-operator/pkg/apis"
	opv1alpha1 "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/v1alpha1"
	. "github.com/kubevirt/cluster-network-addons-operator/test/check"
	. "github.com/kubevirt/cluster-network-addons-operator/test/operations"
	. "github.com/kubevirt/cluster-network-addons-operator/test/releases"
)

func TestMain(m *testing.M) {
	f.MainEntry(m)
}

func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	reporters := make([]Reporter, 0)
	if ginkgoreporters.JunitOutput != "" {
		reporters = append(reporters, ginkgoreporters.NewJunitReporter())
	}
	RunSpecsWithDefaultAndCustomReporters(t, "E2E Lifecycle Test Suite", reporters)
}

var _ = BeforeSuite(func() {
	By("Adding custom resource scheme to framework")
	err := framework.AddToFrameworkScheme(apis.AddToScheme, &opv1alpha1.NetworkAddonsConfigList{})
	Expect(err).ToNot(HaveOccurred())
})

var _ = AfterEach(func() {
	By("Performing cleanup")
	if GetConfig() != nil {
		DeleteConfig()
	}
	CheckComponentsRemoval(AllComponents)
	for _, release := range Releases() {
		UninstallRelease(release)
	}
})
