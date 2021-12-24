package test

import (
	"context"
	"crypto/tls"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"net/http"
	"strings"
	"testing"

	ginkgoreporters "kubevirt.io/qe-tools/pkg/ginkgo-reporters"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	f "github.com/operator-framework/operator-sdk/pkg/test"
	framework "github.com/operator-framework/operator-sdk/pkg/test"

	"github.com/kubevirt/cluster-network-addons-operator/pkg/apis"
	cnaov1 "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/v1"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/components"
	cnaoreporter "github.com/kubevirt/cluster-network-addons-operator/test/reporter"

	promApi "github.com/prometheus/client_golang/api"
	promApiv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	promConfig "github.com/prometheus/common/config"
)

var promClient promApiv1.API

func TestMain(m *testing.M) {
	f.MainEntry(m)
}

func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	reporters := make([]Reporter, 0)
	reporters = append(reporters, cnaoreporter.New("test_logs/e2e/monitoring", components.Namespace))
	if ginkgoreporters.JunitOutput != "" {
		reporters = append(reporters, ginkgoreporters.NewJunitReporter())
	}
	RunSpecsWithDefaultAndCustomReporters(t, "Monitoring E2E Test Suite", reporters)

}

var _ = BeforeSuite(func() {
	By("Adding custom resource scheme to framework")
	err := framework.AddToFrameworkScheme(apis.AddToScheme, &cnaov1.NetworkAddonsConfigList{})
	Expect(err).ToNot(HaveOccurred())

	Expect(framework.AddToFrameworkScheme(monitoringv1.AddToScheme, &monitoringv1.ServiceMonitorList{})).To(Succeed())
	Expect(framework.AddToFrameworkScheme(monitoringv1.AddToScheme, &monitoringv1.PrometheusRuleList{})).To(Succeed())

	monitoringNs := getMonitoringNamespace()
	promClient = initializePromClient(getPrometheusUrl(monitoringNs), getAuthorizationTokenForPrometheus(monitoringNs))
})

var _ = AfterEach(func() {
	By("Performing cleanup")
})



func getMonitoringNamespace() string {
	// TODO: Implement logic for OpenShift
	return "monitoring"
}

func GetAlertByName(alertName string) *promApiv1.Alert {
	alerts, err := promClient.Alerts(context.TODO())
	Expect(err).ShouldNot(HaveOccurred())

	for _, alert := range alerts.Alerts {
		if string(alert.Labels["alertname"]) == alertName {
			return &alert
		}
	}
	return nil
}

func initializePromClient(prometheusUrl string, token string) promApiv1.API {
	defaultRoundTripper := promApi.DefaultRoundTripper
	tripper := defaultRoundTripper.(*http.Transport)
	tripper.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	c, err := promApi.NewClient(promApi.Config{
		Address:      prometheusUrl,
		RoundTripper: promConfig.NewAuthorizationCredentialsRoundTripper("Bearer", promConfig.Secret(token), defaultRoundTripper),
	})

	Expect(err).ToNot(HaveOccurred())

	promClient := promApiv1.NewAPI(c)
	return promClient
}

func getAuthorizationTokenForPrometheus(monitoringNs string) string {
	var err error
	var sa v1.ServiceAccount
	var secretName string

	err = framework.Global.Client.Get(context.TODO(),types.NamespacedName{Name: "prometheus-k8s", Namespace: monitoringNs}, &sa)
	Expect(err).NotTo(HaveOccurred())

	for _, secret := range sa.Secrets {
		if strings.HasPrefix(secret.Name, "prometheus-k8s-token") {
			secretName = secret.Name
		}
	}
	Expect(secretName).NotTo(BeEmpty())

	var secret v1.Secret
	err = framework.Global.Client.Get(context.TODO(),types.NamespacedName{Name: secretName,Namespace: monitoringNs}, &secret)
	Expect(err).NotTo(HaveOccurred())

	data, ok := secret.Data["token"]
	Expect(ok).To(BeTrue())

	return string(data)
}

func getPrometheusUrl(monitoringNs string) string {
	// TODO: Implement logic for OpenShift
	return "http://localhost:9090"
}