package rules_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/rhobs/operator-observability-toolkit/pkg/testutil"

	"github.com/kubevirt/cluster-network-addons-operator/pkg/monitoring/rules"
)

var _ = Describe("Rules Validation", func() {
	var linter *testutil.Linter

	BeforeEach(func() {
		Expect(rules.SetupRules("")).To(Succeed())
		linter = testutil.New()
	})

	It("Should validate alerts", func() {
		linter.AddCustomAlertValidations(
			testutil.ValidateAlertNameLength,
			testutil.ValidateAlertRunbookURLAnnotation,
			testutil.ValidateAlertHealthImpactLabel,
			testutil.ValidateAlertPartOfAndComponentLabels)

		alerts := rules.ListAlerts()
		problems := linter.LintAlerts(alerts)
		Expect(problems).To(BeEmpty())
	})

	It("Should validate recording rules", func() {
		recordingRules := rules.ListRecordingRules()
		problems := linter.LintRecordingRules(recordingRules)
		Expect(problems).To(BeEmpty())
	})
})
