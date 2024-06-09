package test

import (
	"context"
	"fmt"
	"time"

	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/components"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/network"
	. "github.com/kubevirt/cluster-network-addons-operator/test/check"
	testenv "github.com/kubevirt/cluster-network-addons-operator/test/env"
	. "github.com/kubevirt/cluster-network-addons-operator/test/operations"
)

var _ = Describe("NetworkAddonsConfig", func() {
	gvk := GetCnaoV1GroupVersionKind()
	Context("when there is no pre-existing Config", func() {
		DescribeTable("should succeed deploying selected components",
			func(configSpec cnao.NetworkAddonsConfigSpec, components []Component) {
				testConfigCreate(gvk, configSpec, components)

				// Make sure that deployed components remain up and running
				CheckConfigCondition(gvk, ConditionAvailable, ConditionTrue, CheckImmediately, time.Minute)
			},
			Entry(
				"Empty config",
				cnao.NetworkAddonsConfigSpec{},
				[]Component{},
			),
			Entry(
				LinuxBridgeComponent.ComponentName,
				cnao.NetworkAddonsConfigSpec{
					LinuxBridge: &cnao.LinuxBridge{},
				},
				[]Component{LinuxBridgeComponent},
			), //2303
			Entry(
				MultusComponent.ComponentName,
				cnao.NetworkAddonsConfigSpec{
					Multus: &cnao.Multus{},
				},
				[]Component{MultusComponent},
			),
			Entry(
				KubeMacPoolComponent.ComponentName,
				cnao.NetworkAddonsConfigSpec{
					KubeMacPool: &cnao.KubeMacPool{},
				},
				[]Component{KubeMacPoolComponent},
			),
			Entry(
				OvsComponent.ComponentName,
				cnao.NetworkAddonsConfigSpec{
					Ovs: &cnao.Ovs{},
				},
				[]Component{OvsComponent},
			),
			Entry(
				MacvtapComponent.ComponentName,
				cnao.NetworkAddonsConfigSpec{
					MacvtapCni: &cnao.MacvtapCni{},
				},
				[]Component{MacvtapComponent},
			),
			Entry(
				"Multus Dynamic Networks and dependencies",
				cnao.NetworkAddonsConfigSpec{
					Multus:                &cnao.Multus{},
					MultusDynamicNetworks: &cnao.MultusDynamicNetworks{},
				},
				[]Component{MultusComponent, MultusDynamicNetworks},
			),
			Entry(
				KubeSecondaryDNSComponent.ComponentName,
				cnao.NetworkAddonsConfigSpec{
					KubeSecondaryDNS: &cnao.KubeSecondaryDNS{},
				},
				[]Component{KubeSecondaryDNSComponent},
			),
			Entry(
				KubevirtIpamController.ComponentName,
				cnao.NetworkAddonsConfigSpec{
					KubevirtIpamController: &cnao.KubevirtIpamController{},
				},
				[]Component{KubevirtIpamController},
			),
		)
		It("should deploy prometheus if NetworkAddonsConfigSpec is not empty", func() {
			testConfigCreate(gvk, cnao.NetworkAddonsConfigSpec{MacvtapCni: &cnao.MacvtapCni{}}, []Component{MacvtapComponent, MonitoringComponent})
		})
		//2264
		It("should be able to deploy all components at once", func() {
			components := []Component{
				MultusComponent,
				LinuxBridgeComponent,
				KubeMacPoolComponent,
				OvsComponent,
				MacvtapComponent,
				MultusDynamicNetworks,
				KubeSecondaryDNSComponent,
				KubevirtIpamController,
			}
			configSpec := cnao.NetworkAddonsConfigSpec{
				KubeMacPool:            &cnao.KubeMacPool{},
				LinuxBridge:            &cnao.LinuxBridge{},
				Multus:                 &cnao.Multus{},
				Ovs:                    &cnao.Ovs{},
				MacvtapCni:             &cnao.MacvtapCni{},
				MultusDynamicNetworks:  &cnao.MultusDynamicNetworks{},
				KubeSecondaryDNS:       &cnao.KubeSecondaryDNS{},
				KubevirtIpamController: &cnao.KubevirtIpamController{},
			}
			testConfigCreate(gvk, configSpec, components)
		})
		//2304
		It("should be able to deploy all components one by one", func() {
			configSpec := cnao.NetworkAddonsConfigSpec{}
			components := []Component{}

			// Deploy initial empty config
			testConfigCreate(gvk, configSpec, components)

			// Deploy Multus component
			configSpec.Multus = &cnao.Multus{}
			components = append(components, MultusComponent)
			testConfigUpdate(gvk, configSpec, components)
			CheckModifiedEvent(gvk)
			CheckProgressingEvent(gvk)

			// Add Linux bridge component
			configSpec.LinuxBridge = &cnao.LinuxBridge{}
			components = append(components, LinuxBridgeComponent)
			testConfigUpdate(gvk, configSpec, components)

			// Add KubeMacPool component
			configSpec.KubeMacPool = &cnao.KubeMacPool{}
			components = append(components, KubeMacPoolComponent)
			testConfigUpdate(gvk, configSpec, components)

			// Add Ovs component
			configSpec.Ovs = &cnao.Ovs{}
			components = append(components, OvsComponent)
			testConfigUpdate(gvk, configSpec, components)

			// Add Macvtap component
			configSpec.MacvtapCni = &cnao.MacvtapCni{}
			components = append(components, MacvtapComponent)
			testConfigUpdate(gvk, configSpec, components)

			// Add Multus Dynamic Networks component (requires multus ...)
			configSpec.Multus = &cnao.Multus{}
			configSpec.MultusDynamicNetworks = &cnao.MultusDynamicNetworks{}
			components = append(components, MultusComponent, MultusDynamicNetworks)
			testConfigUpdate(gvk, configSpec, components)

			// Add KubeSecondaryDNS component
			configSpec.KubeSecondaryDNS = &cnao.KubeSecondaryDNS{}
			components = append(components, KubeSecondaryDNSComponent)
			testConfigUpdate(gvk, configSpec, components)

			// Add KubevirtIpamController component
			configSpec.KubevirtIpamController = &cnao.KubevirtIpamController{}
			components = append(components, KubevirtIpamController)
			testConfigUpdate(gvk, configSpec, components)
		})
		Context("and workload PlacementConfiguration is deployed on components", func() {
			components := []Component{
				MacvtapComponent,
				OvsComponent,
				LinuxBridgeComponent,
				MultusComponent,
				KubeSecondaryDNSComponent,
				KubevirtIpamController,
			}
			configSpec := cnao.NetworkAddonsConfigSpec{
				LinuxBridge:            &cnao.LinuxBridge{},
				Multus:                 &cnao.Multus{},
				Ovs:                    &cnao.Ovs{},
				MacvtapCni:             &cnao.MacvtapCni{},
				MultusDynamicNetworks:  &cnao.MultusDynamicNetworks{},
				KubeSecondaryDNS:       &cnao.KubeSecondaryDNS{},
				KubevirtIpamController: &cnao.KubevirtIpamController{},
				PlacementConfiguration: &cnao.PlacementConfiguration{},
			}
			checkWorkloadPlacementOnComponents := func(expectedWorkLoadPlacement cnao.Placement) {
				for _, component := range components {
					componentPlacementList, err := PlacementListFromComponentDaemonSets(component)
					Expect(err).ToNot(HaveOccurred(), "Should succeed getting the component Placement config")
					for _, placement := range componentPlacementList {
						Expect(placement).To(Equal(expectedWorkLoadPlacement), fmt.Sprintf("PlacementConfiguration of %s component should equal the default workload PlacementConfiguration", component.ComponentName))
					}
				}
			}

			BeforeEach(func() {
				By("Deploying components with default PlacementConfiguration")
				testConfigCreate(gvk, configSpec, components)

				By("Checking PlacementConfiguration was set on components to default workload PlacementConfiguration")
				checkWorkloadPlacementOnComponents(*network.GetDefaultPlacementConfiguration().Workloads)
			})

			It("should be able to update PlacementConfigurations on components specs", func() {
				configSpec.PlacementConfiguration = &cnao.PlacementConfiguration{
					Workloads: &cnao.Placement{NodeSelector: map[string]string{
						"kubernetes.io/hostname": "node01"},
					},
				}

				By("Re-deploying PlacementConfiguration with different workloads values")
				testConfigUpdate(gvk, configSpec, components)

				By("Checking PlacementConfiguration was set on components to updated workload PlacementConfiguration")
				checkWorkloadPlacementOnComponents(*configSpec.PlacementConfiguration.Workloads)
			})
		})
		Context("and infra PlacementConfiguration is deployed on components", func() {
			components := []Component{
				KubeMacPoolComponent,
			}
			configSpec := cnao.NetworkAddonsConfigSpec{
				KubeMacPool:            &cnao.KubeMacPool{},
				PlacementConfiguration: &cnao.PlacementConfiguration{},
			}
			checkInfraPlacementOnComponents := func(expectedInfraPlacement cnao.Placement) {
				for _, component := range components {
					componentPlacementList, err := PlacementListFromComponentDeployments(component)
					Expect(err).ToNot(HaveOccurred(), "Should succeed getting the component Placement config")
					for _, placement := range componentPlacementList {
						Expect(placement).To(Equal(expectedInfraPlacement), fmt.Sprintf("PlacementConfiguration of %s component should equal the default infra PlacementConfiguration", component.ComponentName))
					}
				}
			}

			BeforeEach(func() {
				By("Deploying components with default PlacementConfiguration")
				testConfigCreate(gvk, configSpec, components)

				By("Checking PlacementConfiguration was set on components to default infra PlacementConfiguration")
				checkInfraPlacementOnComponents(*network.GetDefaultPlacementConfiguration().Infra)
			})

			It("should be able to update infra PlacementConfigurations on components specs", func() {
				configSpec.PlacementConfiguration = &cnao.PlacementConfiguration{
					Infra: &cnao.Placement{NodeSelector: map[string]string{
						"kubernetes.io/hostname": "node01"},
					},
				}

				By("Re-deploying PlacementConfiguration with different infra values")
				testConfigUpdate(gvk, configSpec, components)

				By("Checking PlacementConfiguration was set on components to updated infra PlacementConfiguration")
				checkInfraPlacementOnComponents(*configSpec.PlacementConfiguration.Infra)
			})
		})
		Context("and SelfSignConfiguration is deployed on components", func() {
			components := []Component{
				KubeMacPoolComponent,
			}
			configSpec := cnao.NetworkAddonsConfigSpec{
				KubeMacPool: &cnao.KubeMacPool{},
			}
			checkSelfSignConfigurationOnComponents := func(expectedSelfSignConfiguration *cnao.SelfSignConfiguration) {
				for _, deploymentName := range []string{KubeMacPoolComponent.Deployments[1]} {
					envVars, err := GetEnvVarsFromDeployment(deploymentName)
					Expect(err).ToNot(HaveOccurred(), "Should succeed getting env vars from deployment %s", deploymentName)
					Expect(envVars).To(ContainElement(corev1.EnvVar{Name: "CA_ROTATE_INTERVAL", Value: expectedSelfSignConfiguration.CARotateInterval}))
					Expect(envVars).To(ContainElement(corev1.EnvVar{Name: "CA_OVERLAP_INTERVAL", Value: expectedSelfSignConfiguration.CAOverlapInterval}))
					Expect(envVars).To(ContainElement(corev1.EnvVar{Name: "CERT_ROTATE_INTERVAL", Value: expectedSelfSignConfiguration.CertRotateInterval}))
					Expect(envVars).To(ContainElement(corev1.EnvVar{Name: "CERT_OVERLAP_INTERVAL", Value: expectedSelfSignConfiguration.CertOverlapInterval}))
				}
			}

			BeforeEach(func() {
				By("Deploying components with default SelfSignConfiguration")
				testConfigCreate(gvk, configSpec, components)

				By("Checking cert rotation env vars were set on components according to default SelfSignConfiguration")
				checkSelfSignConfigurationOnComponents(network.DefaultSelfSignConfiguration())
			})

			It("should be able to update SelfSignConfiguration on components specs", func() {
				configSpec.SelfSignConfiguration = network.DefaultSelfSignConfiguration()
				configSpec.SelfSignConfiguration.CARotateInterval = "200h20m2s"

				By("Re-deploying SelfSignConfiguration with different workloads values")
				testConfigUpdate(gvk, configSpec, components)

				By("Checking cert rotation env vars were  set on components to updated SelfSignConfiguration")
				checkSelfSignConfigurationOnComponents(configSpec.SelfSignConfiguration)
			})
		})

	})

	Context("when all components are already deployed", func() {
		components := []Component{
			MultusComponent,
			LinuxBridgeComponent,
			KubeMacPoolComponent,
			OvsComponent,
			MacvtapComponent,
			MonitoringComponent,
			MultusDynamicNetworks,
			KubeSecondaryDNSComponent,
		}
		configSpec := cnao.NetworkAddonsConfigSpec{
			LinuxBridge:           &cnao.LinuxBridge{},
			Multus:                &cnao.Multus{},
			KubeMacPool:           &cnao.KubeMacPool{},
			Ovs:                   &cnao.Ovs{},
			MacvtapCni:            &cnao.MacvtapCni{},
			MultusDynamicNetworks: &cnao.MultusDynamicNetworks{},
			KubeSecondaryDNS:      &cnao.KubeSecondaryDNS{},
		}
		BeforeEach(func() {
			CreateConfig(gvk, configSpec)
			CheckConfigCondition(gvk, ConditionAvailable, ConditionTrue, 15*time.Minute, CheckDoNotRepeat)
		})
		//2305
		It("should remain in Available condition after applying the same config", func() {
			UpdateConfig(gvk, configSpec)
			CheckConfigCondition(gvk, ConditionAvailable, ConditionTrue, CheckImmediately, 30*time.Second)
		})
		//2281
		It("should be able to remove all of them by removing the config", func() {
			DeleteConfig(gvk)
			CheckComponentsRemoval(components)
		})
		//2300
		It("should be able to remove the config and create it again", func() {
			DeleteConfig(gvk)
			//TODO: remove this checking after this [1] issue is resolved
			// [1] https://github.com/kubevirt/cluster-network-addons-operator/issues/394
			CheckComponentsRemoval(components)
			CreateConfig(gvk, configSpec)
			CheckConfigCondition(gvk, ConditionAvailable, ConditionTrue, 15*time.Minute, 30*time.Second)
		})
	})

	//2178
	Context("when kubeMacPool is deployed", func() {
		BeforeEach(func() {
			By("Deploying KubeMacPool")
			config := cnao.NetworkAddonsConfigSpec{KubeMacPool: &cnao.KubeMacPool{}}
			CreateConfig(gvk, config)
			CheckConfigCondition(gvk, ConditionAvailable, ConditionTrue, 15*time.Minute, CheckDoNotRepeat)
		})

		It("should modify the MAC range after being redeployed ", func() {
			oldRangeStart, oldRangeEnd := CheckUnicastAndValidity()
			By("Redeploying KubeMacPool")
			DeleteConfig(gvk)
			CheckComponentsRemoval(AllComponents)

			configSpec := cnao.NetworkAddonsConfigSpec{KubeMacPool: &cnao.KubeMacPool{}}
			CreateConfig(gvk, configSpec)
			CheckConfigCondition(gvk, ConditionAvailable, ConditionTrue, 15*time.Minute, CheckDoNotRepeat)
			rangeStart, rangeEnd := CheckUnicastAndValidity()

			By("Comparing the ranges")
			Expect(rangeStart).ToNot(Equal(oldRangeStart))
			Expect(rangeEnd).ToNot(Equal(oldRangeEnd))
		})
	})
	Context("when Macvtap is deployed", func() {
		Context("with the default device plugin configuration", func() {
			BeforeEach(func() {
				configSpec := cnao.NetworkAddonsConfigSpec{
					MacvtapCni: &cnao.MacvtapCni{},
				}
				CreateConfig(gvk, configSpec)
				CheckConfigCondition(gvk, ConditionAvailable, ConditionTrue, 15*time.Minute, CheckDoNotRepeat)
			})
			Context("and a forbidden day2 change is introduced to Macvtap daemonSet", func() {
				annotationKey := "newDay2Changes"
				BeforeEach(func() {
					By("Setting a new Annotation to the macvtap daemonSet")
					err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
						macvtapDaemonSet := &v1.DaemonSet{}
						err := testenv.Client.Get(context.TODO(), types.NamespacedName{Name: MacvtapComponent.DaemonSets[0], Namespace: components.Namespace}, macvtapDaemonSet)
						Expect(err).ToNot(HaveOccurred(), "should succeed getting the macvtap daemonSet")

						macvtapDaemonSet.Spec.Template.SetAnnotations(map[string]string{annotationKey: ""})
						return testenv.Client.Update(context.TODO(), macvtapDaemonSet)
					})
					Expect(err).ToNot(HaveOccurred(), "should succeed setting a new Annotation to the macvtap daemonSet")
				})

				It("should reconcile the object and remove the new Annotation", func() {
					By("checking that the Annotation eventually removed reconciled out by the CNAO operator")
					Eventually(func() bool {
						macvtapDaemonSet := &v1.DaemonSet{}
						err := testenv.Client.Get(context.TODO(), types.NamespacedName{Name: MacvtapComponent.DaemonSets[0], Namespace: components.Namespace}, macvtapDaemonSet)
						Expect(err).ToNot(HaveOccurred(), "should succeed getting the macvtap daemonSet")

						deamonSetAnnotations := macvtapDaemonSet.Spec.Template.GetAnnotations()
						if _, exist := deamonSetAnnotations[annotationKey]; exist {
							return false
						}
						return true

					}, 2*time.Minute, time.Second).Should(BeTrue(), fmt.Sprintf("Timed out waiting for macvtap daemonset added Annotation to be removed"))
				})
			})
		})

		Context("with a non-default configuration", func() {
			const overriddenConfigMapName = "another-config"

			BeforeEach(func() {
				configSpec := cnao.NetworkAddonsConfigSpec{
					MacvtapCni: &cnao.MacvtapCni{DevicePluginConfig: overriddenConfigMapName},
				}
				CreateConfig(gvk, configSpec)
				CheckConfigCondition(gvk, ConditionAvailable, ConditionFalse, time.Minute, time.Minute)
				CheckConfigCondition(gvk, ConditionProgressing, ConditionTrue, CheckDoNotRepeat, CheckDoNotRepeat)
			})

			It("goes to available once the device plugin configuration is provisioned", func() {
				configMap := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      overriddenConfigMapName,
						Namespace: components.Namespace,
					},
					Data: map[string]string{
						"DP_MACVTAP_CONF": "[]",
					},
				}
				Expect(testenv.KubeClient.CoreV1().ConfigMaps(components.Namespace).Create(context.Background(), configMap, metav1.CreateOptions{}))
				CheckConfigCondition(gvk, ConditionAvailable, ConditionTrue, 2*time.Minute, CheckDoNotRepeat)
			})
		})
	})
})

