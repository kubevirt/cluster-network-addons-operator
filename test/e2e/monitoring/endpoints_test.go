package test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
	. "github.com/kubevirt/cluster-network-addons-operator/test/check"
	. "github.com/kubevirt/cluster-network-addons-operator/test/operations"
)

var _ = Context("Prometheus Endpoints", func() {
	Context("when deploying components and checking CNAO prometheus endpoint", func() {
		gvk := GetCnaoV1GroupVersionKind()
		type prometheusScrapeParams struct {
			configSpec             cnao.NetworkAddonsConfigSpec
			expectedMetricValueMap map[string]string
		}
		DescribeTable("and checking scraped data",
			func(p prometheusScrapeParams) {
				By("deploying the configured NetworkAddonsConfigSpec")
				CreateConfig(gvk, p.configSpec)
				CheckConfigCondition(gvk, ConditionAvailable, ConditionTrue, 15*time.Minute, CheckDoNotRepeat)

				Eventually(func() error {
					By("scraping the monitoring endpoint")
					epAddress, epPort, err := GetMonitoringEndpoint()
					if err != nil {
						return err
					}

					scrapedData, err := ScrapeEndpointAddress(epAddress, epPort)
					Expect(err).ToNot(HaveOccurred())

					By("comparing the scraped Data to the expected metrics' values")
					for metricName, expectedValue := range p.expectedMetricValueMap {
						metricEntry := FindMetric(scrapedData, metricName)
						Expect(metricEntry).ToNot(BeEmpty(), fmt.Sprintf("metric %s does not appear in endpoint scrape", metricName))

						if metricEntry != fmt.Sprintf("%s %s", metricName, expectedValue) {
							return fmt.Errorf("metric %s does not have the expected value %s", metricName, expectedValue)
						}
					}
					return nil
				}, 3*time.Minute, time.Minute).Should(Succeed(), "Should scrape the correct metrics")
			},
			Entry("should report the expected metrics when deploying all components", prometheusScrapeParams{
				configSpec: cnao.NetworkAddonsConfigSpec{
					LinuxBridge: &cnao.LinuxBridge{},
					KubeMacPool: &cnao.KubeMacPool{},
					Ovs:         &cnao.Ovs{},
					MacvtapCni:  &cnao.MacvtapCni{},
				},
				expectedMetricValueMap: map[string]string{
					"kubevirt_cnao_cr_ready":                "1",
					"kubevirt_cnao_cr_kubemacpool_deployed": "1",
				},
			}),
			Entry("should report the expected metrics when deploying all components but kubemacpool", prometheusScrapeParams{
				configSpec: cnao.NetworkAddonsConfigSpec{
					LinuxBridge: &cnao.LinuxBridge{},
					Ovs:         &cnao.Ovs{},
					MacvtapCni:  &cnao.MacvtapCni{},
				},
				expectedMetricValueMap: map[string]string{
					"kubevirt_cnao_cr_ready":                "1",
					"kubevirt_cnao_cr_kubemacpool_deployed": "0",
				},
			}),
		)
	})
})
