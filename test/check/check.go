package check

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	k8snetworkplumbingwgv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	testenv "github.com/kubevirt/cluster-network-addons-operator/test/env"
	conditionsv1 "github.com/openshift/custom-resource-status/conditions/v1"
	securityapi "github.com/openshift/origin/pkg/security/apis/security"
	"gopkg.in/yaml.v2"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	k8slabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
	cnaov1 "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/v1"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/components"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/util/k8s"

	"github.com/kubevirt/cluster-network-addons-operator/pkg/eventemitter"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/names"

	. "github.com/kubevirt/cluster-network-addons-operator/test/kubectl"
	. "github.com/kubevirt/cluster-network-addons-operator/test/operations"
)

const (
	CheckImmediately   = time.Microsecond
	CheckDoNotRepeat   = time.Microsecond
	CheckIgnoreVersion = "IGNORE"
)

func CheckComponentsDeployment(components []Component) {
	for _, component := range components {
		By(fmt.Sprintf("Checking that component %s is deployed", component.ComponentName))
		err := checkForComponent(&component)
		Expect(err).NotTo(HaveOccurred(), "Component has not been fully deployed")
	}
}

func CheckCrdExplainable() {
	By("Checking crd is explainable")
	explain, _, err := Kubectl("explain", "networkaddonsconfigs")
	Expect(err).NotTo(HaveOccurred(), "explain should not return error")

	Expect(explain).ToNot(BeEmpty(), "explain output should not be empty")
	Expect(explain).ToNot(ContainSubstring("<empty>"), "explain output should not contain <empty> fields")
}

func CheckComponentsRemoval(components []Component) {
	for _, component := range components {
		// TODO: Once finalizers are implemented, we should switch to using
		// once-time checks, since after NodeNetworkState removal, no components
		// should be left over.
		By(fmt.Sprintf("Checking that component %s has been removed", component.ComponentName))
		Eventually(func() error {
			return checkForComponentRemoval(&component)
		}, 5*time.Minute, time.Second).ShouldNot(HaveOccurred(), fmt.Sprintf("%s component has not been fully removed within the given timeout", component.ComponentName))
	}

	By(fmt.Sprintf("Checking the %s configmap has been removed", names.AppliedPrefix+names.OperatorConfig))
	Eventually(func() error {
		return checkForConfigMapRemoval(names.AppliedPrefix + names.OperatorConfig)
	}, 5*time.Minute, time.Second).Should(Succeed())
}

func getConfigComponentsMap(gvk schema.GroupVersionKind) (map[string]struct{}, error) {
	config := GetConfig(gvk)
	if config == nil {
		return nil, fmt.Errorf("config not found")
	}
	configV1 := ConvertToConfigV1(config)
	existingComponentsMap := map[string]struct{}{}

	existingComponentsMap[MonitoringComponent.ComponentName] = struct{}{}

	if configV1.Spec.KubeMacPool != nil {
		existingComponentsMap[KubeMacPoolComponent.ComponentName] = struct{}{}
	}
	if configV1.Spec.KubevirtIpamController != nil {
		existingComponentsMap[KubevirtIpamController.ComponentName] = struct{}{}
	}
	if configV1.Spec.KubeSecondaryDNS != nil {
		existingComponentsMap[KubeSecondaryDNSComponent.ComponentName] = struct{}{}
	}
	if configV1.Spec.LinuxBridge != nil {
		existingComponentsMap[LinuxBridgeComponent.ComponentName] = struct{}{}
	}
	if configV1.Spec.MacvtapCni != nil {
		existingComponentsMap[MacvtapComponent.ComponentName] = struct{}{}
	}
	if configV1.Spec.Ovs != nil {
		existingComponentsMap[OvsComponent.ComponentName] = struct{}{}
	}
	if configV1.Spec.Multus != nil {
		existingComponentsMap[MultusComponent.ComponentName] = struct{}{}
	}
	if configV1.Spec.MultusDynamicNetworks != nil {
		existingComponentsMap[MultusDynamicNetworks.ComponentName] = struct{}{}
	}
	return existingComponentsMap, nil
}

func CheckConfigComponents(gvk schema.GroupVersionKind, components []Component) {
	Eventually(func() error {
		deployedComponents, err := getConfigComponentsMap(gvk)
		if err != nil {
			return err
		}
		for _, component := range components {
			if _, exist := deployedComponents[component.ComponentName]; !exist {
				return fmt.Errorf("component %s is not updated in config", component.ComponentName)
			}
		}
		return nil
	}).WithPolling(10 * time.Millisecond).WithTimeout(5 * time.Minute).Should(Succeed())
}

func CheckConfigCondition(gvk schema.GroupVersionKind, conditionType ConditionType, conditionStatus ConditionStatus, timeout time.Duration, duration time.Duration) {
	By(fmt.Sprintf("Checking that condition %q status is set to %s", conditionType, conditionStatus))
	getAndCheckCondition := func() error {

		return checkConfigCondition(gvk, conditionType, conditionStatus)
	}

	if timeout != CheckImmediately {
		Eventually(getAndCheckCondition, timeout, 10*time.Millisecond).ShouldNot(HaveOccurred(), fmt.Sprintf("Timed out waiting for the condition"))
	} else {
		Expect(getAndCheckCondition()).NotTo(HaveOccurred(), fmt.Sprintf("Condition is not in the expected state"))
	}

	if duration != CheckDoNotRepeat {
		Consistently(getAndCheckCondition, duration, 10*time.Millisecond).ShouldNot(HaveOccurred(), fmt.Sprintf("Condition prematurely changed its value"))
	}
}

