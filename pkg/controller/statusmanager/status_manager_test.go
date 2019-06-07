package statusmanager

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"

	opv1alpha1 "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/v1alpha1"
)

var _ = Describe("updateCondition", func() {
	Context("when status has empty conditions", func() {
		status := &opv1alpha1.NetworkAddonsConfigStatus{}
		var testedStatus *opv1alpha1.NetworkAddonsConfigStatus

		Context("and condition is being set", func() {
			condition := opv1alpha1.NetworkAddonsCondition{
				Type:    opv1alpha1.NetworkAddonsConditionProgressing,
				Status:  corev1.ConditionFalse,
				Reason:  "bar",
				Message: "foo",
			}

			BeforeEach(func() {
				// Update condition changes status
				testedStatus = status.DeepCopy()
				updateCondition(testedStatus, condition)
			})

			It("should have length 1", func() {
				Expect(testedStatus.Conditions).To(HaveLen(1))
			})

			It("should have type of condition, status, message, reason", func() {
				Expect(testedStatus.Conditions[0].Type).To(Equal(condition.Type))
				Expect(testedStatus.Conditions[0].Status).To(Equal(condition.Status))
				Expect(testedStatus.Conditions[0].Reason).To(Equal(condition.Reason))
				Expect(testedStatus.Conditions[0].Message).To(Equal(condition.Message))
			})

			It("should have probe time and transition time equal", func() {
				Expect(testedStatus.Conditions[0].LastProbeTime.Time).To(Equal(testedStatus.Conditions[0].LastTransitionTime.Time))
			})
		})
	})
})
