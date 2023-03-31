package test

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"

	. "github.com/onsi/gomega"

	promApi "github.com/prometheus/client_golang/api"
	promApiv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	promConfig "github.com/prometheus/common/config"
	authenticationv1 "k8s.io/api/authentication/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"kubevirt.io/client-go/kubecli"

	testenv "github.com/kubevirt/cluster-network-addons-operator/test/env"
)

var portForwardCmd *exec.Cmd

type promClient struct {
	client     promApiv1.API
	sourcePort int
	namespace  string
}

const prometheusMonitoringNamespace string = "monitoring"

func newPromClient(sourcePort int, monitoringNs string) *promClient {
	prometheusClient := &promClient{
		sourcePort: sourcePort,
		namespace:  monitoringNs,
	}

	prometheusClient.client = initializePromClient(prometheusClient.getPrometheusUrl(), prometheusClient.getAuthorizationTokenForPrometheus())

	return prometheusClient
}

func (p *promClient) checkForAlert(alertName string) {
	Eventually(func() *promApiv1.Alert {
		alert := p.getAlertByName(alertName)
		return alert
	}, 2*time.Minute, 1*time.Second).ShouldNot(BeNil(), fmt.Sprintf("alert %s not fired", alertName))
}

func (p *promClient) checkNoAlertsFired() {
	Consistently(func() (alerts []promApiv1.Alert) {
		alertsResult, err := p.client.Alerts(context.TODO())
		Expect(err).ShouldNot(HaveOccurred())
		return alertsResult.Alerts
	}, 2*time.Minute, 10*time.Second).Should(BeEmpty(), "unexpected alerts fired")
}

func (p *promClient) getPrometheusUrl() string {
	return fmt.Sprintf("http://localhost:%d", p.sourcePort)
}

func (p *promClient) getAlertByName(alertName string) *promApiv1.Alert {
	alerts, err := p.client.Alerts(context.TODO())
	Expect(err).ShouldNot(HaveOccurred())

	for _, alert := range alerts.Alerts {
		if string(alert.Labels["alertname"]) == alertName {
			return &alert
		}
	}
	return nil
}

func (p *promClient) getAuthorizationTokenForPrometheus() string {
	virtCli, err := kubecli.GetKubevirtClientFromFlags("", os.Getenv("KUBECONFIG"))
	Expect(err).NotTo(HaveOccurred())

	var token string
	Eventually(func() bool {
		treq, err := virtCli.CoreV1().ServiceAccounts(p.namespace).CreateToken(
			context.TODO(),
			"prometheus-k8s",
			&authenticationv1.TokenRequest{
				Spec: authenticationv1.TokenRequestSpec{
					// Avoid specifying any audiences so that the token will be
					// issued for the default audience of the issuer.
				},
			},
			metav1.CreateOptions{},
		)
		if err != nil {
			return false
		}
		token = treq.Status.Token
		return true
	}, 10*time.Second, time.Second).Should(BeTrue())
	return token
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

	PromClient := promApiv1.NewAPI(c)
	return PromClient
}

func checkMonitoringRoleBindingConfig(roleBindingName, namespace string) error {
	monitoringRoleBinding := rbacv1.RoleBinding{}
	err := testenv.Client.Get(context.TODO(), types.NamespacedName{Name: roleBindingName, Namespace: namespace}, &monitoringRoleBinding)
	if err != nil {
		return err
	}

	for _, subject := range monitoringRoleBinding.Subjects {
		var svcAccnt v1.ServiceAccount
		err = testenv.Client.Get(context.TODO(), types.NamespacedName{Name: subject.Name, Namespace: subject.Namespace}, &svcAccnt)
		if err != nil {
			return err
		}
	}

	return nil
}