func CheckConfigConditionChangedAfter(
	gvk schema.GroupVersionKind,
	conditionType ConditionType,
	expectedStatus ConditionStatus,
	checkCondTimestampMap map[conditionsv1.ConditionType]time.Time,
	timeout, duration time.Duration,
) {
	By(fmt.Sprintf("Checking that condition %q status is %q and changed after previous transition time", conditionType, expectedStatus))

	getAndCheck := func() error {
		return checkConfigConditionChangedAfter(gvk, conditionsv1.ConditionType(conditionType), corev1.ConditionStatus(expectedStatus), checkCondTimestampMap)
	}

	if timeout != CheckImmediately {
		Eventually(getAndCheck, timeout, 10*time.Millisecond).ShouldNot(HaveOccurred(), fmt.Sprintf("Timed out waiting for condition %q to change", conditionType))
	} else {
		Expect(getAndCheck()).NotTo(HaveOccurred(), fmt.Sprintf("Condition %q is not in expected state or hasn't changed", conditionType))
	}

	if duration != CheckDoNotRepeat {
		Consistently(getAndCheck, duration, 10*time.Millisecond).ShouldNot(HaveOccurred(), fmt.Sprintf("Condition %q prematurely changed again", conditionType))
	}
}

func checkConfigConditionChangedAfter(
	gvk schema.GroupVersionKind,
	conditionType conditionsv1.ConditionType,
	expectedStatus corev1.ConditionStatus,
	checkCondTimestampMap map[conditionsv1.ConditionType]time.Time,
) error {
	confStatus := GetConfigStatus(gvk)
	if confStatus == nil {
		return fmt.Errorf("Config Status not found")
	}

	for _, cond := range confStatus.Conditions {
		if cond.Type == conditionType {
			if cond.Status != expectedStatus {
				return fmt.Errorf("condition %q is in state %q, expected %q", conditionType, cond.Status, expectedStatus)
			}

			oldTime, found := checkCondTimestampMap[conditionType]
			if found && !cond.LastTransitionTime.Time.After(oldTime) {
				return fmt.Errorf("condition %q has not changed since %s", conditionType, oldTime.Format(time.RFC3339))
			}

			// If not found, treat as new condition
			return nil
		}
	}

	// If expected status is False, it's okay to be missing
	if expectedStatus == corev1.ConditionFalse {
		return nil
	}

	return fmt.Errorf("condition %q not found", conditionType)
}

func CaptureConditionTimestamps(gvk schema.GroupVersionKind) map[conditionsv1.ConditionType]time.Time {
	confStatus := GetConfigStatus(gvk)
	ts := make(map[conditionsv1.ConditionType]time.Time)
	if confStatus != nil {
		for _, cond := range confStatus.Conditions {
			ts[cond.Type] = cond.LastTransitionTime.Time
		}
	}
	return ts
}

func PlacementListFromComponentDaemonSets(component Component) ([]cnao.Placement, error) {
	placementList := []cnao.Placement{}
	for _, daemonSetName := range component.DaemonSets {
		daemonSet := appsv1.DaemonSet{}
		err := testenv.Client.Get(context.Background(), types.NamespacedName{Name: daemonSetName, Namespace: components.Namespace}, &daemonSet)
		if err != nil {
			return placementList, err
		}

		daemonSetPlacement := cnao.Placement{}
		daemonSetPlacement.NodeSelector = daemonSet.Spec.Template.Spec.NodeSelector
		daemonSetPlacement.Affinity = *daemonSet.Spec.Template.Spec.Affinity
		daemonSetPlacement.Tolerations = daemonSet.Spec.Template.Spec.Tolerations

		placementList = append(placementList, daemonSetPlacement)
	}

	return placementList, nil
}

func PlacementListFromComponentDeployments(component Component) ([]cnao.Placement, error) {
	placementList := []cnao.Placement{}
	for _, deploymentName := range component.Deployments {
		deployment := appsv1.Deployment{}
		err := testenv.Client.Get(context.Background(), types.NamespacedName{Name: deploymentName, Namespace: components.Namespace}, &deployment)
		if err != nil {
			return placementList, err
		}

		deploymentPlacement := cnao.Placement{}
		deploymentPlacement.NodeSelector = deployment.Spec.Template.Spec.NodeSelector
		deploymentPlacement.Affinity = *deployment.Spec.Template.Spec.Affinity
		deploymentPlacement.Tolerations = deployment.Spec.Template.Spec.Tolerations

		placementList = append(placementList, deploymentPlacement)
	}

	return placementList, nil
}

func GetEnvVarsFromDeployment(deploymentName string) ([]corev1.EnvVar, error) {
	deployment := appsv1.Deployment{}
	err := testenv.Client.Get(context.Background(), types.NamespacedName{Name: deploymentName, Namespace: components.Namespace}, &deployment)
	if err != nil {
		return nil, err
	}
	envVars := []corev1.EnvVar{}
	for _, container := range deployment.Spec.Template.Spec.Containers {
		envVars = append(envVars, container.Env...)
	}

	return envVars, nil
}

