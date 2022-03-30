package operations

import (
	"context"
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"gopkg.in/yaml.v2"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
	cnaov1 "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/v1"
	cnaov1alpha1 "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/v1alpha1"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/names"
	testenv "github.com/kubevirt/cluster-network-addons-operator/test/env"
)

func GetConfig(gvk schema.GroupVersionKind) *unstructured.Unstructured {
	config := &unstructured.Unstructured{}
	config.SetGroupVersionKind(gvk)
	err := testenv.Client.Get(context.TODO(), types.NamespacedName{Name: names.OPERATOR_CONFIG}, config)
	if apierrors.IsNotFound(err) {
		return nil
	}
	Expect(err).NotTo(HaveOccurred(), "Failed to fetch Config")
	return config
}

func CreateConfig(gvk schema.GroupVersionKind, configSpec cnao.NetworkAddonsConfigSpec) {
	By(fmt.Sprintf("Applying NetworkAddonsConfig spec:\n%s", configSpecToYaml(configSpec)))

	config := &unstructured.Unstructured{}
	config.SetGroupVersionKind(gvk)
	config.SetName(names.OPERATOR_CONFIG)

	unstructuredConfigSpec, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&configSpec)
	Expect(err).NotTo(HaveOccurred(), "Failed to convert config spec to unstructured")
	config.Object["spec"] = unstructuredConfigSpec

	err = testenv.Client.Create(context.TODO(), config)
	Expect(err).NotTo(HaveOccurred(), "Failed to create the Config")
}

func UpdateConfig(gvk schema.GroupVersionKind, configSpec cnao.NetworkAddonsConfigSpec) {
	By(fmt.Sprintf("Updating NetworkAddonsConfig:\n%s", configSpecToYaml(configSpec)))

	// Get current Config
	config := GetConfig(gvk)

	// Update the Config with the desired Spec
	config.Object["spec"] = configSpec
	err := testenv.Client.Update(context.TODO(), config)
	Expect(err).NotTo(HaveOccurred(), "Failed to update the Config")
}

func DeleteConfig(gvk schema.GroupVersionKind) {
	By("Removing NetworkAddonsConfig")

	config := &unstructured.Unstructured{}
	config.SetGroupVersionKind(gvk)
	config.SetName(names.OPERATOR_CONFIG)

	err := testenv.Client.Delete(context.TODO(), config)
	ExpectWithOffset(1, err).NotTo(HaveOccurred(), "Failed to remove the Config")

	// Wait until the config is deleted
	EventuallyWithOffset(1, func() error {
		return testenv.Client.Get(context.TODO(), types.NamespacedName{Name: names.OPERATOR_CONFIG}, config)
	}, 60*time.Second, 1*time.Second).Should(SatisfyAll(HaveOccurred(), WithTransform(apierrors.IsNotFound, BeTrue())), fmt.Sprintf("should successfuly delete config '%s'", config.GetName()))

}

func GetConfigStatus(gvk schema.GroupVersionKind) *cnao.NetworkAddonsConfigStatus {
	config := GetConfig(gvk)
	if config != nil {
		switch gvk {
		case GetCnaoV1GroupVersionKind():
			return &ConvertToConfigV1(config).Status
		case GetCnaoV1alpha1GroupVersionKind():
			return &ConvertToConfigV1alpha1(config).Status
		}

		Fail(fmt.Sprintf("gvk %v not supported", gvk))
	}
	return nil
}

func ConvertToConfigV1(unstructuredConfig *unstructured.Unstructured) *cnaov1.NetworkAddonsConfig {
	configV1 := &cnaov1.NetworkAddonsConfig{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredConfig.Object, configV1)
	Expect(err).NotTo(HaveOccurred(), "Failed to convert unstructured config to cnaov1 Config")

	return configV1
}

func ConvertToConfigV1alpha1(unstructuredConfig *unstructured.Unstructured) *cnaov1alpha1.NetworkAddonsConfig {
	configV1alpha1 := &cnaov1alpha1.NetworkAddonsConfig{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredConfig.Object, configV1alpha1)
	Expect(err).NotTo(HaveOccurred(), "Failed to convert unstructured config to cnaov1alpha1 Config")

	return configV1alpha1
}

func GetCnaoV1GroupVersionKind() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   "networkaddonsoperator.network.kubevirt.io",
		Version: "v1",
		Kind:    "NetworkAddonsConfig",
	}
}

func GetCnaoV1alpha1GroupVersionKind() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   "networkaddonsoperator.network.kubevirt.io",
		Version: "v1alpha1",
		Kind:    "NetworkAddonsConfig",
	}
}

// Convert NetworkAddonsConfig specification to a yaml format we would expect in a manifest
func configSpecToYaml(configSpec cnao.NetworkAddonsConfigSpec) string {
	manifest, err := yaml.Marshal(configSpec)
	if err != nil {
		panic(err)
	}

	manifestLines := strings.Split(string(manifest), "\n")

	// We don't want to show non-set (default) values, usually null. Try our best to filter those out.
	manifestLinesWithoutEmptyValues := []string{}
	for _, line := range manifestLines {
		// If root attribute (e.g. ImagePullPolicy) is set to default, drop it. If it
		// is a nested attribute (e.g. KubeMacPool's RangeEnd), keep it.
		rootAttributeSetToDefault := !strings.Contains(line, "  ") && (strings.Contains(line, ": \"\"") || strings.Contains(line, ": null"))
		if line != "" && !rootAttributeSetToDefault {
			manifestLinesWithoutEmptyValues = append(manifestLinesWithoutEmptyValues, line)
		}
	}

	// If any values has been set, return Spec in a nice YAML format
	if len(manifestLinesWithoutEmptyValues) > 0 {
		indentedManifest := strings.TrimSpace(strings.Join(manifestLinesWithoutEmptyValues, "\n"))
		return indentedManifest
	}

	// Note that it is empty otherwise
	return "Empty Spec"
}
