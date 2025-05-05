package k8s

import (
	"log"
	"regexp"

	cnaov1 "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/v1"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/names"
)

var labelAntiSelector *regexp.Regexp

func init() {
	var err error

	// Regex matching all non-allowed label characters
	labelAntiSelector, err = regexp.Compile("[^a-z0-9A-Z_.-]+")
	if err != nil {
		log.Fatal(err)
	}
}

// StringToLabel makes given string label-compatible
func StringToLabel(s string) string {
	// We need to remove characters that are not allowed
	s = labelAntiSelector.ReplaceAllString(s, "_")

	// We need to trim the string to allowed length
	if len(s) > 63 {
		s = s[0:63]
	}

	return s
}

// RelationLabels returns the list of the relationship labels
func RelationLabels() []string {
	return []string{
		names.COMPONENT_LABEL_KEY,
		names.PART_OF_LABEL_KEY,
		names.VERSION_LABEL_KEY,
		names.MANAGED_BY_LABEL_KEY,
	}
}

// RemovedLabels returns the list of the labels we remove for backward compatibility
func RemovedLabels() []string {
	labels := RelationLabels()
	labels = append(labels, []string{
		names.PrometheusLabelKey,
		names.KUBEMACPOOL_CONTROL_PLANE_KEY,
		cnaov1.GroupVersion.Group + "/version"}...)
	return labels
}