func CheckConfigVersions(gvk schema.GroupVersionKind, operatorVersion, observedVersion, targetVersion string, timeout, duration time.Duration) {
	By(fmt.Sprintf("Checking that status contains expected versions Operator: %q, Observed: %q, Target: %q", operatorVersion, observedVersion, targetVersion))
	getAndCheckVersions := func() error {
		configStatus := GetConfigStatus(gvk)

		errs := []error{}
		errsAppend := func(err error) {
			if err != nil {
				errs = append(errs, err)
			}
		}

		if operatorVersion != CheckIgnoreVersion && configStatus.OperatorVersion != operatorVersion {
			errsAppend(fmt.Errorf("OperatorVersion %q does not match expected %q", configStatus.OperatorVersion, operatorVersion))
		}

		if observedVersion != CheckIgnoreVersion && configStatus.ObservedVersion != observedVersion {
			errsAppend(fmt.Errorf("ObservedVersion %q does not match expected %q", configStatus.ObservedVersion, observedVersion))
		}

		if targetVersion != CheckIgnoreVersion && configStatus.TargetVersion != targetVersion {
			errsAppend(fmt.Errorf("TargetVersion %q does not match expected %q", configStatus.TargetVersion, targetVersion))
		}

		return errsToErr(errs)
	}

	if timeout != CheckImmediately {
		Eventually(getAndCheckVersions, timeout, time.Second).ShouldNot(HaveOccurred(), fmt.Sprintf("Timed out waiting for the expected versions"))
	} else {
		Expect(getAndCheckVersions()).NotTo(HaveOccurred(), fmt.Sprintf("Versions are not in the expected state"))
	}

	if duration != CheckDoNotRepeat {
		Consistently(getAndCheckVersions, duration, time.Second).ShouldNot(HaveOccurred(), fmt.Sprintf("Versions prematurely changed their values"))
	}
}

func CheckOperatorIsReady(timeout time.Duration) {
	By("Checking that the operator is up and running")
	if timeout != CheckImmediately {
		Eventually(func() error {
			return checkForDeployment(components.Name, false, false)
		}, timeout, time.Second).ShouldNot(HaveOccurred(), fmt.Sprintf("Timed out waiting for the operator to become ready"))
	} else {
		Expect(checkForDeployment(components.Name, false, false)).ShouldNot(HaveOccurred(), "Operator is not ready")
	}
}

func CheckForLeftoverObjects(currentVersion string) {
	listOptions := client.ListOptions{}
	key := cnaov1.GroupVersion.Group + "/version"
	labelSelector, err := k8slabels.Parse(fmt.Sprintf("%s,%s != %s", key, key, currentVersion))
	Expect(err).NotTo(HaveOccurred())
	listOptions.LabelSelector = labelSelector

	Eventually(func(g Gomega) {
		deployments := appsv1.DeploymentList{}
		err = testenv.Client.List(context.Background(), &deployments, &listOptions)
		Expect(err).NotTo(HaveOccurred())
		Expect(deployments.Items).To(BeEmpty(), "Found leftover objects from the previous operator version")

		daemonSets := appsv1.DaemonSetList{}
		err = testenv.Client.List(context.Background(), &daemonSets, &listOptions)
		Expect(err).NotTo(HaveOccurred())
		Expect(daemonSets.Items).To(BeEmpty(), "Found leftover objects from the previous operator version")

		configMaps := corev1.ConfigMapList{}
		err = testenv.Client.List(context.Background(), &configMaps, &listOptions)
		Expect(err).NotTo(HaveOccurred())
		Expect(configMaps.Items).To(BeEmpty(), "Found leftover objects from the previous operator version")

		namespaces := corev1.NamespaceList{}
		err = testenv.Client.List(context.Background(), &namespaces, &listOptions)
		Expect(err).NotTo(HaveOccurred())
		Expect(namespaces.Items).To(BeEmpty(), "Found leftover objects from the previous operator version")

		clusterRoles := rbacv1.ClusterRoleList{}
		err = testenv.Client.List(context.Background(), &clusterRoles, &listOptions)
		Expect(err).NotTo(HaveOccurred())
		Expect(clusterRoles.Items).To(BeEmpty(), "Found leftover objects from the previous operator version")

		clusterRoleBindings := rbacv1.ClusterRoleBindingList{}
		err = testenv.Client.List(context.Background(), &clusterRoleBindings, &listOptions)
		Expect(err).NotTo(HaveOccurred())
		Expect(clusterRoleBindings.Items).To(BeEmpty(), "Found leftover objects from the previous operator version")

		roles := rbacv1.RoleList{}
		err = testenv.Client.List(context.Background(), &roles, &listOptions)
		Expect(err).NotTo(HaveOccurred())
		Expect(roles.Items).To(BeEmpty(), "Found leftover objects from the previous operator version")

		roleBindings := rbacv1.RoleBindingList{}
		err = testenv.Client.List(context.Background(), &roleBindings, &listOptions)
		Expect(err).NotTo(HaveOccurred())
		Expect(roleBindings.Items).To(BeEmpty(), "Found leftover objects from the previous operator version")

		serviceAccounts := corev1.ServiceAccountList{}
		err = testenv.Client.List(context.Background(), &serviceAccounts, &listOptions)
		Expect(err).NotTo(HaveOccurred())
		Expect(serviceAccounts.Items).To(BeEmpty(), "Found leftover objects from the previous operator version")

		services := corev1.ServiceList{}
		Expect(testenv.Client.List(context.Background(), &services, &listOptions)).To(Succeed())
		Expect(services.Items).To(BeEmpty(), "Found leftover objects from the previous operator version")

		serviceMonitors := monitoringv1.ServiceMonitorList{}
		Expect(testenv.Client.List(context.Background(), &serviceMonitors, &listOptions)).To(Succeed())
		Expect(serviceMonitors.Items).To(BeEmpty(), "Found leftover objects from the previous operator version")

		prometheusRules := monitoringv1.PrometheusRuleList{}
		Expect(testenv.Client.List(context.Background(), &prometheusRules, &listOptions)).To(Succeed())
		Expect(prometheusRules.Items).To(BeEmpty(), "Found leftover objects from the previous operator version")
	}, 2*time.Minute, time.Second).Should(Succeed())
}

