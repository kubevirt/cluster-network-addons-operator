package network

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	osv1 "github.com/openshift/api/operator/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
)

var _ = Describe("Testing kubevirt ipam controller", func() {
	Context("Render KubevirtIpamController", func() {
		conf := &cnao.NetworkAddonsConfigSpec{ImagePullPolicy: v1.PullAlways, Multus: &cnao.Multus{}, KubevirtIpamController: &cnao.KubevirtIpamController{}, PlacementConfiguration: &cnao.PlacementConfiguration{Workloads: &cnao.Placement{}}}
		manifestDir := "../../data"
		openshiftNetworkConf := &osv1.Network{}
		clusterInfo := &ClusterInfo{}
		expectedGroupVersionKind := schema.GroupVersionKind{
			Group:   "k8s.cni.cncf.io",
			Version: "v1",
			Kind:    "NetworkAttachmentDefinition",
		}
		const expectedName = "primary-udn-kubevirt-binding"

		It("should add the primary-udn network-attach-def obj", func() {
			objs, err := Render(conf, manifestDir, openshiftNetworkConf, clusterInfo)
			Expect(err).NotTo(HaveOccurred())
			Expect(objs).NotTo(BeEmpty())

			Expect(objs).To(ContainElement(
				SatisfyAll(
					WithTransform(func(obj *unstructured.Unstructured) string {
						return obj.GetName()
					}, Equal(expectedName)),
					WithTransform(func(obj *unstructured.Unstructured) schema.GroupVersionKind {
						return obj.GetObjectKind().GroupVersionKind()
					}, Equal(expectedGroupVersionKind)),
				),
			))
		})
	})
})