func testConfigCreate(gvk schema.GroupVersionKind, configSpec cnao.NetworkAddonsConfigSpec, components []Component) {
	checkConfigChange(gvk, components, func() {
		CreateConfig(gvk, configSpec)
	})
}

func testConfigUpdate(gvk schema.GroupVersionKind, configSpec cnao.NetworkAddonsConfigSpec, components []Component) {
	checkConfigChange(gvk, components, func() {
		UpdateConfig(gvk, configSpec)
	})
}

// checkConfigChange verifies that given components transition through
// Progressing to Available state while and after the given callback function is
// executed. We start the monitoring sooner than the callback to ensure we catch
// all transitions from the very beginning.
//
// TODO This should be replaced by a solution based around `Watch` once it is
// available on operator-sdk test framework:
// https://github.com/operator-framework/operator-sdk/issues/2655
func checkConfigChange(gvk schema.GroupVersionKind, components []Component, while func()) {

	// Start the function with a little delay to give the Progressing check a better chance
	// of catching the event
	go func() {
		time.Sleep(time.Second)
		while()
	}()

	if len(components) == 0 {
		// Wait until Available condition is reported. Should be fast when no components are
		// being deployed
		CheckConfigCondition(gvk, ConditionAvailable, ConditionTrue, 5*time.Minute, CheckDoNotRepeat)
	} else {
		// If there are any components to deploy wait until Progressing condition is reported
		CheckConfigCondition(gvk, ConditionProgressing, ConditionTrue, time.Minute, CheckDoNotRepeat)
		// Wait until Available condition is reported. It may take a few minutes the first time
		// we are pulling component images to the Node
		CheckConfigCondition(gvk, ConditionAvailable, ConditionTrue, 15*time.Minute, CheckDoNotRepeat)
		CheckConfigCondition(gvk, ConditionProgressing, ConditionFalse, CheckImmediately, CheckDoNotRepeat)

		// Check that all requested components have been deployed
		CheckComponentsDeployment(components)
	}
}