func CheckForLeftoverLabels() {
	namespace := corev1.Namespace{}
	err := testenv.Client.Get(context.Background(), types.NamespacedName{Name: components.Namespace}, &namespace)
	Expect(err).NotTo(HaveOccurred())

	labels := namespace.GetLabels()
	for _, label := range k8s.RemovedLabels() {
		value, found := labels[label]
		Expect(found).To(BeFalse(), fmt.Sprintf("unexpected label %s:%s found", label, value))
	}
}

func KeepCheckingWhile(check func(), while func()) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	done := make(chan bool)

	go func() {
		// Perform some long running operation
		while()

		// Finally close the validator
		close(done)
	}()

	// Keep checking while the goroutine is running
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			check()
		}
	}
}

func checkForComponent(component *Component) error {
	errs := []error{}
	errsAppend := func(err error) {
		if err != nil {
			errs = append(errs, err)
		}
	}

	if component.ClusterRole != "" {
		errsAppend(checkForClusterRole(component.ClusterRole))
	}

	if component.ClusterRoleBinding != "" {
		errsAppend(checkForClusterRoleBinding(component.ClusterRoleBinding))
	}

	if component.SecurityContextConstraints != "" {
		errsAppend(checkForSecurityContextConstraints(component.SecurityContextConstraints))
	}

	for _, daemonSet := range component.DaemonSets {
		errsAppend(checkForDaemonSet(daemonSet))
	}

	for _, deployment := range component.Deployments {
		errsAppend(checkForDeployment(deployment, true, true))
	}

	if component.Secret != "" {
		errsAppend(checkForSecret(component.Secret))
	}

	if component.MutatingWebhookConfiguration != "" {
		errsAppend(checkForMutatingWebhookConfiguration(component.MutatingWebhookConfiguration))
	}

	if component.Service != "" {
		errsAppend(checkForService(component.Service))
	}

	if component.ServiceMonitor != "" {
		errsAppend(checkForServiceMonitor(component.ServiceMonitor))
	}

	if component.PrometheusRule != "" {
		errsAppend(checkForPrometheusRule(component.PrometheusRule))
	}

	if component.NetworkAttachmentDefinition != "" {
		errsAppend(checkForNetworkAttachmentDefinition(component.NetworkAttachmentDefinition))
	}

	if component.ConfigMap != "" {
		errsAppend(checkForConfigMap(component.ConfigMap))
	}

	return errsToErr(errs)
}

func checkForComponentRemoval(component *Component) error {
	errs := []error{}
	errsAppend := func(err error) {
		if err != nil {
			errs = append(errs, err)
		}
	}

	if component.ClusterRole != "" {
		errsAppend(checkForClusterRoleRemoval(component.ClusterRole))
	}

	if component.ClusterRoleBinding != "" {
		errsAppend(checkForClusterRoleBindingRemoval(component.ClusterRoleBinding))
	}

	if component.SecurityContextConstraints != "" {
		errsAppend(checkForSecurityContextConstraintsRemoval(component.SecurityContextConstraints))
	}

	for _, daemonSet := range component.DaemonSets {
		errsAppend(checkForDaemonSetRemoval(daemonSet))
	}

	for _, deployment := range component.Deployments {
		errsAppend(checkForDeploymentRemoval(deployment))
	}

	if component.Secret != "" {
		errsAppend(checkForSecretRemoval(component.Secret))
	}

	if component.MutatingWebhookConfiguration != "" {
		errsAppend(checkForMutatingWebhookConfigurationRemoval(component.MutatingWebhookConfiguration))
	}

	if component.Service != "" {
		errsAppend(checkForServiceRemoval(component.Service))
	}

	if component.ServiceMonitor != "" {
		errsAppend(checkForServiceMonitorRemoval(component.ServiceMonitor))
	}

	if component.PrometheusRule != "" {
		errsAppend(checkForPrometheusRuleRemoval(component.PrometheusRule))
	}

	if component.NetworkAttachmentDefinition != "" {
		errsAppend(checkForNetworkAttachmentDefinitionRemoval(component.NetworkAttachmentDefinition))
	}

	if component.ConfigMap != "" {
		errsAppend(checkForConfigMapRemoval(component.ConfigMap))
	}

	return errsToErr(errs)
}

