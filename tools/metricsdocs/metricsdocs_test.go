package main

import (
	"io/ioutil"
	"os"
	"sort"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/kubevirt/cluster-network-addons-operator/pkg/monitoring"
)

var _ = Describe("Test metricsdocs", func() {
	var metricsOptsList = map[monitoring.MetricsKey]monitoring.MetricsOpts{
		"metric2": {
			Name: "kubevirt_metric_2",
			Help: "Metric 2 Help",
		},
		"metric1": {
			Name: "kubevirt_metric_1",
			Help: "Metric 1 Help",
		},
	}

	Describe("metricsOptsToMetricList function", func() {
		Context("when metricsOptsList is empty", func() {
			var metricsList metricList

			BeforeEach(func() {
				By("transforming metricsOpts to metricList")
				metricsList = metricsOptsToMetricList(map[monitoring.MetricsKey]monitoring.MetricsOpts{})
			})

			It("should return an empty metricList", func() {
				Expect(metricsList).To(HaveLen(0))
			})
		})

		Context("when metricsOptsList is not empty", func() {
			var metricsList metricList

			BeforeEach(func() {
				By("transforming metricsOpts to metricList")
				metricsList = metricsOptsToMetricList(metricsOptsList)
				sort.Sort(metricsList)
			})

			It("should return a metricList with the same number of elements as metricsOptsList", func() {
				Expect(metricsList).To(HaveLen(len(metricsOptsList)))
			})

			It("should return a metricList sorted by metric name", func() {
				metricA := metricsList[0]
				metricB := metricsList[1]
				Expect(metricA.name <= metricB.name).To(BeTrue())
			})
		})
	})

	Describe("metricsdocs", func() {
		var metricsList metricList
		var stdout string

		BeforeEach(func() {
			By("writing metrics to file")

			r, w, _ := os.Pipe()
			tmp := os.Stdout
			defer func() {
				os.Stdout = tmp
			}()
			os.Stdout = w
			go func() {
				metricsList = metricsOptsToMetricList(metricsOptsList)
				sort.Sort(metricsList)
				writeToFile(metricsList)
				w.Close()
			}()
			b, _ := ioutil.ReadAll(r)
			stdout = string(b)
		})

		It("should create a file with the header", func() {
			Expect(stdout).To(ContainSubstring(title))
		})

		It("should create a file with the footer", func() {
			Expect(stdout).To(ContainSubstring(footer))
		})

		It("should create a file with all metrics and their descriptions", func() {
			Expect(stdout).To(
				ContainSubstring("### " + metricsOptsList["metric1"].Name + "\n" + metricsOptsList["metric1"].Help),
			)

			Expect(stdout).To(
				ContainSubstring("### " + metricsOptsList["metric2"].Name + "\n" + metricsOptsList["metric2"].Help),
			)
		})
	})
})
