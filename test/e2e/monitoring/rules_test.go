package test

import (
	"context"
	"fmt"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
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

		Context("CNAO alert rules", func() {
			var prometheusRule monitoringv1.PrometheusRule

			BeforeEach(func() {
				err := testenv.Client.Get(context.Background(), types.NamespacedName{Name: "prometheus-rules-cluster-network-addons-operator", Namespace: components.Namespace}, &prometheusRule)
				Expect(err).ToNot(HaveOccurred())
			})

			It("should have all the requried annotations", func() {
				for _, group := range prometheusRule.Spec.Groups {
					for _, rule := range group.Rules {
						if len(rule.Alert) > 0 {
							Expect(rule.Annotations).ToNot(BeNil())
							checkRequiredAnnotations(rule)
						}
					}
				}
			})

			It("should have all the requried labels", func() {
				for _, group := range prometheusRule.Spec.Groups {
					for _, rule := range group.Rules {
						if len(rule.Alert) > 0 {
							Expect(rule.Labels).ToNot(BeNil())
							checkRequiredLabels(rule)
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

func checkRequiredAnnotations(rule monitoringv1.Rule) {
	ExpectWithOffset(1, rule.Annotations).To(HaveKeyWithValue("summary", Not(BeEmpty())),
		"%s summary is missing or empty", rule.Alert)
	ExpectWithOffset(1, rule.Annotations).To(HaveKey("runbook_url"),
		"%s runbook_url is missing", rule.Alert)
	ExpectWithOffset(1, rule.Annotations).To(HaveKeyWithValue("runbook_url", ContainSubstring(rule.Alert)),
		"%s runbook_url doesn't include alert name", rule.Alert)

	resp, err := http.Head(rule.Annotations["runbook_url"])
	ExpectWithOffset(1, err).ToNot(HaveOccurred(), fmt.Sprintf("%s runbook is not available", rule.Alert))
	ExpectWithOffset(1, resp.StatusCode).Should(Equal(http.StatusOK), fmt.Sprintf("%s runbook is not available", rule.Alert))
}

func checkRequiredLabels(rule monitoringv1.Rule) {
	ExpectWithOffset(1, rule.Labels).To(HaveKeyWithValue("severity", BeElementOf("info", "warning", "critical")),
		"%s severity label is missing or not valid", rule.Alert)
	ExpectWithOffset(1, rule.Labels).To(HaveKeyWithValue("operator_health_impact", BeElementOf("none", "warning", "critical")),
		"%s operator_health_impact label is missing or not valid", rule.Alert)
	ExpectWithOffset(1, rule.Labels).To(HaveKeyWithValue("kubernetes_operator_part_of", "kubevirt"),
		"%s kubernetes_operator_part_of label is missing or not valid", rule.Alert)
	ExpectWithOffset(1, rule.Labels).To(HaveKeyWithValue("kubernetes_operator_component", "cluster-network-addons-operator"),
		"%s kubernetes_operator_component label is missing or not valid", rule.Alert)
}
