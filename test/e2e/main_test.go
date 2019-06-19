package test

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	f "github.com/operator-framework/operator-sdk/pkg/test"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/kubevirt/cluster-network-addons-operator/pkg/apis"
	opv1alpha1 "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/v1alpha1"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/components"
	. "github.com/kubevirt/cluster-network-addons-operator/test/check"
	. "github.com/kubevirt/cluster-network-addons-operator/test/operations"
)

var operatorVersion string

func TestMain(m *testing.M) {
	f.MainEntry(m)
}

func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2E Test Suite")
}

var _ = BeforeSuite(func() {
	By("Adding custom resource scheme to framework")
	err := framework.AddToFrameworkScheme(apis.AddToScheme, &opv1alpha1.NetworkAddonsConfigList{})
	Expect(err).ToNot(HaveOccurred())

	By("Detecting operator version")
	operatorVersion, err = getRunningOperatorVersion()
	Expect(err).ToNot(HaveOccurred())
})

var _ = AfterEach(func() {
	By("Performing cleanup")
	if GetConfig() != nil {
		DeleteConfig()
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
