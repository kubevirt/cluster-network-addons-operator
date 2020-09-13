package operations

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	framework "github.com/operator-framework/operator-sdk/pkg/test"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
	cnaov1alpha1 "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/v1alpha1"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/names"
)

type ConfigV1alpha1 struct {}

func (c *ConfigV1alpha1) GetConfig() map[string]interface{} {
	By("Getting the current config")
	configV1alpha1 := &cnaov1alpha1.NetworkAddonsConfig{}
	err := framework.Global.Client.Get(context.TODO(), types.NamespacedName{Name: names.OPERATOR_CONFIG}, configV1alpha1)
	if apierrors.IsNotFound(err) {
		return nil
	}
	Expect(err).NotTo(HaveOccurred(), "Failed to fetch Config")

	unstructuredConfig, err := runtime.DefaultUnstructuredConverter.ToUnstructured(configV1alpha1)
	Expect(err).NotTo(HaveOccurred(), "Failed to convert config to unstructured")

	return unstructuredConfig
}

func (c *ConfigV1alpha1) CreateConfig(configSpec cnao.NetworkAddonsConfigSpec) {
	By(fmt.Sprintf("Applying NetworkAddonsConfig:\n%s", configSpecToYaml(configSpec)))

	config := &cnaov1alpha1.NetworkAddonsConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: names.OPERATOR_CONFIG,
		},
		Spec: configSpec,
	}

	err := framework.Global.Client.Create(context.TODO(), config, &framework.CleanupOptions{})
	Expect(err).NotTo(HaveOccurred(), "Failed to create the Config")
}

func (c *ConfigV1alpha1) UpdateConfig(configSpec cnao.NetworkAddonsConfigSpec) {
	By(fmt.Sprintf("Updating NetworkAddonsConfig:\n%s", configSpecToYaml(configSpec)))

	// Get current Config
	unstructuredConfig := c.GetConfig()
	configV1alpha1 := &cnaov1alpha1.NetworkAddonsConfig{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredConfig, configV1alpha1)
	Expect(err).NotTo(HaveOccurred(), "Failed to convert unstructured config to cnaov1alpha1 Config")

	// Update the Config with the desired Spec
	configV1alpha1.Spec = configSpec
	err = framework.Global.Client.Update(context.TODO(), configV1alpha1)
	Expect(err).NotTo(HaveOccurred(), "Failed to update the Config")
}

func (c *ConfigV1alpha1) DeleteConfig() {
	By("Removing NetworkAddonsConfig")

	config := &cnaov1alpha1.NetworkAddonsConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: names.OPERATOR_CONFIG,
		},
	}

	err := framework.Global.Client.Delete(context.TODO(), config)
	ExpectWithOffset(1, err).NotTo(HaveOccurred(), "Failed to remove the Config")

	// Wait until the config is deleted
	EventuallyWithOffset(1, func() error {
		return framework.Global.Client.Get(context.TODO(), types.NamespacedName{Name: names.OPERATOR_CONFIG}, &cnaov1alpha1.NetworkAddonsConfig{})
	}, 60*time.Second, 1*time.Second).Should(SatisfyAll(HaveOccurred(), WithTransform(apierrors.IsNotFound, BeTrue())), fmt.Sprintf("should successfuly delete config '%s'", config.Name))

}

func (c *ConfigV1alpha1) GetStatus() cnao.NetworkAddonsConfigStatus {
	// Get current Config
	unstructuredConfig := c.GetConfig()
	configV1alpha1 := &cnaov1alpha1.NetworkAddonsConfig{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredConfig, configV1alpha1)
	Expect(err).NotTo(HaveOccurred(), "Failed to convert unstructured config to cnaov1alpha1 Config")

	return configV1alpha1.Status
}
