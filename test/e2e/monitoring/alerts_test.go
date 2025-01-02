package test

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"kubevirt.io/kubevirt/tests/framework/checks"
	"kubevirt.io/kubevirt/tests/testsuite"

	k8sv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8slabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	v1 "kubevirt.io/api/core/v1"
	"kubevirt.io/kubevirt/pkg/libvmi"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/components"
	. "github.com/kubevirt/cluster-network-addons-operator/test/check"
	testenv "github.com/kubevirt/cluster-network-addons-operator/test/env"
	"github.com/kubevirt/cluster-network-addons-operator/test/kubectl"
	. "github.com/kubevirt/cluster-network-addons-operator/test/operations"
)

var _ = Context("Prometheus Alerts", func() {
	var prometheusClient *promClient

	BeforeEach(func() {
		var err error
		sourcePort := 4321 + rand.Intn(6000)
		targetPort := 9090
		By(fmt.Sprintf("issuing a port forwarding command to access prometheus API on port %d", sourcePort))

		prometheusClient = newPromClient(sourcePort, prometheusMonitoringNamespace)
		portForwardCmd, err = kubectl.StartPortForwardCommand(prometheusClient.namespace, "prometheus-k8s", prometheusClient.sourcePort, targetPort)
		Expect(err).ToNot(HaveOccurred())
	})
	AfterEach(func() {
		By("removing the port-forwarding command")
		Expect(kubectl.KillPortForwardCommand(portForwardCmd)).To(Succeed())
	})

	Context("when networkaddonsconfig CR is deployed with all components", func() {
		BeforeEach(func() {
			By("delpoying CNAO CR with all component")
			gvk := GetCnaoV1GroupVersionKind()
			configSpec := cnao.NetworkAddonsConfigSpec{
				LinuxBridge: &cnao.LinuxBridge{},
				Multus:      &cnao.Multus{},
				KubeMacPool: &cnao.KubeMacPool{},
				Ovs:         &cnao.Ovs{},
				MacvtapCni:  &cnao.MacvtapCni{},
			}
			CreateConfig(gvk, configSpec)
			CheckConfigCondition(gvk, ConditionAvailable, ConditionTrue, 15*time.Minute, CheckDoNotRepeat)
		})
		AfterEach(func() {
			By("removing CNAO CR")
			gvk := GetCnaoV1GroupVersionKind()
			if GetConfig(gvk) != nil {
				DeleteConfig(gvk)
			}
		})

		It("should fire no alerts", func() {
			By("waiting for the max amount of time it takes the alert to fire on CNAO")
			time.Sleep(5 * time.Minute)
			By("checking non-existence of alerts")
			prometheusClient.checkNoAlertsFired()
		})
	})

	Context("and cluster-network-addons-operator deploys a faulty Kubemacpool", func() {
		noNodePlacementConf := cnao.PlacementConfiguration{
			Infra: &cnao.Placement{
				NodeSelector: map[string]string{
					"node-role.kubernetes.io/no-node": "",
				},
			},
		}
		BeforeEach(func() {
			By("Deploying Kubemacpool component but with a PlacementConfiguration that will prevent it from scheduling")
			gvk := GetCnaoV1GroupVersionKind()
			configSpec := cnao.NetworkAddonsConfigSpec{
				KubeMacPool:            &cnao.KubeMacPool{},
				PlacementConfiguration: &noNodePlacementConf,
			}
			CreateConfig(gvk, configSpec)
			CheckConfigCondition(gvk, ConditionAvailable, ConditionFalse, 1*time.Minute, 1*time.Minute)
		})
		AfterEach(func() {
			By("removing CNAO CR")
			gvk := GetCnaoV1GroupVersionKind()
			if GetConfig(gvk) != nil {
				DeleteConfig(gvk)
			}
		})

		It("should issue NetworkAddonsConfigNotReady and KubemacpoolDown alerts", func() {
			By("waiting for the amount of time it takes the alerts to fire")
			time.Sleep(5 * time.Minute)
			By("checking existence of alerts")
			prometheusClient.checkForAlert("NetworkAddonsConfigNotReady")
			prometheusClient.checkForAlert("KubemacpoolDown")
		})
	})

	Context("when networkaddonsconfig CR is deployed with one component", func() {
		BeforeEach(func() {
			By("delpoying CNAO CR with at least one component")
			gvk := GetCnaoV1GroupVersionKind()
			configSpec := cnao.NetworkAddonsConfigSpec{
				KubeMacPool: &cnao.KubeMacPool{},
			}
			CreateConfig(gvk, configSpec)
			CheckConfigCondition(gvk, ConditionAvailable, ConditionTrue, 15*time.Minute, CheckDoNotRepeat)
		})
		AfterEach(func() {
			By("removing CNAO CR")
			gvk := GetCnaoV1GroupVersionKind()
			if GetConfig(gvk) != nil {
				DeleteConfig(gvk)
			}
		})

		It("configuration: role-binding should point to the prometheus serviceAccount", func() {
			By("checking the monitoring role-binding points to an existing serviceAccount")
			Expect(checkMonitoringRoleBindingConfig("cluster-network-addons-operator-monitoring", components.Namespace)).To(Succeed(), "check value of MONITORING_NAMESPACE env")
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

		Context("and there are duplicate MACs", func() {
			var err error

			AfterEach(func() {
				By("deleting test namespace")
				err = testenv.Client.Delete(context.Background(), &k8sv1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: testsuite.NamespaceTestDefault}})
				Expect(err).ToNot(HaveOccurred())
			})

			BeforeEach(func() {
				By("creating test namespace that is not managed by kubemacpool (opted-out)")
				namespace := &k8sv1.Namespace{ObjectMeta: metav1.ObjectMeta{
					Name: testsuite.NamespaceTestDefault,
					Labels: map[string]string{
						"mutatevirtualmachines.kubemacpool.io": "ignore",
					},
				}}
				err := testenv.Client.Create(context.Background(), namespace)
				Expect(err).ToNot(HaveOccurred())

				By("creating 2 VMs with a duplicate MAC")
				err = createVirtualMachineWithPrimaryInterfaceMacAddress("00-B0-D0-63-C2-26")
				Expect(err).ToNot(HaveOccurred())
				err = createVirtualMachineWithPrimaryInterfaceMacAddress("00-B0-D0-63-C2-26")
				Expect(err).ToNot(HaveOccurred())

				By("cleaning namespace labels, returning the namespace to managed by kubemacpool")
				err = cleanNamespaceLabels(testsuite.NamespaceTestDefault)
				Expect(err).ToNot(HaveOccurred())

				By("restaring kubemacpool pods")
				restartKubemacpoolPods()
			})

			It("should issue KubeMacPoolDuplicateMacsFound alert", func() {
				By("waiting for the amount of time it takes the alert to fire")
				time.Sleep(5 * time.Minute)

				By("checking existence of alert")
				prometheusClient.checkForAlert("KubeMacPoolDuplicateMacsFound")
			})
		})
	})
})