func errsToErr(errs []error) error {
	if len(errs) == 0 {
		return nil
	}
	errsStrings := []string{}
	for _, err := range errs {
		errsStrings = append(errsStrings, err.Error())
	}
	return errors.New(strings.Join(errsStrings, "\n"))
}

func checkForClusterRole(name string) error {
	clusterRole := rbacv1.ClusterRole{}
	err := testenv.Client.Get(context.Background(), types.NamespacedName{Name: name}, &clusterRole)
	if err != nil {
		return err
	}

	err = checkRelationshipLabels(clusterRole.GetLabels(), "ClusterRole", name)
	if err != nil {
		return err
	}

	return nil
}

func checkForClusterRoleBinding(name string) error {
	clusterRoleBinding := rbacv1.ClusterRoleBinding{}
	err := testenv.Client.Get(context.Background(), types.NamespacedName{Name: name}, &clusterRoleBinding)
	if err != nil {
		return err
	}

	err = checkRelationshipLabels(clusterRoleBinding.GetLabels(), "ClusterRoleBinding", name)
	if err != nil {
		return err
	}

	return nil
}

func checkForSecurityContextConstraints(name string) error {
	scc := securityapi.SecurityContextConstraints{}
	err := testenv.Client.Get(context.Background(), types.NamespacedName{Name: name}, &scc)
	if isNotSupportedKind(err) {
		return nil
	}
	if err != nil {
		return err
	}

	err = checkRelationshipLabels(scc.GetLabels(), "SecurityContextConstraint", name)
	if err != nil {
		return err
	}

	return nil
}

func checkForDeployment(name string, checkVersionLabels, checkRelationshipLabels bool) error {
	return CheckForGenericDeployment(name, components.Namespace, checkVersionLabels, checkRelationshipLabels)
}

func CheckForGenericDeployment(name, namespace string, checkVersionLabels, checkRelationshipLabels bool) error {
	deployment := appsv1.Deployment{}

	err := testenv.Client.Get(context.Background(), types.NamespacedName{Name: name, Namespace: namespace}, &deployment)
	if err != nil {
		return err
	}

	if checkVersionLabels {
		labels := deployment.GetLabels()
		if labels != nil {
			if _, operatorLabelSet := labels[cnaov1.GroupVersion.Group+"/version"]; !operatorLabelSet {
				return fmt.Errorf("Deployment %s/%s is missing operator label", namespace, name)
			}
		} else {
			return fmt.Errorf("Deployment %s/%s has no labels. Should have operator label", namespace, name)
		}
	}

	if checkRelationshipLabels {
		err := checkWorkloadRelationshipLabels([]map[string]string{deployment.GetLabels(), deployment.Spec.Template.GetLabels()}, "Deployment", name)
		if err != nil {
			return err
		}
	}

	if deployment.Status.UnavailableReplicas > 0 || deployment.Status.AvailableReplicas == 0 {
		manifest, err := yaml.Marshal(deployment)
		if err != nil {
			panic(err)
		}
		return fmt.Errorf("Deployment %s/%s is not ready, current state:\n%v", namespace, name, string(manifest))
	}

	return nil
}

func checkWorkloadRelationshipLabels(labelMapList []map[string]string, kind, name string) error {
	for _, labels := range labelMapList {
		if err := checkRelationshipLabels(labels, kind, name); err != nil {
			return err
		}
	}

	return nil
}

// CheckOperatorPodStability makes sure that the CNAO pod has not restarted since it started working
func CheckOperatorPodStability(continuesDuration time.Duration) {
	By("Checking that cnao operator pod hasn't performed any resets")
	if continuesDuration != CheckImmediately {
		Consistently(func() error {
			return CalculateOperatorPodStability()
		}, continuesDuration, time.Second).ShouldNot(HaveOccurred(), "CNAO operator pod should not restart consistently")
	} else {
		Expect(CalculateOperatorPodStability()).ShouldNot(HaveOccurred(), "Operator is not ready")
	}
}

func PrintOperatorPodStability() {
	if err := CalculateOperatorPodStability(); err != nil {
		fmt.Fprintln(GinkgoWriter, "WARNING: CNAO operator pod is not stable: "+err.Error())
	}
	return
}

func CalculateOperatorPodStability() error {
	pods := corev1.PodList{}
	listOptions := []client.ListOption{
		client.MatchingLabels(map[string]string{"name": components.Name}),
		client.InNamespace(components.Namespace),
	}

	err := testenv.Client.List(context.Background(), &pods, listOptions...)
	Expect(err).NotTo(HaveOccurred(), "should succeed getting the cnao pod")

	if len(pods.Items) != 1 {
		return fmt.Errorf("cnao operator pod should only have 1 replica")
	}

	for _, containerStatus := range pods.Items[0].Status.ContainerStatuses {
		if containerStatus.RestartCount != 0 {
			return fmt.Errorf("cnao operator pod restarted %d time since created", containerStatus.RestartCount)
		}
	}
	return nil
}

