package test

import (
	"context"
	"fmt"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
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
							checkForRunbookURL(rule)
							checkForSummary(rule)
						}
					}
				}
			})

			It("should have all the requried labels", func() {
				for _, group := range prometheusRule.Spec.Groups {
					for _, rule := range group.Rules {
						if len(rule.Alert) > 0 {
							Expect(rule.Labels).ToNot(BeNil())
							checkForSeverityLabel(rule)
							checkForHealthImpactLabel(rule)
							checkForPartOfLabel(rule)
							checkForComponentLabel(rule)
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

func checkForRunbookURL(rule monitoringv1.Rule) {
	url, ok := rule.Annotations["runbook_url"]
	ExpectWithOffset(1, ok).To(BeTrue(), fmt.Sprintf("%s does not have runbook_url annotation", rule.Alert))
	resp, err := http.Head(url)
	ExpectWithOffset(1, err).ToNot(HaveOccurred(), fmt.Sprintf("%s runbook is not available", rule.Alert))
	ExpectWithOffset(1, resp.StatusCode).Should(Equal(http.StatusOK), fmt.Sprintf("%s runbook is not available", rule.Alert))
}

func checkForSummary(rule monitoringv1.Rule) {
	summary, ok := rule.Annotations["summary"]
	ExpectWithOffset(1, ok).To(BeTrue(), fmt.Sprintf("%s does not have summary annotation", rule.Alert))
	ExpectWithOffset(1, summary).ToNot(BeEmpty(), fmt.Sprintf("%s has an empty summary", rule.Alert))
}

func checkForSeverityLabel(rule monitoringv1.Rule) {
	severity, ok := rule.Labels["severity"]
	ExpectWithOffset(1, ok).To(BeTrue(), fmt.Sprintf("%s does not have severity label", rule.Alert))
	ExpectWithOffset(1, severity).To(BeElementOf("info", "warning", "critical"), fmt.Sprintf("%s severity label is not valid", rule.Alert))
}

func checkForHealthImpactLabel(rule monitoringv1.Rule) {
	operatorHealthImpact, ok := rule.Labels["operator_health_impact"]
	ExpectWithOffset(1, ok).To(BeTrue(), fmt.Sprintf("%s does not have operator_health_impact label", rule.Alert))
	ExpectWithOffset(1, operatorHealthImpact).To(BeElementOf("none", "warning", "critical"), fmt.Sprintf("%s operator_health_impact label is not valid", rule.Alert))
}

func checkForPartOfLabel(rule monitoringv1.Rule) {
	kubernetesOperatorPartOf, ok := rule.Labels["kubernetes_operator_part_of"]
	ExpectWithOffset(1, ok).To(BeTrue(), fmt.Sprintf("%s does not have kubernetes_operator_part_of label", rule.Alert))
	ExpectWithOffset(1, kubernetesOperatorPartOf).To(Equal("kubevirt"), fmt.Sprintf("%s kubernetes_operator_part_of label is not valid", rule.Alert))
}

func checkForComponentLabel(rule monitoringv1.Rule) {
	kubernetesOperatorComponent, ok := rule.Labels["kubernetes_operator_component"]
	ExpectWithOffset(1, ok).To(BeTrue(), fmt.Sprintf("%s does not have kubernetes_operator_component label", rule.Alert))
	ExpectWithOffset(1, kubernetesOperatorComponent).To(Equal("cluster-network-addons-operator"), fmt.Sprintf("%s kubernetes_operator_component label is not valid", rule.Alert))
}
