package test

import (
	"context"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	"k8s.io/apimachinery/pkg/types"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/components"
	. "github.com/kubevirt/cluster-network-addons-operator/test/check"
	testenv "github.com/kubevirt/cluster-network-addons-operator/test/env"
	. "github.com/kubevirt/cluster-network-addons-operator/test/operations"
)

var _ = Context("Prometheus Rules", func() {

	Context("when networkaddonsconfig CR is deployed", func() {

		BeforeEach(func() {
			By("delpoying CNAO CR with at least one component")
			gvk := GetCnaoV1GroupVersionKind()
			configSpec := cnao.NetworkAddonsConfigSpec{
				MacvtapCni: &cnao.MacvtapCni{},
			}
			CreateConfig(gvk, configSpec)
			CheckConfigCondition(gvk, ConditionAvailable, ConditionTrue, 15*time.Minute, CheckDoNotRepeat)
		})

		Context("CNAO rules", func() {
			It("should have available runbook URLs", func() {
				prometheusRule := monitoringv1.PrometheusRule{}
				err := testenv.Client.Get(context.Background(), types.NamespacedName{Name: "prometheus-rules-cluster-network-addons-operator", Namespace: components.Namespace}, &prometheusRule)
				Expect(err).ToNot(HaveOccurred())
				for _, group := range prometheusRule.Spec.Groups {
					for _, rule := range group.Rules {
						if len(rule.Alert) > 0 {
							Expect(rule.Annotations).ToNot(BeNil())
							url, ok := rule.Annotations["runbook_url"]
							Expect(ok).To(BeTrue())
							resp, err := http.Head(url)
							Expect(err).ToNot(HaveOccurred())
							Expect(resp.StatusCode).Should(Equal(http.StatusOK))
						}
					}
				}
			})
		})

		AfterEach(func() {
			By("removing CNAO CR")
			gvk := GetCnaoV1GroupVersionKind()
			if GetConfig(gvk) != nil {
				DeleteConfig(gvk)
			}
		})
	})
})