func checkForDaemonSet(name string) error {
	daemonSet := appsv1.DaemonSet{}

	err := testenv.Client.Get(context.Background(), types.NamespacedName{Name: name, Namespace: components.Namespace}, &daemonSet)
	if err != nil {
		return err
	}

	labels := daemonSet.GetLabels()
	if labels != nil {
		if _, operatorLabelSet := labels[cnaov1.GroupVersion.Group+"/version"]; !operatorLabelSet {
			return fmt.Errorf("DaemonSet %s/%s is missing operator label", components.Namespace, name)
		}
	}

	err = checkWorkloadRelationshipLabels([]map[string]string{daemonSet.GetLabels(), daemonSet.Spec.Template.GetLabels()}, "DaemonSet", name)
	if err != nil {
		return err
	}

	if daemonSet.Status.NumberUnavailable > 0 || (daemonSet.Status.NumberAvailable == 0 && daemonSet.Status.DesiredNumberScheduled != 0) {
		manifest, err := yaml.Marshal(daemonSet)
		if err != nil {
			panic(err)
		}
		return fmt.Errorf("DaemonSet %s/%s is not ready, current state:\n%v", components.Namespace, name, string(manifest))
	}

	return nil
}

func checkForSecret(name string) error {
	secret := corev1.Secret{}
	err := testenv.Client.Get(context.Background(), types.NamespacedName{Name: name, Namespace: components.Namespace}, &secret)
	if err != nil {
		return err
	}

	err = checkRelationshipLabels(secret.GetLabels(), "Secret", name)
	if err != nil {
		return err
	}

	return nil
}

func checkForMutatingWebhookConfiguration(name string) error {
	mutatingWebhookConfig := admissionregistrationv1.MutatingWebhookConfiguration{}
	err := testenv.Client.Get(context.Background(), types.NamespacedName{Name: name}, &mutatingWebhookConfig)
	if err != nil {
		return err
	}

	err = checkRelationshipLabels(mutatingWebhookConfig.GetLabels(), "MutatingWebhookConfiguration", name)
	if err != nil {
		return err
	}

	return nil
}

func checkForService(name string) error {
	service := corev1.Service{}
	err := testenv.Client.Get(context.Background(), types.NamespacedName{Name: name, Namespace: components.Namespace}, &service)
	if err != nil {
		return err
	}

	err = checkRelationshipLabels(service.GetLabels(), "Service", name)
	if err != nil {
		return err
	}

	return nil
}

func checkForServiceMonitor(name string) error {
	serviceMonitor := monitoringv1.ServiceMonitor{}
	err := testenv.Client.Get(context.Background(), types.NamespacedName{Name: name, Namespace: components.Namespace}, &serviceMonitor)
	if err != nil {
		return err
	}

	err = checkRelationshipLabels(serviceMonitor.GetLabels(), "ServiceMonitor", name)
	if err != nil {
		return err
	}

	return nil
}

func checkForPrometheusRule(name string) error {
	prometheusRule := monitoringv1.PrometheusRule{}
	err := testenv.Client.Get(context.Background(), types.NamespacedName{Name: name, Namespace: components.Namespace}, &prometheusRule)
	if err != nil {
		return err
	}

	err = checkRelationshipLabels(prometheusRule.GetLabels(), "PrometheusRule", name)
	if err != nil {
		return err
	}

	return nil
}

func checkForNetworkAttachmentDefinition(name string) error {
	networkAttachmentDefinition := k8snetworkplumbingwgv1.NetworkAttachmentDefinition{}
	err := testenv.Client.Get(context.Background(), types.NamespacedName{Name: name, Namespace: corev1.NamespaceDefault}, &networkAttachmentDefinition)
	if err != nil {
		return err
	}

	err = checkRelationshipLabels(networkAttachmentDefinition.GetLabels(), "NetworkAttachmentDefinition", name)
	if err != nil {
		return err
	}

	return nil
}

func checkForConfigMap(name string) error {
	configMap := corev1.ConfigMap{}
	err := testenv.Client.Get(context.Background(), types.NamespacedName{Name: name, Namespace: components.Namespace}, &configMap)
	if err != nil {
		return err
	}

	err = checkRelationshipLabels(configMap.GetLabels(), "ConfigMap", name)
	if err != nil {
		return err
	}

	return nil
}

func checkRelationshipLabels(labels map[string]string, kind, name string) error {
	expectedValues := map[string]string{
		names.ComponentLabelKey: names.ComponentLabelDefaultValue,
		names.ManagedByLabelKey: names.ManagedByLabelDefaultValue,
	}

	for key, expectedValue := range expectedValues {
		value, found := labels[key]
		if !found || value != expectedValue {
			return fmt.Errorf("%s %s is missing label %s", kind, name, key)
		}
	}

	return nil
}

func checkForClusterRoleRemoval(name string) error {
	err := testenv.Client.Get(context.Background(), types.NamespacedName{Name: name}, &rbacv1.ClusterRole{})
	return isNotFound("ClusterRole", name, err)
}

func checkForClusterRoleBindingRemoval(name string) error {
	err := testenv.Client.Get(context.Background(), types.NamespacedName{Name: name}, &rbacv1.ClusterRoleBinding{})
	return isNotFound("ClusterRoleBinding", name, err)
}

func checkForSecurityContextConstraintsRemoval(name string) error {
	err := testenv.Client.Get(context.Background(), types.NamespacedName{Name: name}, &securityapi.SecurityContextConstraints{})
	if isNotSupportedKind(err) {
		return nil
	}
	return isNotFound("SecurityContextConstraints", name, err)
}

