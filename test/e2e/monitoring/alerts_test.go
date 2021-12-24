package test

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	promApiv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"time"
)

var _ = Context("Prometheus Alerts", func() {

	Context("when installed from master release", func() {

		BeforeEach(func() {

		})

		AfterEach(func() {
			ScaleDeployment("cluster-network-addons-operator", 1)
		})

		It("should be true", func() {
			ScaleDeployment("cluster-network-addons-operator", 0)
			WaitForAlert("CnaoDown")
		})

	})
})

func WaitForAlert(alertName string) {
	Eventually(func() *promApiv1.Alert{
		alert := GetAlertByName(alertName)
		return alert
	}, 120 * time.Second, 1*time.Second).ShouldNot(BeNil())
}


func ScaleDeployment(deploymentName string, replicas int32) {
	Eventually(func() error{
		var deployment appsv1.Deployment
		var err error
		err = framework.Global.Client.Get(context.TODO(), types.NamespacedName{Name: deploymentName,Namespace: "cluster-network-addons"}, &deployment)
		if err != nil{
			return err
		}

		deployment.Spec.Replicas = &replicas
		return framework.Global.Client.Update(context.TODO(), &deployment)
	},60*time.Second, 1*time.Second).Should(BeNil())
}
