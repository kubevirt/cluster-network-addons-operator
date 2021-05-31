package check

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	conditionsv1 "github.com/openshift/custom-resource-status/conditions/v1"
	securityapi "github.com/openshift/origin/pkg/security/apis/security"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
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

	"github.com/kubevirt/cluster-network-addons-operator/pkg/eventemitter"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/names"

	. "github.com/kubevirt/cluster-network-addons-operator/test/kubectl"
	. "github.com/kubevirt/cluster-network-addons-operator/test/okd"
	. "github.com/kubevirt/cluster-network-addons-operator/test/operations"
)

const (
	CheckImmediately   = time.Microsecond
	CheckDoNotRepeat   = time.Microsecond
	CheckIgnoreVersion = "IGNORE"
)

func CheckComponentsDeployment(components []Component) {
	for _, component := range components {
		if component.ComponentName == MultusComponent.ComponentName && IsOnOKDCluster() {
			// On OpenShift 4, Multus is not owned by us
			continue
		}

		By(fmt.Sprintf("Checking that component %s is deployed", component.ComponentName))
		err := checkForComponent(&component)
		Expect(err).NotTo(HaveOccurred(), "Component has not been fully deployed")
	}
}

func CheckCrdExplainable() {
	By("Checking crd is explainable")
	explain, err := Kubectl("explain", "networkaddonsconfigs")
	Expect(err).NotTo(HaveOccurred(), "explain should not return error")

	Expect(explain).ToNot(BeEmpty(), "explain output should not be empty")
	Expect(explain).ToNot(ContainSubstring("<empty>"), "explain output should not contain <empty> fields")
}

func CheckComponentsRemoval(components []Component) {
	for _, component := range components {
		if component.ComponentName == MultusComponent.ComponentName && IsOnOKDCluster() {
			// On OpenShift 4, Multus is not owned by us
			continue
		}

		// TODO: Once finalizers are implemented, we should switch to using
		// once-time checks, since after NodeNetworkState removal, no components
		// should be left over.
		By(fmt.Sprintf("Checking that component %s has been removed", component.ComponentName))
		Eventually(func() error {
			return checkForComponentRemoval(&component)
		}, 5*time.Minute, time.Second).ShouldNot(HaveOccurred(), fmt.Sprintf("%s component has not been fully removed within the given timeout\ncluster Info:\n%v", component.ComponentName, gatherClusterInfo()))
	}
}

func CheckConfigCondition(gvk schema.GroupVersionKind, conditionType ConditionType, conditionStatus ConditionStatus, timeout time.Duration, duration time.Duration) {
	By(fmt.Sprintf("Checking that condition %q status is set to %s", conditionType, conditionStatus))
	getAndCheckCondition := func() error {

		return checkConfigCondition(gvk, conditionType, conditionStatus)
	}

	if timeout != CheckImmediately {
		Eventually(getAndCheckCondition, timeout, 10*time.Millisecond).ShouldNot(HaveOccurred(), fmt.Sprintf("Timed out waiting for the condition, current config:\n%v\ncluster Info:\n%v", configToYaml(gvk), gatherClusterInfo()))
	} else {
		Expect(getAndCheckCondition()).NotTo(HaveOccurred(), fmt.Sprintf("Condition is not in the expected state, current config:\n%v\ncluster Info:\n%v", configToYaml(gvk), gatherClusterInfo()))
	}

	if duration != CheckDoNotRepeat {
		Consistently(getAndCheckCondition, duration, 10*time.Millisecond).ShouldNot(HaveOccurred(), fmt.Sprintf("Condition prematurely changed its value, current config:\n%v\ncluster Info:\n%v", configToYaml(gvk), gatherClusterInfo()))
	}
}