func checkForDaemonSetRemoval(name string) error {
	err := testenv.Client.Get(context.Background(), types.NamespacedName{Name: name, Namespace: components.Namespace}, &appsv1.DaemonSet{})
	return isNotFound("DaemonSets", name, err)
}

func checkForDeploymentRemoval(name string) error {
	err := testenv.Client.Get(context.Background(), types.NamespacedName{Name: name, Namespace: components.Namespace}, &appsv1.Deployment{})
	return isNotFound("Deployments", name, err)
}

func checkForSecretRemoval(name string) error {
	err := testenv.Client.Get(context.Background(), types.NamespacedName{Name: name, Namespace: components.Namespace}, &corev1.Secret{})
	return isNotFound("Secret", name, err)
}

func checkForMutatingWebhookConfigurationRemoval(name string) error {
	err := testenv.Client.Get(context.Background(), types.NamespacedName{Name: name}, &admissionregistrationv1.MutatingWebhookConfiguration{})
	return isNotFound("MutatingWebhookConfiguration", name, err)
}

func checkForServiceRemoval(name string) error {
	err := testenv.Client.Get(context.Background(), types.NamespacedName{Name: name, Namespace: components.Namespace}, &corev1.Service{})
	return isNotFound("Service", name, err)
}

func checkForServiceMonitorRemoval(name string) error {
	err := testenv.Client.Get(context.Background(), types.NamespacedName{Name: name, Namespace: components.Namespace}, &monitoringv1.ServiceMonitor{})
	return isNotFound("ServiceMonitor", name, err)
}

func checkForPrometheusRuleRemoval(name string) error {
	err := testenv.Client.Get(context.Background(), types.NamespacedName{Name: name, Namespace: components.Namespace}, &monitoringv1.PrometheusRule{})
	return isNotFound("PrometheusRule", name, err)
}

func checkForNetworkAttachmentDefinitionRemoval(name string) error {
	err := testenv.Client.Get(context.Background(), types.NamespacedName{Name: name, Namespace: corev1.NamespaceDefault}, &k8snetworkplumbingwgv1.NetworkAttachmentDefinition{})
	if err != nil && isKindNotFound(err) {
		return nil
	}
	return isNotFound("NetworkAttachmentDefinition", name, err)
}

func checkForConfigMapRemoval(name string) error {
	err := testenv.Client.Get(context.Background(), types.NamespacedName{Name: name, Namespace: components.Namespace}, &corev1.ConfigMap{})
	return isNotFound("ConfigMap", name, err)
}

func getMonitoringEndpoint() (*corev1.Endpoints, error) {
	By("Finding CNAO prometheus endpoint")
	endpoint := &corev1.Endpoints{}
	err := testenv.Client.Get(context.Background(), types.NamespacedName{Name: MonitoringComponent.Service, Namespace: components.Namespace}, endpoint)
	if err != nil {
		return nil, err
	}
	return endpoint, nil
}

func ScrapeEndpointAddress(epAddress *corev1.EndpointAddress, epPort int32) (string, error) {
	token, err := getPrometheusToken()
	if err != nil {
		return "", err
	}

	bearer := "Authorization: Bearer " + token
	stdout, _, err := Kubectl("exec", "-n", epAddress.TargetRef.Namespace, epAddress.TargetRef.Name, "--", "curl", "-s", "-k",
		"--header", bearer, fmt.Sprintf("https://%s:%d/metrics", epAddress.IP, epPort))
	if err != nil {
		return "", err
	}
	return stdout, nil
}

func GetMonitoringEndpoint() (*corev1.EndpointAddress, int32, error) {
	endpoint, err := getMonitoringEndpoint()
	if err != nil {
		return nil, 0, err
	}

	epPort := endpoint.Subsets[0].Ports[0].Port
	for _, epAddr := range endpoint.Subsets[0].Addresses {
		if !strings.HasPrefix(epAddr.TargetRef.Name, components.Name) {
			continue
		}
		epAddress := epAddr
		return &epAddress, epPort, nil
	}

	return nil, 0, errors.New("no endpoint target ref name matches CNAO component")
}

func FindMetric(metrics string, expectedMetric string) string {
	for _, line := range strings.Split(metrics, "\n") {
		if strings.HasPrefix(line, expectedMetric+" ") {
			return line
		}
	}
	return ""
}

func isNotFound(componentType string, componentName string, clientGetOutput error) error {
	if clientGetOutput != nil {
		if apierrors.IsNotFound(clientGetOutput) {
			return nil
		}
		return clientGetOutput
	}
	return fmt.Errorf("%s %q has been found", componentType, componentName)
}

func checkConfigCondition(gvk schema.GroupVersionKind, conditionType ConditionType, conditionStatus ConditionStatus) error {
	confStatus := GetConfigStatus(gvk)
	if confStatus == nil {
		return fmt.Errorf("Config Status not found")
	}
	for _, condition := range confStatus.Conditions {
		if condition.Type == conditionsv1.ConditionType(conditionType) {
			if condition.Status == corev1.ConditionStatus(conditionStatus) {
				return nil
			}
			return fmt.Errorf("condition %q is not in expected state %q, obtained state %q", conditionType, conditionStatus, condition.Status)
		}
	}

	// If a condition is missing, it is considered to be False
	if conditionStatus == ConditionFalse {
		return nil
	}

	return fmt.Errorf("condition %q has not been found in the config", conditionType)
}

