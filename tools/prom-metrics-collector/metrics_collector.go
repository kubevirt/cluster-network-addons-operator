/*
 * This file is part of the KubeVirt project
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2023 Red Hat, Inc.
 *
 */

package main

import (
	"github.com/kubevirt/cluster-network-addons-operator/pkg/monitoring"
	"github.com/kubevirt/cluster-network-addons-operator/tools/metrics-parser"
	"github.com/kubevirt/monitoring/pkg/metrics/parser"
	dto "github.com/prometheus/client_model/go"
)

// This should be used only for very rare cases where the naming conventions that are explained in the best practices:
// https://sdk.operatorframework.io/docs/best-practices/observability-best-practices/#metrics-guidelines
// should be ignored.
var excludedMetrics = map[string]struct{}{}

// Read the metrics and parse them to a MetricFamily
func ReadMetrics() []*dto.MetricFamily {
	cnaoMetrics := metricsparser.ReadFromPrometheusCR()
	cnaoMetrics = metricsparser.MetricsOptsToMetricList(monitoring.MetricsOptsList, cnaoMetrics)

	metricsList := make([]parser.Metric, len(cnaoMetrics))
	var metricFamily []*dto.MetricFamily
	for i, cnaoMetric := range cnaoMetrics {
		metricsList[i] = parser.Metric{
			Name: cnaoMetric.Name,
			Help: cnaoMetric.Description,
			Type: cnaoMetric.MType,
		}
	}
	for _, hcoMetric := range metricsList {
		// Remove ignored metrics from all rules
		if _, isExcludedMetric := excludedMetrics[hcoMetric.Name]; !isExcludedMetric {
			mf := parser.CreateMetricFamily(hcoMetric)
			metricFamily = append(metricFamily, mf)
		}
	}
	return metricFamily
}
