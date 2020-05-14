package apply_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/kubevirt/cluster-network-addons-operator/pkg/apply"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/util/k8s"
)

var _ = Describe("MergeObjectForUpdate", func() {
	// Namespaces use the "generic" logic; deployments and services
	// have custom logic
	Context("when given a generic object (Namespace)", func() {
		cur := k8s.UnstructuredFromYaml(`
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

		upd := k8s.UnstructuredFromYaml(`
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
		cur := k8s.UnstructuredFromYaml(`
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

		upd := k8s.UnstructuredFromYaml(`
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
		cur := k8s.UnstructuredFromYaml(`
apiVersion: v1
kind: Service
metadata:
  name: d1
spec:
  clusterIP: cur`)

		upd := k8s.UnstructuredFromYaml(`
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
		cur := k8s.UnstructuredFromYaml(`
apiVersion: v1
kind: ServiceAccount
metadata:
  name: d1
  annotations:
    a: cur
secrets:
- foo`)

		upd := k8s.UnstructuredFromYaml(`
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
		cur := k8s.UnstructuredFromYaml(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: d1`)

		upd := k8s.UnstructuredFromYaml(`
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
		cur := k8s.UnstructuredFromYaml(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: d1`)

		upd := k8s.UnstructuredFromYaml(`
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
		cur := k8s.UnstructuredFromYaml(`
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

		upd := k8s.UnstructuredFromYaml(`
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
		sa := k8s.UnstructuredFromYaml(`
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

var _ = Describe("MergeMetadataForUpdate", func() {
	Context("when given current unstructured and empty updated", func() {
		current := k8s.UnstructuredFromYaml(`
apiVersion: v1
kind: Deployment
metadata:
  name: foo
  creationTimestamp: 2019-06-12T13:49:20Z
  generation: 1
  resourceVersion: "439"
  selfLink: /apis/extensions/v1beta1/namespaces/kube-system/deployments/foo
  uid: e0ecf168-8d18-11e9-b398-525500d15501
`)
		updated := k8s.UnstructuredFromYaml(`
apiVersion: v1
kind: Deployment
metadata:
  name: foo`)

		It("should merge metadate from current to updated", func() {
			err := apply.MergeMetadataForUpdate(current, updated)
			Expect(err).ToNot(HaveOccurred())
			Expect(updated.GetCreationTimestamp()).To(Equal(current.GetCreationTimestamp()))
			Expect(updated.GetGeneration()).To(Equal(current.GetGeneration()))
			Expect(updated.GetResourceVersion()).To(Equal(current.GetResourceVersion()))
			Expect(updated.GetSelfLink()).To(Equal(current.GetSelfLink()))
			Expect(updated.GetUID()).To(Equal(current.GetUID()))
		})
	})
	Context("when current has non empty caBundle and update is empty", func() {
		template := `
apiVersion: admissionregistration.k8s.io/v1beta1
kind: MutatingWebhookConfiguration
metadata:
  name: nmstate
  labels:
    app: %s
  annotations: {}
webhooks:
  - name: nodenetworkconfigurationpolicies-mutate.nmstate.io
    clientConfig:
      %s
      service:
        name: nmstate-webhook
        namespace: nmstate
        path: "/nodenetworkconfigurationpolicies-mutate"
    rules:
      - operations: ["CREATE", "UPDATE"]
        apiGroups: ["*"]
        apiVersions: ["v1alpha1"]
        resources: ["nodenetworkconfigurationpolicies"]
  - name: nodenetworkconfigurationpolicies-status-mutate.nmstate.io
    clientConfig:
      %s
      service:
        name: nmstate-webhook
        namespace: nmstate
        path: "/nodenetworkconfigurationpolicies-status-mutate"
    rules:
      - operations: ["CREATE", "UPDATE"]
        apiGroups: ["*"]
        apiVersions: ["v1alpha1"]
        resources: ["nodenetworkconfigurationpolicies/status"]
  - name: nodenetworkconfigurationpolicies-timestamp-mutate.nmstate.io
    clientConfig:
      %s
      service:
        name: nmstate-webhook
        namespace: nmstate
        path: "/nodenetworkconfigurationpolicies-timestamp-mutate"
    rules:
      - operations: ["CREATE", "UPDATE"]
        apiGroups: ["*"]
        apiVersions: ["v1alpha1"]
        resources: ["nodenetworkconfigurationpolicies", "nodenetworkconfigurationpolicies/status"]
`
		var (
			updated *unstructured.Unstructured
			current *unstructured.Unstructured
		)
		BeforeEach(func() {
			updated = k8s.UnstructuredFromYaml(fmt.Sprintf(template, "kubernetes-nmstate-2", "", "", ""))
			current = k8s.UnstructuredFromYaml(fmt.Sprintf(template, "kubernetes-nmstate-1", "caBundle: ca1", "caBundle: ca2", "caBundle: ca3"))
		})
		It("should now overwrite existing caBundle in webhook configuration", func() {
			expected := k8s.UnstructuredFromYaml(fmt.Sprintf(template, "kubernetes-nmstate-2", "caBundle: ca1", "caBundle: ca2", "caBundle: ca3"))
			err := apply.MergeObjectForUpdate(current, updated)
			Expect(err).ToNot(HaveOccurred(), "should successfully execut merge function")
			Expect(*updated).To(Equal(*expected))

		})
	})
})