func isNotSupportedKind(err error) bool {
	return strings.Contains(err.Error(), "no kind is registered for the type")
}

func isKindNotFound(err error) bool {
	return strings.Contains(err.Error(), "no matches for kind")
}

// CheckUnicast return an error in case that the given addresses support multicast or is invalid.
func CheckUnicastAndValidity() (string, string) {
	rangeStart, rangeEnd := retrieveRange()
	parsedRangeStart, err := net.ParseMAC(rangeStart)
	Expect(err).ToNot(HaveOccurred())
	checkUnicast(parsedRangeStart)

	parsedRangeEnd, err := net.ParseMAC(rangeEnd)
	Expect(err).ToNot(HaveOccurred())
	checkUnicast(parsedRangeEnd)
	return rangeStart, rangeEnd
}

func checkUnicast(mac net.HardwareAddr) {
	// A bitwise AND between 00000001 and the mac address first octet.
	// In case where the LSB of the first octet (the multicast bit) is on, it will return 1, and 0 otherwise.
	multicastBit := 1 & mac[0]
	Expect(multicastBit).ToNot(BeNumerically("==", 1), "invalid mac address. Multicast addressing is not supported. Unicast addressing must be used. The first octet is %#0X", mac[0])
}

func retrieveRange() (string, string) {
	configMap := &corev1.ConfigMap{}
	Eventually(func() error {

		return testenv.Client.Get(context.TODO(),
			types.NamespacedName{Namespace: components.Namespace, Name: names.AppliedPrefix + names.OperatorConfig}, configMap)

	}, 50*time.Second, 5*time.Second).ShouldNot(HaveOccurred())

	appliedData, exist := configMap.Data["applied"]
	Expect(exist).To(BeTrue(), "applied data not found in configMap")

	appliedConfig := &cnao.NetworkAddonsConfigSpec{}
	err := json.Unmarshal([]byte(appliedData), appliedConfig)
	Expect(err).ToNot(HaveOccurred())

	Expect(appliedConfig.KubeMacPool).ToNot(BeNil(), "kubemacpool config doesn't exist")

	rangeStart := appliedConfig.KubeMacPool.RangeStart
	rangeEnd := appliedConfig.KubeMacPool.RangeEnd
	return rangeStart, rangeEnd
}

func CheckAvailableEvent(gvk schema.GroupVersionKind) {
	By("Check for Available event")
	config := GetConfig(gvk)
	configV1 := ConvertToConfigV1(config)
	objectEventWatcher := NewObjectEventWatcher(configV1).SinceNow().Timeout(time.Duration(15) * time.Minute)
	stopChan := make(chan struct{})
	defer close(stopChan)
	objectEventWatcher.WaitFor(stopChan, NormalEvent, eventemitter.AvailableReason)
}

func CheckProgressingEvent(gvk schema.GroupVersionKind) {
	By("Check for Progressing event")
	config := GetConfig(gvk)
	configV1 := ConvertToConfigV1(config)
	objectEventWatcher := NewObjectEventWatcher(configV1).SinceNow().Timeout(time.Duration(5) * time.Minute)
	stopChan := make(chan struct{})
	defer close(stopChan)
	objectEventWatcher.WaitFor(stopChan, NormalEvent, eventemitter.ProgressingReason)
}

func CheckModifiedEvent(gvk schema.GroupVersionKind) {
	By("Check for Modified event")
	config := GetConfig(gvk)
	configV1 := ConvertToConfigV1(config)
	objectEventWatcher := NewObjectEventWatcher(configV1).SinceNow().Timeout(time.Duration(5) * time.Minute)
	stopChan := make(chan struct{})
	defer close(stopChan)
	objectEventWatcher.WaitFor(stopChan, NormalEvent, eventemitter.ModifiedReason)
}

func CheckFailedEvent(gvk schema.GroupVersionKind, reason string) {
	By("Check for Failed event")
	config := GetConfig(gvk)
	configV1 := ConvertToConfigV1(config)
	objectEventWatcher := NewObjectEventWatcher(configV1).SinceWatchedObjectResourceVersion().Timeout(time.Duration(5) * time.Minute)
	stopChan := make(chan struct{})
	defer close(stopChan)
	objectEventWatcher.WaitFor(stopChan, WarningEvent, fmt.Sprintf("%s: %s", eventemitter.FailedReason, reason))
}

func CheckNoWarningEvents(gvk schema.GroupVersionKind, rv string) {
	By("Check absence of Warning events")
	config := GetConfig(gvk)
	configV1 := ConvertToConfigV1(config)
	objectEventWatcher := NewObjectEventWatcher(configV1).SinceResourceVersion(rv).Timeout(time.Minute)
	stopChan := make(chan struct{})
	defer close(stopChan)
	objectEventWatcher.WaitNotForType(stopChan, WarningEvent)
}

func getPrometheusToken() (string, error) {
	const (
		monitoringNamespace = "monitoring"
		prometheusPod       = "prometheus-k8s-0"
		container           = "prometheus"
		tokenPath           = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	)

	stdout, stderr, err := Kubectl("exec", "-n", monitoringNamespace, prometheusPod, "-c", container, "--", "cat", tokenPath)
	if err != nil {
		return "", fmt.Errorf("failed getting prometheus token: %w\nstderr: %s", err, stderr)
	}

	return stdout, nil
}
