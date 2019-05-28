package apply_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"bytes"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/kubevirt/cluster-network-addons-operator/pkg/apply"
)

var _ = Describe("MergeObjectForUpdate", func() {
	// Namespaces use the "generic" logic; deployments and services
	// have custom logic
	Context("when given a generic object (Namespace)", func() {
		cur := unstructuredFromYaml(`
apiVersion: v1
kind: Namespace
metadata:
  name: ns1
  labels:
    a: cur
    b: cur
  annotations:
    a: cur
    b: cur`)

		upd := unstructuredFromYaml(`
apiVersion: v1
kind: Namespace
metadata:
  name: ns1
  labels:
    a: upd
    c: upd
  annotations:
    a: upd
    c: upd`)

		It("should successfully merge", func() {
			// this mutates updated
			err := apply.MergeObjectForUpdate(cur, upd)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should merge annotations", func() {
			Expect(upd.GetLabels()).To(Equal(map[string]string{
				"a": "upd",
				"b": "cur",
				"c": "upd",
			}))
		})

		It("should overwrite everything else", func() {
			Expect(upd.GetAnnotations()).To(Equal(map[string]string{
				"a": "upd",
				"b": "cur",
				"c": "upd",
			}))
		})
	})

	Context("when given a Deployment", func() {
		cur := unstructuredFromYaml(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: d1
  labels:
    a: cur
    b: cur
  annotations:
    deployment.kubernetes.io/revision: cur
    a: cur
    b: cur`)

		upd := unstructuredFromYaml(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: d1
  labels:
    a: upd
    c: upd
  annotations:
    deployment.kubernetes.io/revision: upd
    a: upd
    c: upd`)

		It("should successfully merge", func() {
			// this mutates updated
			err := apply.MergeObjectForUpdate(cur, upd)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should merge annotations", func() {
			Expect(upd.GetAnnotations()).To(Equal(map[string]string{
				"a": "upd",
				"b": "cur",
				"c": "upd",

				"deployment.kubernetes.io/revision": "cur",
			}))
		})

		It("should not merge labels", func() {
			Expect(upd.GetLabels()).To(Equal(map[string]string{
				"a": "upd",
				"b": "cur",
				"c": "upd",
			}))
		})
	})

	Context("when given a Service", func() {
		cur := unstructuredFromYaml(`
apiVersion: v1
kind: Service
metadata:
  name: d1
spec:
  clusterIP: cur`)

		upd := unstructuredFromYaml(`
apiVersion: v1
kind: Service
metadata:
  name: d1
spec:
  clusterIP: upd`)

		It("should successfully merge", func() {
			// this mutates updated
			err := apply.MergeObjectForUpdate(cur, upd)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should keep the original clusterIP", func() {
			ip, _, err := unstructured.NestedString(upd.Object, "spec", "clusterIP")
			Expect(err).NotTo(HaveOccurred())
			Expect(ip).To(Equal("cur"))
		})
	})

	Context("when given a ServiceAccount", func() {
		cur := unstructuredFromYaml(`
apiVersion: v1
kind: ServiceAccount
metadata:
  name: d1
  annotations:
    a: cur
secrets:
- foo`)

		upd := unstructuredFromYaml(`
apiVersion: v1
kind: ServiceAccount
metadata:
  name: d1
  annotations:
    b: upd`)

		It("should successfully merge", func() {
			// this mutates updated
			err := apply.MergeObjectForUpdate(cur, upd)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should keep original secrets after merging", func() {
			s, ok, err := unstructured.NestedSlice(upd.Object, "secrets")
			Expect(err).NotTo(HaveOccurred())
			Expect(ok).To(BeTrue())
			Expect(s).To(ConsistOf("foo"))
		})
	})

	Context("when merging an empty Deployment into an empty Deployment", func() {
		cur := unstructuredFromYaml(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: d1`)

		upd := unstructuredFromYaml(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: d1`)

		It("should successfully merge", func() {
			// this mutates updated
			err := apply.MergeObjectForUpdate(cur, upd)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should stay empty", func() {
			Expect(upd.GetLabels()).To(BeEmpty())
		})
	})

	Context("when merging a non-empty Deployment into an empty Deployment", func() {
		cur := unstructuredFromYaml(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: d1`)

		upd := unstructuredFromYaml(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: d1
  labels:
    a: upd
    c: upd
  annotations:
    a: upd
    c: upd`)

		It("should successfully merge", func() {
			// this mutates updated
			err := apply.MergeObjectForUpdate(cur, upd)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should use values from the updating Deployment", func() {
			Expect(upd.GetLabels()).To(Equal(map[string]string{
				"a": "upd",
				"c": "upd",
			}))

			Expect(upd.GetAnnotations()).To(Equal(map[string]string{
				"a": "upd",
				"c": "upd",
			}))
		})
	})

	Context("when merging an empty Deployment into a non-empty Deployment", func() {
		cur := unstructuredFromYaml(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: d1
  labels:
    a: cur
    b: cur
  annotations:
    a: cur
    b: cur`)

		upd := unstructuredFromYaml(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: d1`)

		It("should successfully merge", func() {
			// this mutates updated
			err := apply.MergeObjectForUpdate(cur, upd)
			Expect(err).NotTo(HaveOccurred())
		})

		It("keep the original values and not overwrite them with pure void and emptiness", func() {
			Expect(upd.GetLabels()).To(Equal(map[string]string{
				"a": "cur",
				"b": "cur",
			}))

			Expect(upd.GetAnnotations()).To(Equal(map[string]string{
				"a": "cur",
				"b": "cur",
			}))
		})
	})
})

var _ = Describe("IsObjectSupported", func() {
	Context("when given a ServiceAccount with a secret", func() {
		sa := unstructuredFromYaml(`
apiVersion: v1
kind: ServiceAccount
metadata:
  name: d1
  annotations:
    a: cur
secrets:
- foo`)

		It("should return an error", func() {
			err := apply.IsObjectSupported(sa)
			Expect(err).To(MatchError(ContainSubstring("cannot create ServiceAccount with secrets")))
		})
	})
})

// unstructuredFromYaml creates an unstructured object from a raw yaml string
func unstructuredFromYaml(obj string) *unstructured.Unstructured {
	buf := bytes.NewBufferString(obj)
	decoder := yaml.NewYAMLOrJSONDecoder(buf, 4096)

	u := unstructured.Unstructured{}
	err := decoder.Decode(&u)
	if err != nil {
		panic(fmt.Sprintf("failed to parse test yaml: %v", err))
	}

	return &u
}
