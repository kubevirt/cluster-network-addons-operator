package test

import (
	"math/rand"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/components"
	"github.com/kubevirt/cluster-network-addons-operator/test/kubectl"
	. "github.com/kubevirt/cluster-network-addons-operator/test/check"
	. "github.com/kubevirt/cluster-network-addons-operator/test/operations"
)

var _ = Context("Prometheus Alerts", func() {

	Context("when networkaddonsconfig CR is deployed", func() {
		var prometheusClient *promClient
		configSpec := cnao.NetworkAddonsConfigSpec{
			MacvtapCni: &cnao.MacvtapCni{},
		}

		BeforeEach(func() {
			By("issuing a port forwarding command to access prometheus API")
			var err error
			sourcePort := 4321 + rand.Intn(6000)
			targetPort := 9090

			prometheusClient = newPromClient(sourcePort, prometheusMonitoringNamespace)
			portForwardCmd, err = kubectl.StartPortForwardCommand(prometheusClient.namespace, "prometheus-k8s", prometheusClient.sourcePort, targetPort)
			Expect(err).ToNot(HaveOccurred())
		})
		BeforeEach(func() {
			By("delpoying CNAO CR with at least one component")
			gvk := GetCnaoV1GroupVersionKind()
			CreateConfig(gvk, configSpec)
			CheckConfigCondition(gvk, ConditionAvailable, ConditionTrue, 15*time.Minute, CheckDoNotRepeat)
		})

		Context("and cluster-network-addons-operator deployment has no ready replicas", func() {
			BeforeEach(func() {
				By("setting CNAO operator deployment replicas to 0")
				ScaleDeployment(components.Name, components.Namespace, 0)
			})

			It("should issue CnaoDown alert", func() {
				By("waiting for the amount of time it takes the alert to fire")
				time.Sleep(5 * time.Minute)
				By("checking existence of alert")
				prometheusClient.checkForAlert("CnaoDown")
			})

			AfterEach(func() {
				By("restoring CNAO operator deployment replicas to 1")
				ScaleDeployment(components.Name, components.Namespace, 1)
			})
		})

		AfterEach(func() {
			By("removing CNAO CR")
			gvk := GetCnaoV1GroupVersionKind()
			if GetConfig(gvk) != nil {
				DeleteConfig(gvk)
			}
		})

		AfterEach(func() {
			By("removing the port-forwarding command")
			Expect(kubectl.KillPortForwardCommand(portForwardCmd)).To(Succeed())
		})

	})
})
