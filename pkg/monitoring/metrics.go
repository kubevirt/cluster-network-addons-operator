package monitoring

import (
	"regexp"
	"strconv"

	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const defaultMetricPort = 8080

func init() {
	metrics.Registry.MustRegister()
}

func GetMetricsAddress() string {
	return metrics.DefaultBindAddress
}

func GetMetricsPort() int32 {
	re := regexp.MustCompile(`(?m).*:(\d+)`)

	portString := re.ReplaceAllString(metrics.DefaultBindAddress, "$1")
	portInt64, err := strconv.ParseUint(portString, 10, 32)
	if err != nil {
		return defaultMetricPort
	}
	return int32(portInt64)
}