func PlacementListFromComponentDaemonSets(component Component) ([]cnao.Placement, error) {
	placementList := []cnao.Placement{}
	for _, daemonSetName := range component.DaemonSets {
		daemonSet := appsv1.DaemonSet{}
		err := framework.Global.Client.Get(context.Background(), types.NamespacedName{Name: daemonSetName, Namespace: components.Namespace}, &daemonSet)
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

func GetEnvVarsFromDeployment(deploymentName string) ([]corev1.EnvVar, error) {
	deployment := appsv1.Deployment{}
	err := framework.Global.Client.Get(context.Background(), types.NamespacedName{Name: deploymentName, Namespace: components.Namespace}, &deployment)
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
		Eventually(getAndCheckVersions, timeout, time.Second).ShouldNot(HaveOccurred(), fmt.Sprintf("Timed out waiting for the expected versions, current config:\n%v\ncluster Info:\n%v", configToYaml(gvk), gatherClusterInfo()))
	} else {
		Expect(getAndCheckVersions()).NotTo(HaveOccurred(), fmt.Sprintf("Versions are not in the expected state, current config:\n%v\ncluster Info:\n%v", configToYaml(gvk), gatherClusterInfo()))
	}

	if duration != CheckDoNotRepeat {
		Consistently(getAndCheckVersions, duration, time.Second).ShouldNot(HaveOccurred(), fmt.Sprintf("Versions prematurely changed their values, current config:\n%v\ncluster Info:\n%v", configToYaml(gvk), gatherClusterInfo()))
	}
}

func CheckOperatorIsReady(timeout time.Duration) {
	By("Checking that the operator is up and running")
	if timeout != CheckImmediately {
		Eventually(func() error {
			return checkForDeployment(components.Name)
		}, timeout, time.Second).ShouldNot(HaveOccurred(), fmt.Sprintf("Timed out waiting for the operator to become ready"))
	} else {
		Expect(checkForDeployment(components.Name)).ShouldNot(HaveOccurred(), "Operator is not ready")
	}
}

func CheckForLeftoverObjects(currentVersion string) {
	listOptions := client.ListOptions{}
	key := cnaov1.SchemeGroupVersion.Group + "/version"
	labelSelector, err := k8slabels.Parse(fmt.Sprintf("%s,%s != %s", key, key, currentVersion))
	Expect(err).NotTo(HaveOccurred())
	listOptions.LabelSelector = labelSelector

	deployments := appsv1.DeploymentList{}
	err = framework.Global.Client.List(context.Background(), &deployments, &listOptions)
	Expect(err).NotTo(HaveOccurred())
	Expect(deployments.Items).To(BeEmpty(), "Found leftover objects from the previous operator version")

	daemonSets := appsv1.DaemonSetList{}
	err = framework.Global.Client.List(context.Background(), &daemonSets, &listOptions)
	Expect(err).NotTo(HaveOccurred())
	Expect(daemonSets.Items).To(BeEmpty(), "Found leftover objects from the previous operator version")

	configMaps := corev1.ConfigMapList{}
	err = framework.Global.Client.List(context.Background(), &configMaps, &listOptions)
	Expect(err).NotTo(HaveOccurred())
	Expect(configMaps.Items).To(BeEmpty(), "Found leftover objects from the previous operator version")

	namespaces := corev1.NamespaceList{}
	err = framework.Global.Client.List(context.Background(), &namespaces, &listOptions)
	Expect(err).NotTo(HaveOccurred())
	Expect(namespaces.Items).To(BeEmpty(), "Found leftover objects from the previous operator version")

	clusterRoles := rbacv1.ClusterRoleList{}
	err = framework.Global.Client.List(context.Background(), &clusterRoles, &listOptions)
	Expect(err).NotTo(HaveOccurred())
	Expect(clusterRoles.Items).To(BeEmpty(), "Found leftover objects from the previous operator version")

	clusterRoleBindings := rbacv1.ClusterRoleList{}
	err = framework.Global.Client.List(context.Background(), &clusterRoleBindings, &listOptions)
	Expect(err).NotTo(HaveOccurred())
	Expect(clusterRoleBindings.Items).To(BeEmpty(), "Found leftover objects from the previous operator version")

	roles := rbacv1.RoleList{}
	err = framework.Global.Client.List(context.Background(), &roles, &listOptions)
	Expect(err).NotTo(HaveOccurred())
	Expect(roles.Items).To(BeEmpty(), "Found leftover objects from the previous operator version")

	roleBindings := rbacv1.RoleBindingList{}
	err = framework.Global.Client.List(context.Background(), &roleBindings, &listOptions)
	Expect(err).NotTo(HaveOccurred())
	Expect(roleBindings.Items).To(BeEmpty(), "Found leftover objects from the previous operator version")

	serviceAccounts := corev1.ServiceAccountList{}
	err = framework.Global.Client.List(context.Background(), &serviceAccounts, &listOptions)
	Expect(err).NotTo(HaveOccurred())
	Expect(serviceAccounts.Items).To(BeEmpty(), "Found leftover objects from the previous operator version")
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
		errsAppend(checkForDeployment(deployment))
	}

	if component.Secret != "" {
		errsAppend(checkForSecret(component.Secret))
	}

	if component.MutatingWebhookConfiguration != "" {
		errsAppend(checkForMutatingWebhookConfiguration(component.MutatingWebhookConfiguration))
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

	if component.Secret != "" {
		errsAppend(checkForSecretRemoval(component.Secret))
	}

	if component.MutatingWebhookConfiguration != "" {
		errsAppend(checkForMutatingWebhookConfigurationRemoval(component.MutatingWebhookConfiguration))
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
	return framework.Global.Client.Get(context.Background(), types.NamespacedName{Name: name}, &rbacv1.ClusterRole{})
}

func checkForClusterRoleBinding(name string) error {
	return framework.Global.Client.Get(context.Background(), types.NamespacedName{Name: name}, &rbacv1.ClusterRoleBinding{})
}

func checkForSecurityContextConstraints(name string) error {
	err := framework.Global.Client.Get(context.Background(), types.NamespacedName{Name: name}, &securityapi.SecurityContextConstraints{})
	if isNotSupportedKind(err) {
		return nil
	}
	return err
}

func checkForDeployment(name string) error {
	deployment := appsv1.Deployment{}

	err := framework.Global.Client.Get(context.Background(), types.NamespacedName{Name: name, Namespace: components.Namespace}, &deployment)
	if err != nil {
		return err
	}

	labels := deployment.GetLabels()
	if labels != nil {
		if _, operatorLabelSet := labels[cnaov1.SchemeGroupVersion.Group+"/version"]; !operatorLabelSet {
			return fmt.Errorf("Deployment %s/%s is missing operator label", components.Namespace, name)
		}
	}

	if deployment.Status.UnavailableReplicas > 0 || deployment.Status.AvailableReplicas == 0 {
		manifest, err := yaml.Marshal(deployment)
		if err != nil {
			panic(err)
		}
		return fmt.Errorf("Deployment %s/%s is not ready, current state:\n%v\ncluster Info:\n%v", components.Namespace, name, string(manifest), gatherClusterInfo())
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

	err := framework.Global.Client.List(context.Background(), &pods, listOptions...)
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

	err := framework.Global.Client.Get(context.Background(), types.NamespacedName{Name: name, Namespace: components.Namespace}, &daemonSet)
	if err != nil {
		return err
	}

	labels := daemonSet.GetLabels()
	if labels != nil {
		if _, operatorLabelSet := labels[cnaov1.SchemeGroupVersion.Group+"/version"]; !operatorLabelSet {
			return fmt.Errorf("DaemonSet %s/%s is missing operator label", components.Namespace, name)
		}
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
	return framework.Global.Client.Get(context.Background(), types.NamespacedName{Name: name, Namespace: components.Namespace}, &corev1.Secret{})
}

func checkForMutatingWebhookConfiguration(name string) error {
	return framework.Global.Client.Get(context.Background(), types.NamespacedName{Name: name}, &admissionregistrationv1.MutatingWebhookConfiguration{})
}

func checkForClusterRoleRemoval(name string) error {
	err := framework.Global.Client.Get(context.Background(), types.NamespacedName{Name: name}, &rbacv1.ClusterRole{})
	return isNotFound("ClusterRole", name, err)
}

func checkForClusterRoleBindingRemoval(name string) error {
	err := framework.Global.Client.Get(context.Background(), types.NamespacedName{Name: name}, &rbacv1.ClusterRoleBinding{})
	return isNotFound("ClusterRoleBinding", name, err)
}

func checkForSecurityContextConstraintsRemoval(name string) error {
	err := framework.Global.Client.Get(context.Background(), types.NamespacedName{Name: name}, &securityapi.SecurityContextConstraints{})
	if isNotSupportedKind(err) {
		return nil
	}
	return isNotFound("SecurityContextConstraints", name, err)
}

func checkForSecretRemoval(name string) error {
	err := framework.Global.Client.Get(context.Background(), types.NamespacedName{Name: name, Namespace: components.Namespace}, &corev1.Secret{})
	return isNotFound("Secret", name, err)
}

func checkForMutatingWebhookConfigurationRemoval(name string) error {
	err := framework.Global.Client.Get(context.Background(), types.NamespacedName{Name: name}, &admissionregistrationv1.MutatingWebhookConfiguration{})
	return isNotFound("MutatingWebhookConfiguration", name, err)
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
			return fmt.Errorf("condition %q is not in expected state %q, obtained state %q, obtained config:\n%vcluster Info:\n%v", conditionType, conditionStatus, condition.Status, configToYaml(gvk), gatherClusterInfo())
		}
	}

	// If a condition is missing, it is considered to be False
	if conditionStatus == ConditionFalse {
		return nil
	}

	return fmt.Errorf("condition %q has not been found in the config", conditionType)
}

func gatherClusterInfo() string {
	podsStatus := cnaoPodsStatus()
	describeAll := describeAll()
	return strings.Join([]string{podsStatus, describeAll}, "\n")
}

func cnaoPodsStatus() string {
	podsStatus, err := Kubectl("-n", components.Namespace, "get", "pods")
	return fmt.Sprintf("CNAO pods Status:\n%v\nerror:\n%v", podsStatus, err)
}

func describeAll() string {
	description, err := Kubectl("-n", components.Namespace, "describe", "all")
	return fmt.Sprintf("describe all CNAO components:\n%v\nerror:\n%v", description, err)
}

func isNotSupportedKind(err error) bool {
	return strings.Contains(err.Error(), "no kind is registered for the type")
}

func configToYaml(gvk schema.GroupVersionKind) string {
	config := GetConfig(gvk)
	manifest, err := yaml.Marshal(config)
	if err != nil {
		panic(err)
	}
	return string(manifest)
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

		return framework.Global.Client.Get(context.TODO(),
			types.NamespacedName{Namespace: components.Namespace, Name: names.APPLIED_PREFIX + names.OPERATOR_CONFIG}, configMap)

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
