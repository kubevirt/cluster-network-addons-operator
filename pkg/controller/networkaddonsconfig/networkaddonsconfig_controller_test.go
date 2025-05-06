package networkaddonsconfig

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kubevirt/cluster-network-addons-operator/pkg/names"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/render"
)

type checkUnit struct {
	key           string
	shouldExist   bool
	expectedValue string
}

var _ = Describe("Networkaddonsconfig", func() {
	const testDataFile = "testdata/data.yaml"

	Context("When CR is not labeled", func() {
		var objs []*unstructured.Unstructured

		BeforeEach(func() {
			var err error
			renderData := render.MakeRenderData()
			objs, err = render.RenderTemplate(testDataFile, &renderData)
			Expect(err).NotTo(HaveOccurred())

			crLabels := map[string]string{}
			err = updateObjectsLabels(crLabels, objs)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Should find only default relationship labels", func() {
			appLabelKeys := []checkUnit{
				{
					key:           names.ComponentLabelKey,
					shouldExist:   true,
					expectedValue: names.ComponentLabelDefaultValue,
				},
				{
					key:           names.ManagedByLabelKey,
					shouldExist:   true,
					expectedValue: names.ManagedByLabelDefaultValue,
				},
				{
					key:           names.PartOfLabelKey,
					shouldExist:   false,
					expectedValue: "Invalid",
				},
				{
					key:           names.VersionLabelKey,
					shouldExist:   false,
					expectedValue: "Invalid",
				},
			}

			checkObjectsRelationshipLabels(objs, appLabelKeys)
		})
	})
	Context("When CR is labeled", func() {
		var objs []*unstructured.Unstructured
		var crLabels map[string]string
		const componentCrLabelValue = "component_unit_tests"
		const managedByCrLabelValue = "managed_by_unit_tests"
		const partOfCrLabelValue = "part_of_unit_tests"
		const versionCrLabelValue = "version_of_unit_tests"

		BeforeEach(func() {
			var err error
			renderData := render.MakeRenderData()
			objs, err = render.RenderTemplate(testDataFile, &renderData)
			Expect(err).NotTo(HaveOccurred())

			crLabels = map[string]string{}
			crLabels[names.ComponentLabelKey] = componentCrLabelValue
			crLabels[names.ManagedByLabelKey] = managedByCrLabelValue
			crLabels[names.PartOfLabelKey] = partOfCrLabelValue
			crLabels[names.VersionLabelKey] = versionCrLabelValue
			err = updateObjectsLabels(crLabels, objs)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Should find all labels overridden by CR labels except managed-by that should keep original", func() {
			expectedAppLabelKeys := []checkUnit{
				{
					key:           names.ComponentLabelKey,
					shouldExist:   true,
					expectedValue: componentCrLabelValue,
				},
				{
					key:           names.ManagedByLabelKey,
					shouldExist:   true,
					expectedValue: names.ManagedByLabelDefaultValue,
				},
				{
					key:           names.PartOfLabelKey,
					shouldExist:   true,
					expectedValue: partOfCrLabelValue,
				},
				{
					key:           names.VersionLabelKey,
					shouldExist:   true,
					expectedValue: versionCrLabelValue,
				},
			}

			checkObjectsRelationshipLabels(objs, expectedAppLabelKeys)
		})
	})
})

func checkObjectsRelationshipLabels(objs []*unstructured.Unstructured, appLabelKeys []checkUnit) {
	appLabelsNotExists := []checkUnit{
		{
			key:           names.ComponentLabelKey,
			shouldExist:   false,
			expectedValue: "Invalid",
		},
		{
			key:           names.ManagedByLabelKey,
			shouldExist:   false,
			expectedValue: "Invalid",
		},
		{
			key:           names.PartOfLabelKey,
			shouldExist:   false,
			expectedValue: "Invalid",
		},
		{
			key:           names.VersionLabelKey,
			shouldExist:   false,
			expectedValue: "Invalid",
		},
	}

	for _, obj := range objs {
		labels := obj.GetLabels()
		checkLabels(labels, appLabelKeys)

		kind := obj.GetKind()
		templateLabels, _, _ := unstructured.NestedStringMap(obj.Object, "spec", "template", "metadata", "labels")
		if kind == "DaemonSet" || kind == "ReplicaSet" || kind == "Deployment" || kind == "StatefulSet" {
			checkLabels(templateLabels, appLabelKeys)
		} else {
			checkLabels(templateLabels, appLabelsNotExists)
		}
	}
}

func checkLabels(labels map[string]string, units []checkUnit) {
	for _, unit := range units {
		value, found := labels[unit.key]
		Expect(found).To(Equal(unit.shouldExist), fmt.Sprintf("%s exist test", unit.key))
		if unit.shouldExist {
			Expect(value).To(Equal(unit.expectedValue), fmt.Sprintf("%s value test", unit.key))
		}
	}
}
