package main

import (
	"fmt"

	"github.com/kubevirt/cluster-network-addons-operator/pkg/monitoring/rules"

	"github.com/rhobs/operator-observability-toolkit/pkg/docs"

	"github.com/kubevirt/cluster-network-addons-operator/pkg/monitoring/metrics"
)

const title = `Cluster Network Addons Operator Metrics`

func main() {
	if err := metrics.SetupMetrics(); err != nil {
		panic(err)
	}

	if err := rules.SetupRules("test"); err != nil {
		panic(err)
	}

	docsString := docs.BuildMetricsDocs(title, metrics.ListMetrics(), rules.ListRecordingRules())
	fmt.Print(docsString)
}
