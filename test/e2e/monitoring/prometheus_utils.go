package test

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

	. "github.com/onsi/gomega"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	promApi "github.com/prometheus/client_golang/api"
	promApiv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	promConfig "github.com/prometheus/common/config"

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
	var err error
	var secretName string
	sa := v1.ServiceAccount{}

	err = testenv.Client.Get(context.TODO(), types.NamespacedName{Name: "prometheus-k8s", Namespace: p.namespace}, &sa)
	Expect(err).NotTo(HaveOccurred())
	Expect(sa).ToNot(BeNil())

	for _, secret := range sa.Secrets {
		if strings.HasPrefix(secret.Name, "prometheus-k8s-token") {
			secretName = secret.Name
		}
	}
	Expect(secretName).NotTo(BeEmpty())

	var secret v1.Secret
	err = testenv.Client.Get(context.TODO(), types.NamespacedName{Name: secretName, Namespace: p.namespace}, &secret)
	Expect(err).NotTo(HaveOccurred())

	data, ok := secret.Data["token"]
	Expect(ok).To(BeTrue())

	return string(data)
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
