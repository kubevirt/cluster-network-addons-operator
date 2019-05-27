package k8s_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"encoding/json"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	k8sutil "github.com/kubevirt/cluster-network-addons-operator/pkg/util/k8s"
)

var _ = Describe("ToUnstructured", func() {
	Context("when a valid Kubernetes object is passed as a parameter", func() {
		pod := &apiv1.Pod{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Pod",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "foo",
			},
			Spec: apiv1.PodSpec{
				Containers: []apiv1.Container{
					{
						Name:  "bar",
						Image: "bar",
					},
				},
			},
		}
		It("should successfully convert it to Unstructured", func() {
			unstructuredPod, err := k8sutil.ToUnstructured(pod)
			Expect(err).NotTo(HaveOccurred())

			// Convert back to Pod object to check that no data was lost
			unstructuredPodJSON, err := json.Marshal(unstructuredPod)
			Expect(err).NotTo(HaveOccurred())
			structuredPod := &apiv1.Pod{}
			err = json.Unmarshal(unstructuredPodJSON, structuredPod)
			Expect(err).NotTo(HaveOccurred())

			Expect(structuredPod.Name).To(Equal("foo"))
			Expect(structuredPod.Spec.Containers[0].Name).To(Equal("bar"))
			Expect(structuredPod.Spec.Containers[0].Image).To(Equal("bar"))
		})
	})
})
