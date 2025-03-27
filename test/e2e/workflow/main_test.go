package test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/kubevirt/cluster-network-addons-operator/pkg/components"
	. "github.com/kubevirt/cluster-network-addons-operator/test/check"
	testenv "github.com/kubevirt/cluster-network-addons-operator/test/env"
	. "github.com/kubevirt/cluster-network-addons-operator/test/operations"
	"github.com/kubevirt/cluster-network-addons-operator/test/reporter"
)

var operatorVersion string

var cnaoReporter *reporter.KubernetesCNAOReporter

func TestE2E(t *testing.T) {
	cnaoReporter = reporter.New("_out/e2e/workflow/", components.Namespace)
	cnaoReporter.Cleanup()

	RegisterFailHandler(Fail)
	RunSpecs(t, "workflow Test Suite")
}

var _ = BeforeSuite(func() {

	// Change to root directory some test expect that
	os.Chdir("../../../")

	testenv.Start()

	By("Detecting operator version")
	var err error
	operatorVersion, err = getRunningOperatorVersion()
	Expect(err).ToNot(HaveOccurred())
})

var _ = AfterSuite(func() {
	CheckOperatorPodStability(time.Minute)
})

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

func getRunningOperatorVersion() (string, error) {
	operatorDep := &appsv1.Deployment{}

	err := testenv.Client.Get(context.Background(), types.NamespacedName{Name: components.Name, Namespace: components.Namespace}, operatorDep)
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
