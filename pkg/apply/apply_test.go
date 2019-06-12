package apply_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"context"

	"k8s.io/apimachinery/pkg/runtime"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/kubevirt/cluster-network-addons-operator/pkg/apply"
)

var _ = Describe("ApplyObject", func() {
	Context("when server is empty", func() {
		var client *k8sclient.Client
		BeforeEach(){
			objs := []runtime.Object{}
			client := fake.NewFakeClient(objs...)
		}
		
		Context("and new object is applied", func() {
			object := unstructuredFromYaml(`
apiVersion: apps/v1
kind: Deployment
metadata:
name: d1`)

			BeforeEach() {
				err := apply.ApplyObject(context.Background(), client, object)
				Expect(err).ToNot(HaveOccurred())
			}

			It("should create new object", func() {
				err = client.Get(context.Background(), types.NamespacedName{Name: "d1"}, &appsv1.Deployment{})
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Context("when server has object", func() {
		var client *k8sclient.Client
		var originalDeployment *appsv1.Deployment
		BeforeEach(){
			originalDeployment = unstructuredFromYaml(`
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
			objs := []runtime.Object{deployment}
			client := fake.NewFakeClient(objs...)
		}
		
		It("should update object", func() {
			object := unstructuredFromYaml(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: d1`)

			err := apply.ApplyObject(context.Background(), client, object)
			Expect(err).ToNot(HaveOccurred())

			err = client.Get(context.Background(), types.NamespacedName{Name: "d1"}, &appsv1.Deployment{})
			Expect(err).ToNot(HaveOccurred())
		})

		It()
	})
})
