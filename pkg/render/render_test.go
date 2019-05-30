package render_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"encoding/json"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/kubevirt/cluster-network-addons-operator/pkg/render"
)

var _ = Describe("RenderTemplate", func() {
	Context("when a valid manifest is given as a parameter", func() {
		DescribeTable("should be able to render templates from different formats", func(template string) {
			renderData := render.MakeRenderData()

			renderedFromYAML, err := render.RenderTemplate(template, &renderData)
			Expect(err).NotTo(HaveOccurred())
			Expect(renderedFromYAML).To(HaveLen(1))

			renderedPod, err := unstructuredToPod(renderedFromYAML[0])
			Expect(err).NotTo(HaveOccurred())

			expectedPod := &apiv1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "busybox1",
					Namespace: "ns",
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Image: "busybox",
						},
					},
				},
			}

			Expect(renderedPod).To(Equal(expectedPod))
		},
			Entry("YAML", "testdata/simple.yaml"),
			Entry("JSON", "testdata/simple.json"),
		)
	})

	Context("when YAML manifest with multiple objects given as a parameter", func() {
		It("should successfully render list of objects", func() {
			renderData := render.MakeRenderData()

			rendered, err := render.RenderTemplate("testdata/multiple.yaml", &renderData)
			Expect(err).NotTo(HaveOccurred())

			Expect(rendered).To(HaveLen(3))

			Expect(rendered[0].GetObjectKind().GroupVersionKind().String()).To(Equal("/v1, Kind=Pod"))
			Expect(rendered[1].GetObjectKind().GroupVersionKind().String()).To(Equal("rbac.authorization.k8s.io/v1, Kind=ClusterRoleBinding"))
			Expect(rendered[2].GetObjectKind().GroupVersionKind().String()).To(Equal("/v1, Kind=ConfigMap"))
		})
	})

	Context("when YAML manifest including templating given as a parameter", func() {
		Context("when requested functions are missing", func() {
			It("should fail with expected error", func() {
				renderData := render.MakeRenderData()
				renderData.Data["Namespace"] = "myns"
				_, err := render.RenderTemplate("testdata/template.yaml", &renderData)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(HaveSuffix(`function "fname" not defined`))
			})
		})

		Context("when requested variables are missing", func() {
			It("should fail with expected error", func() {
				renderData := render.MakeRenderData()
				renderData.Funcs["fname"] = func(s string) string { return "test-" + s }
				_, err := render.RenderTemplate("testdata/template.yaml", &renderData)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(HaveSuffix(`has no entry for key "Namespace"`))
			})
		})

		Context("when all functions variables are provided", func() {
			It("should successfully render", func() {
				renderData := render.MakeRenderData()
				renderData.Data["Namespace"] = "myns"
				renderData.Funcs["fname"] = func(s string) string { return "test-" + s }
				rendered, err := render.RenderTemplate("testdata/template.yaml", &renderData)
				Expect(err).NotTo(HaveOccurred())

				renderedPod, err := unstructuredToPod(rendered[0])
				Expect(err).NotTo(HaveOccurred())

				expectedPod := &apiv1.Pod{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Pod",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-podname",
						Namespace: "myns",
					},
					Spec: apiv1.PodSpec{
						Containers: []apiv1.Container{
							{
								Image: "busybox",
							},
						},
					},
				}

				Expect(renderedPod).To(Equal(expectedPod))
			})
		})
	})
})

var _ = Describe("RenderDir", func() {
	Context("rendering all files in a directory", func() {
		It("should successfully render all JSON/YAML manifest, including templates", func() {
			renderData := render.MakeRenderData()
			renderData.Funcs["fname"] = func(s string) string { return s }
			renderData.Data["Namespace"] = "myns"

			rendered, err := render.RenderDir("testdata", &renderData)
			Expect(err).NotTo(HaveOccurred())
			Expect(rendered).To(HaveLen(6))
		})
	})
})

func unstructuredToPod(unstructuredPod *unstructured.Unstructured) (*apiv1.Pod, error) {
	unstructuredPodJSON, err := json.Marshal(unstructuredPod)
	if err != nil {
		return nil, err
	}

	structuredPod := &apiv1.Pod{}
	err = json.Unmarshal(unstructuredPodJSON, structuredPod)
	if err != nil {
		return nil, err
	}

	return structuredPod, nil
}
