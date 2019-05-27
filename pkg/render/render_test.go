package render

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RenderTemplate", func() {
	Context("when a valid YAML manifest is given as a parameter", func() {
		It("should successfully render", func() {
			renderData := MakeRenderData()

			renderedFromYAML, err := RenderTemplate("testdata/simple.yaml", &renderData)
			Expect(err).NotTo(HaveOccurred())

			Expect(renderedFromYAML).To(HaveLen(1))
			expected := `
{
	"apiVersion": "v1",
	"kind": "Pod",
	"metadata": {
		"name": "busybox1",
		"namespace": "ns"
	},
	"spec": {
		"containers": [
			{
  				"image": "busybox"
			}
		]
	}
}
`
			Expect(renderedFromYAML[0].MarshalJSON()).To(MatchJSON(expected))
		})
	})

	Context("when a identical YAML and JSONs manifest are given as a parameters", func() {
		It("should return identical objects", func() {
			renderData := MakeRenderData()

			renderedFromYAML, err := RenderTemplate("testdata/simple.yaml", &renderData)
			Expect(err).NotTo(HaveOccurred())

			renderedFromJSON, err := RenderTemplate("testdata/simple.json", &renderData)
			Expect(err).NotTo(HaveOccurred())

			Expect(renderedFromJSON).To(Equal(renderedFromYAML))
		})
	})

	Context("when YAML manifest with multiple objects given as a parameter", func() {
		It("should successfully render list of objects", func() {
			renderData := MakeRenderData()

			rendered, err := RenderTemplate("testdata/multiple.yaml", &renderData)
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
				renderData := MakeRenderData()
				_, err := RenderTemplate("testdata/template.yaml", &renderData)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(HaveSuffix(`function "fname" not defined`))
			})
		})

		Context("when requested variables are missing", func() {
			It("should fail with expected error", func() {
				renderData := MakeRenderData()
				renderData.Funcs["fname"] = func(s string) string { return "test-" + s }
				_, err := RenderTemplate("testdata/template.yaml", &renderData)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(HaveSuffix(`has no entry for key "Namespace"`))
			})
		})

		Context("when all functions variables are provided", func() {
			It("should successfully render", func() {
				renderData := MakeRenderData()
				renderData.Data["Namespace"] = "myns"
				renderData.Funcs["fname"] = func(s string) string { return "test-" + s }
				rendered, err := RenderTemplate("testdata/template.yaml", &renderData)
				Expect(err).NotTo(HaveOccurred())

				Expect(rendered[0].GetName()).To(Equal("test-podname"))
				Expect(rendered[0].GetNamespace()).To(Equal("myns"))
			})
		})
	})
})

var _ = Describe("RenderDir", func() {
	Context("rendering all files in a directory", func() {
		It("should successfully render all JSON/YAML manifest, including templates", func() {
			renderData := MakeRenderData()
			renderData.Funcs["fname"] = func(s string) string { return s }
			renderData.Data["Namespace"] = "myns"

			rendered, err := RenderDir("testdata", &renderData)
			Expect(err).NotTo(HaveOccurred())
			Expect(rendered).To(HaveLen(6))
		})
	})
})