func newRandomVMI() *v1.VirtualMachineInstance {
	vmi := libvmi.New(
		libvmi.WithInterface(libvmi.InterfaceDeviceWithMasqueradeBinding()),
		libvmi.WithNetwork(v1.DefaultPodNetwork()),
	)
	vmi.ObjectMeta.Namespace = testsuite.NamespaceTestDefault
	vmi.Spec.Domain.Resources.Requests = k8sv1.ResourceList{}

	if checks.IsARM64(testsuite.Arch) {
		// Cirros image need 256M to boot on ARM64,
		vmi.Spec.Domain.Resources.Requests[k8sv1.ResourceMemory] = resource.MustParse("256Mi")
	} else {
		vmi.Spec.Domain.Resources.Requests[k8sv1.ResourceMemory] = resource.MustParse("128Mi")
	}

	return vmi
}

func createVirtualMachineWithPrimaryInterfaceMacAddress(macAddress string) error {
	vmi := newRandomVMI()
	vm := libvmi.NewVirtualMachine(vmi, libvmi.WithRunStrategy(v1.RunStrategyAlways))

	vm.Spec.Template.Spec.Domain.Devices.Interfaces[0].MacAddress = macAddress
	err := testenv.Client.Create(context.Background(), vm)

	return err
}

func cleanNamespaceLabels(namespace string) error {
	nsObject := &k8sv1.Namespace{}
	err := testenv.Client.Get(context.Background(), types.NamespacedName{Name: namespace}, nsObject)
	if err != nil {
		return err
	}

	nsObject.Labels = make(map[string]string)

	return testenv.Client.Update(context.Background(), nsObject)
}

func restartKubemacpoolPods() {
	labelSelector, err := k8slabels.Parse("app=kubemacpool")
	Expect(err).ToNot(HaveOccurred())

	listOptions := k8sclient.ListOptions{
		LabelSelector: labelSelector,
		Namespace:     components.Namespace,
	}

	pods := &k8sv1.PodList{}
	err = testenv.Client.List(context.Background(), pods, &listOptions)
	Expect(err).ToNot(HaveOccurred())
	nPods := len(pods.Items)

	for _, pod := range pods.Items {
		err = testenv.Client.Delete(context.Background(), &pod)
		Expect(err).ToNot(HaveOccurred())
	}

	Eventually(func() error {
		err = testenv.Client.List(context.Background(), pods, &listOptions)
		Expect(err).ToNot(HaveOccurred())

		if len(pods.Items) != nPods {
			return fmt.Errorf("not all pods are up yet")
		}

		for _, pod := range pods.Items {
			if pod.Status.Phase != k8sv1.PodRunning {
				return fmt.Errorf("pod %s is not running", pod.Name)
			}
		}

		return nil
	}, 5*time.Minute, 5*time.Second).Should(Not(HaveOccurred()), "failed to restart kubemacpool pods")
}
