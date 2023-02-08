package test

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"text/template"
	"time"

	"github.com/pkg/errors"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	conditionsv1 "github.com/openshift/custom-resource-status/conditions/v1"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/components"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/network"
	. "github.com/kubevirt/cluster-network-addons-operator/test/check"
	testenv "github.com/kubevirt/cluster-network-addons-operator/test/env"
	"github.com/kubevirt/cluster-network-addons-operator/test/kubectl"
	. "github.com/kubevirt/cluster-network-addons-operator/test/operations"
)

var _ = Describe("NetworkAddonsConfig", func() {
	gvk := GetCnaoV1GroupVersionKind()
	Context("when there is no pre-existing Config", func() {
		DescribeTable("should succeed deploying single component",
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
				NMStateComponent.ComponentName,
				cnao.NetworkAddonsConfigSpec{
					NMState: &cnao.NMState{},
				},
				[]Component{NMStateComponent},
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
				NMStateComponent,
				OvsComponent,
				MacvtapComponent,
			}
			configSpec := cnao.NetworkAddonsConfigSpec{
				KubeMacPool: &cnao.KubeMacPool{},
				LinuxBridge: &cnao.LinuxBridge{},
				Multus:      &cnao.Multus{},
				NMState:     &cnao.NMState{},
				Ovs:         &cnao.Ovs{},
				MacvtapCni:  &cnao.MacvtapCni{},
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

			// Add NMState component
			configSpec.NMState = &cnao.NMState{}
			components = append(components, NMStateComponent)
			testConfigUpdate(gvk, configSpec, components)

			// Add Ovs component
			configSpec.Ovs = &cnao.Ovs{}
			components = append(components, OvsComponent)
			testConfigUpdate(gvk, configSpec, components)

			// Add Macvtap component
			configSpec.MacvtapCni = &cnao.MacvtapCni{}
			components = append(components, MacvtapComponent)
			testConfigUpdate(gvk, configSpec, components)
		})
		Context("and workload PlacementConfiguration is deployed on components", func() {
			components := []Component{
				MacvtapComponent,
				OvsComponent,
				LinuxBridgeComponent,
				MultusComponent,
			}
			configSpec := cnao.NetworkAddonsConfigSpec{
				LinuxBridge:            &cnao.LinuxBridge{},
				Multus:                 &cnao.Multus{},
				Ovs:                    &cnao.Ovs{},
				MacvtapCni:             &cnao.MacvtapCni{},
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
		Context("and SelfSignConfiguration is deployed on components", func() {
			components := []Component{
				KubeMacPoolComponent,
				NMStateComponent,
			}
			configSpec := cnao.NetworkAddonsConfigSpec{
				KubeMacPool: &cnao.KubeMacPool{},
				NMState:     &cnao.NMState{},
			}
			checkSelfSignConfigurationOnComponents := func(expectedSelfSignConfiguration *cnao.SelfSignConfiguration) {
				for _, deploymentName := range []string{NMStateComponent.Deployments[1], KubeMacPoolComponent.Deployments[1]} {
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
			NMStateComponent,
			KubeMacPoolComponent,
			OvsComponent,
			MacvtapComponent,
			MonitoringComponent,
		}
		configSpec := cnao.NetworkAddonsConfigSpec{
			LinuxBridge: &cnao.LinuxBridge{},
			Multus:      &cnao.Multus{},
			NMState:     &cnao.NMState{},
			KubeMacPool: &cnao.KubeMacPool{},
			Ovs:         &cnao.Ovs{},
			MacvtapCni:  &cnao.MacvtapCni{},
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
	Context("when deploying components and checking CNAO prometheus endpoint", func() {
		type prometheusScrapeParams struct {
			configSpec             cnao.NetworkAddonsConfigSpec
			expectedMetricValueMap map[string]string
		}
		DescribeTable("and checking scraped data",
			func(p prometheusScrapeParams) {
				By("deploying the configured NetworkAddonsConfigSpec")
				CreateConfig(gvk, p.configSpec)
				CheckConfigCondition(gvk, ConditionAvailable, ConditionTrue, 15*time.Minute, CheckDoNotRepeat)

				Eventually(func() error {
					By("scraping the monitoring endpoint")
					scrapedData, err := GetScrapedDataFromMonitoringEndpoint()
					Expect(err).ToNot(HaveOccurred())

					By("comparing the scraped Data to the expected metrics' values")
					for metricName, expectedValue := range p.expectedMetricValueMap {
						metricEntry := FindMetric(scrapedData, metricName)
						Expect(metricEntry).ToNot(BeEmpty(), fmt.Sprintf("metric %s does not appear in endpoint scrape", metricName))

						if metricEntry != fmt.Sprintf("%s %s", metricName, expectedValue) {
							return fmt.Errorf("metric %s does not have the expected value %s", metricName, expectedValue)
						}
					}
					return nil
				}, 3*time.Minute, time.Minute).Should(Succeed(), "Should scrape the correct metrics")
			},
			Entry("should report the expected metrics when deploying all components", prometheusScrapeParams{
				configSpec: cnao.NetworkAddonsConfigSpec{
					LinuxBridge: &cnao.LinuxBridge{},
					Multus:      &cnao.Multus{},
					NMState:     &cnao.NMState{},
					KubeMacPool: &cnao.KubeMacPool{},
					Ovs:         &cnao.Ovs{},
					MacvtapCni:  &cnao.MacvtapCni{},
				},
				expectedMetricValueMap: map[string]string{
					"kubevirt_cnao_cr_ready":                "1",
					"kubevirt_cnao_cr_kubemacpool_deployed": "1",
				},
			}),
			Entry("should report the expected metrics when deploying all components but kubemacpool", prometheusScrapeParams{
				configSpec: cnao.NetworkAddonsConfigSpec{
					LinuxBridge: &cnao.LinuxBridge{},
					Multus:      &cnao.Multus{},
					NMState:     &cnao.NMState{},
					Ovs:         &cnao.Ovs{},
					MacvtapCni:  &cnao.MacvtapCni{},
				},
				expectedMetricValueMap: map[string]string{
					"kubevirt_cnao_cr_ready":                "1",
					"kubevirt_cnao_cr_kubemacpool_deployed": "0",
				},
			}),
		)
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
	Describe("NMState", func() {
		var (
			configSpec cnao.NetworkAddonsConfigSpec
		)
		BeforeEach(func() {
			configSpec = cnao.NetworkAddonsConfigSpec{
				NMState: &cnao.NMState{},
			}
		})
		Context("with nmstate-operator installed", func() {
			JustBeforeEach(func() {
				// Install nmstate-operator here
				installNMStateOperator()
				CheckNMStateOperatorIsReady(5 * time.Minute)

				CreateConfig(gvk, configSpec)
				CheckConfigCondition(gvk, ConditionAvailable, ConditionTrue, 15*time.Minute, CheckDoNotRepeat)
			})
			JustAfterEach(func() {
				uninstallNMStateOperator()
			})
			Context("when it is already deployed", func() {
				It("should run nmstate from the operator", func() {
					By("checking for NMState in nmstate namespace")
					Eventually(func() error {
						nmstateHandlerDaemonSet := &v1.DaemonSet{}
						return testenv.Client.Get(context.TODO(), types.NamespacedName{Name: NMStateComponent.DaemonSets[0], Namespace: "nmstate"}, nmstateHandlerDaemonSet)
					}, 5*time.Minute, time.Second).Should(BeNil(), "Timed out waiting for nmstate-operator daemonset")
				})

				Context("when NetworkAddonsConfig is then updated", func() {
					BeforeEach(func() {
						configSpec = cnao.NetworkAddonsConfigSpec{
							NMState: &cnao.NMState{},
							SelfSignConfiguration: &cnao.SelfSignConfiguration{
								CARotateInterval:    "42h0m0s",
								CAOverlapInterval:   "42h0m0s",
								CertRotateInterval:  "42h0m0s",
								CertOverlapInterval: "42h0m0s",
							},
						}
					})
					It("shouldn't update NMState CR", func() {
						By("checking for NMState in nmstate namespace")
						CheckNMStateOperatorIsReady(5 * time.Minute)

						By("checking Nmstate is created with SelfSignConfiguration from NetworkAddonsConfig")
						nmstate, err := kubectl.Kubectl("get", "nmstate", "nmstate", "-n", "nmstate", "-o", "yaml")
						Expect(err).NotTo(HaveOccurred())
						Expect(nmstate).To(ContainSubstring("caOverlapInterval: 42h0m0s"))
						Expect(nmstate).To(ContainSubstring("caRotateInterval: 42h0m0s"))
						Expect(nmstate).To(ContainSubstring("certOverlapInterval: 42h0m0s"))
						Expect(nmstate).To(ContainSubstring("certRotateInterval: 42h0m0s"))

						By("By updating SelfSignConfiguration configuration in networkAddonsConfig")
						configSpec = cnao.NetworkAddonsConfigSpec{
							NMState: &cnao.NMState{},
							SelfSignConfiguration: &cnao.SelfSignConfiguration{
								CARotateInterval:    "24h0m0s",
								CAOverlapInterval:   "24h0m0s",
								CertRotateInterval:  "24h0m0s",
								CertOverlapInterval: "24h0m0s",
							},
						}
						UpdateConfig(gvk, configSpec)
						CheckConfigCondition(gvk, ConditionAvailable, ConditionTrue, 15*time.Minute, CheckDoNotRepeat)

						By("checking NMState still has initial configuration")
						nmstate, err = kubectl.Kubectl("get", "nmstate", "nmstate", "-n", "nmstate", "-o", "yaml")
						Expect(err).NotTo(HaveOccurred())
						Expect(nmstate).To(ContainSubstring("caOverlapInterval: 42h0m0s"))
						Expect(nmstate).To(ContainSubstring("caRotateInterval: 42h0m0s"))
						Expect(nmstate).To(ContainSubstring("certOverlapInterval: 42h0m0s"))
						Expect(nmstate).To(ContainSubstring("certRotateInterval: 42h0m0s"))
					})
				})
			})
			Context("when it is not already deployed", func() {
				BeforeEach(func() {
					configSpec = cnao.NetworkAddonsConfigSpec{}
				})
				It("should run nmstate from the operator", func() {
					configSpec = cnao.NetworkAddonsConfigSpec{
						NMState: &cnao.NMState{},
					}
					UpdateConfig(gvk, configSpec)
					CheckConfigCondition(gvk, ConditionAvailable, ConditionTrue, 15*time.Minute, CheckDoNotRepeat)
					By("checking for NMState in nmstate namespace")
					Eventually(func() error {
						nmstateHandlerDaemonSet := &v1.DaemonSet{}
						return testenv.Client.Get(context.TODO(), types.NamespacedName{Name: NMStateComponent.DaemonSets[0], Namespace: "nmstate"}, nmstateHandlerDaemonSet)
					}, 5*time.Minute, time.Second).Should(BeNil(), "Timed out waiting for nmstate-operator daemonset")
				})
			})
		})
		Context("without nmstate-operator pre-installed", func() {
			BeforeEach(func() {
				By("Deploying Nmstate")
				config := cnao.NetworkAddonsConfigSpec{NMState: &cnao.NMState{}}
				CreateConfig(gvk, config)
			})
			It("should deploy nmstate via CNAO", func() {
				CheckConfigCondition(gvk, ConditionAvailable, ConditionTrue, 15*time.Minute, CheckDoNotRepeat)
				CheckConfigCondition(gvk, ConditionUpgradeable, ConditionFalse, 15*time.Minute, CheckDoNotRepeat)
			})
			Context("when nmstate-operator is then installed", func() {
				checkUpgradeableConditionNotSet := func() {
					By("Checking that Upgradeable condition is not set")
					confStatus := GetConfigStatus(gvk)
					Expect(confStatus).NotTo(BeNil())
					for _, condition := range confStatus.Conditions {
						Expect(condition.Type).NotTo(Equal(conditionsv1.ConditionType(ConditionUpgradeable)))
					}
				}
				BeforeEach(func() {
					installNMStateOperator()
					CheckNMStateOperatorIsReady(5 * time.Minute)
				})
				AfterEach(func() {
					uninstallNMStateOperator()
				})
				It("should switch nmstate from CNAO deployment to nmstate-operator deployment", func() {
					By("checking for NMState in CNAO namespace")
					cnaoNmstateNotFound := func() bool {
						nmstateHandlerDaemonSet := &v1.DaemonSet{}
						errHandler := testenv.Client.Get(context.TODO(), types.NamespacedName{Name: NMStateComponent.DaemonSets[0], Namespace: "cluster-network-addons"}, nmstateHandlerDaemonSet)
						nmstateWebhookPDB := &policyv1.PodDisruptionBudget{}
						errWebhookPDB := testenv.Client.Get(context.TODO(), types.NamespacedName{Name: "nmstate-webhook", Namespace: "cluster-network-addons"}, nmstateWebhookPDB)
						nmstateWebhookService := &corev1.Service{}
						errWebhookService := testenv.Client.Get(context.TODO(), types.NamespacedName{Name: "nmstate-webhook", Namespace: "cluster-network-addons"}, nmstateWebhookService)
						return apierrors.IsNotFound(errHandler) && apierrors.IsNotFound(errWebhookPDB) && apierrors.IsNotFound(errWebhookService)
					}
					Eventually(func() bool {
						return cnaoNmstateNotFound()
					}, 5*time.Minute, time.Second).Should(BeTrue(), "Timed out waiting for CNAO nmstate deployment to be removed")

					By("checking for NMState in nmstate namespace")
					nmstateOperatorHandlersReady := func() bool {
						nmstateHandlerDaemonSet := &v1.DaemonSet{}
						err := testenv.Client.Get(context.TODO(), types.NamespacedName{Name: NMStateComponent.DaemonSets[0], Namespace: "nmstate"}, nmstateHandlerDaemonSet)
						if err != nil {
							return false
						}
						return nmstateHandlerDaemonSet.Status.DesiredNumberScheduled == nmstateHandlerDaemonSet.Status.NumberReady
					}
					Eventually(func() bool {
						return nmstateOperatorHandlersReady()
					}, 5*time.Minute, time.Second).Should(BeTrue(), "Timed out waiting for nmstate-operator daemonset")
					CheckNMStateOperatorIsReady(5 * time.Minute)
					By("checking Nmstate is not owned by CNAO")
					nmstate, err := kubectl.Kubectl("get", "nmstate", "nmstate", "-n", "nmstate", "-o", "yaml")
					Expect(err).NotTo(HaveOccurred())
					Expect(nmstate).NotTo(ContainSubstring("ownerReferences"))

					checkUpgradeableConditionNotSet()
				})
			})
		})
		Context("when cluser-reader exists", func() {
			const (
				clusterReaderRoleName = "cluster-reader"
				testUserNamespace     = "default"
				serviceAccountName    = "test-user"
				testUserBindingName   = "test-user-binding"
			)

			var clusterReaderCreated bool

			JustBeforeEach(func() {
				CreateConfig(gvk, configSpec)
				CheckConfigCondition(gvk, ConditionAvailable, ConditionTrue, 15*time.Minute, CheckDoNotRepeat)
			})

			BeforeEach(func() {
				err := createClusterReaderCR(clusterReaderRoleName)
				Expect(err).To(SatisfyAny(Succeed(), WithTransform(apierrors.IsAlreadyExists, BeTrue())))
				if err == nil {
					clusterReaderCreated = true
				}

				Expect(createTestUserSA(testUserNamespace, serviceAccountName)).To(Succeed(),
					"should success creating a serviceaccount")
				Expect(createTestUserCRB(testUserBindingName, testUserNamespace, serviceAccountName, clusterReaderRoleName)).To(Succeed(),
					"should success creating a clusterrolebinding")
			})

			AfterEach(func() {
				Expect(deleteTestUserCRB(testUserBindingName)).To(Succeed())
			})
			AfterEach(func() {
				Expect(deleteTestUserSA(testUserNamespace, serviceAccountName)).To(Succeed())
			})
			AfterEach(func() {
				if clusterReaderCreated {
					Expect(deleteClusterReaderCR(clusterReaderRoleName)).To(Succeed())
				}
			})

			It("should be able to read NMState resources with cluster-reader user", func() {
				Eventually(func() error {
					_, err := kubectl.Kubectl("get", "nns", fmt.Sprintf("--as=system:serviceaccount:%s:%s", testUserNamespace, serviceAccountName))
					return err
				}, 10*time.Second, time.Second).Should(Succeed())
			})
		})
	})
})

func installNMStateOperator() {
	By("Installing kubernetes-nmstate-operator")
	componentSource, err := GetComponentSource("nmstate")
	Expect(err).ToNot(HaveOccurred(), "Error getting the component source")

	result, err := kubectl.Kubectl("apply", "-f", fmt.Sprintf("https://raw.githubusercontent.com/nmstate/kubernetes-nmstate/%s/deploy/crds/nmstate.io_nmstates.yaml", componentSource.Metadata))
	Expect(err).ToNot(HaveOccurred(), "Error applying CRD: %s", result)

	// Create temp directory
	tmpdir, err := ioutil.TempDir("", "operator-test")
	Expect(err).ToNot(HaveOccurred(), "Error creating temporary dir")
	manifests := []string{"namespace", "service_account", "role", "role_binding", "operator"}
	for _, manifest := range manifests {
		yamlString, err := parseManifest(fmt.Sprintf("https://raw.githubusercontent.com/nmstate/kubernetes-nmstate/%s/deploy/operator/%s.yaml", componentSource.Metadata, manifest), componentSource.Metadata)
		Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("Error parsing manifest file to string: %s", manifest))

		yamlFile := filepath.Join(tmpdir, fmt.Sprintf("%s.yaml", manifest))
		ioutil.WriteFile(yamlFile, []byte(yamlString), 0666)
		result, err = kubectl.Kubectl("apply", "-f", yamlFile)
		Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("Error when running kubectl: %s", result))
	}
}

func uninstallNMStateOperator() {
	By("Uninstalling kubernetes-nmstate-operator")
	componentSource, err := GetComponentSource("nmstate")
	Expect(err).ToNot(HaveOccurred(), "Error getting the component source")

	// Create temp directory
	tmpdir, err := ioutil.TempDir("", "operator-test")
	Expect(err).ToNot(HaveOccurred(), "Error creating temporary dir")
	manifests := []string{"operator", "role_binding", "role", "service_account", "namespace"}
	for _, manifest := range manifests {
		yamlString, err := parseManifest(fmt.Sprintf("https://raw.githubusercontent.com/nmstate/kubernetes-nmstate/%s/deploy/operator/%s.yaml", componentSource.Metadata, manifest), componentSource.Metadata)
		Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("Error parsing manifest file to string: %s", manifest))

		yamlFile := filepath.Join(tmpdir, fmt.Sprintf("%s.yaml", manifest))
		ioutil.WriteFile(yamlFile, []byte(yamlString), 0666)
		result, err := kubectl.Kubectl("delete", "-f", yamlFile)
		Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("Error when running kubectl: %s", result))
	}
	result, err := kubectl.Kubectl("delete", "-f", fmt.Sprintf("https://raw.githubusercontent.com/nmstate/kubernetes-nmstate/%s/deploy/crds/nmstate.io_nmstates.yaml", componentSource.Metadata))
	Expect(err).ToNot(HaveOccurred(), "Error deleting CRD: %s", result)
}

func parseManifest(url string, tag string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", errors.Wrapf(err, "Could not get url: %s", url)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrapf(err, "Error reading body of url: %s", url)
	}
	var tmpl *template.Template
	tmpl = template.Must(template.New("manifest").Parse(string(body)))

	data := struct {
		OperatorNamespace  string
		OperatorImage      string
		OperatorPullPolicy string
		HandlerNamespace   string
		HandlerImage       string
		HandlerPullPolicy  string
	}{
		OperatorNamespace:  "nmstate",
		OperatorImage:      fmt.Sprintf("quay.io/nmstate/kubernetes-nmstate-operator:%s", tag),
		OperatorPullPolicy: "Always",
		HandlerNamespace:   "nmstate",
		HandlerImage:       fmt.Sprintf("quay.io/nmstate/kubernetes-nmstate-handler:%s", tag),
		HandlerPullPolicy:  "Always",
	}
	var out bytes.Buffer
	err = tmpl.Execute(&out, data)
	if err != nil {
		return "", errors.Wrapf(err, "Error parsing template")
	}
	return out.String(), nil
}

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

func createClusterReaderCR(clusterReaderRoleName string) error {
	clusterReader := rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRole",
			APIVersion: rbacv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterReaderRoleName,
		},
		AggregationRule: &rbacv1.AggregationRule{
			ClusterRoleSelectors: []metav1.LabelSelector{
				{
					MatchLabels: map[string]string{"rbac.authorization.k8s.io/aggregate-to-cluster-reader": "true"},
				},
			},
		},
	}
	return testenv.Client.Create(context.TODO(), &clusterReader)
}

func createTestUserSA(testUserNamespace, serviceAccountName string) error {
	testUserSA := corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testUserNamespace,
			Name:      serviceAccountName,
		},
	}
	return testenv.Client.Create(context.TODO(), &testUserSA)
}

func createTestUserCRB(testUserBindingName, testUserNamespace, serviceAccountName, clusterReaderRoleName string) error {
	testUserBinding := rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: rbacv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: testUserBindingName,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Namespace: testUserNamespace,
				Name:      serviceAccountName,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     clusterReaderRoleName,
			APIGroup: rbacv1.GroupName,
		},
	}
	return testenv.Client.Create(context.TODO(), &testUserBinding)
}

func deleteClusterReaderCR(clusterReaderRoleName string) error {
	clusterReader := rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterReaderRoleName,
		},
	}
	return testenv.Client.Delete(context.TODO(), &clusterReader)
}

func deleteTestUserSA(testUserNamespace, serviceAccountName string) error {
	testUserSA := corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testUserNamespace,
			Name:      serviceAccountName,
		},
	}
	return testenv.Client.Delete(context.TODO(), &testUserSA)
}

func deleteTestUserCRB(testUserBindingName string) error {
	testUserBinding := rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: testUserBindingName,
		},
	}
	return testenv.Client.Delete(context.TODO(), &testUserBinding)
}
