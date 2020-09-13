package operations

import (
	"strings"

	"gopkg.in/yaml.v2"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
)

type ConfigInterface interface {
	GetConfig() map[string]interface{}
	CreateConfig(configSpec cnao.NetworkAddonsConfigSpec)
	UpdateConfig(configSpec cnao.NetworkAddonsConfigSpec)
	DeleteConfig()
	GetStatus() cnao.NetworkAddonsConfigStatus
}

// Convert NetworkAddonsConfig specification to a yaml format we would expect in a manifest
func configSpecToYaml(configSpec cnao.NetworkAddonsConfigSpec) string {
	manifest, err := yaml.Marshal(configSpec)
	if err != nil {
		panic(err)
	}

	manifestLines := strings.Split(string(manifest), "\n")

	// We don't want to show non-set (default) values, usually null. Try our best to filter those out.
	manifestLinesWithoutEmptyValues := []string{}
	for _, line := range manifestLines {
		// If root attribute (e.g. ImagePullPolicy) is set to default, drop it. If it
		// is a nested attribute (e.g. KubeMacPool's RangeEnd), keep it.
		rootAttributeSetToDefault := !strings.Contains(line, "  ") && (strings.Contains(line, ": \"\"") || strings.Contains(line, ": null"))
		if line != "" && !rootAttributeSetToDefault {
			manifestLinesWithoutEmptyValues = append(manifestLinesWithoutEmptyValues, line)
		}
	}

	// If any values has been set, return Spec in a nice YAML format
	if len(manifestLinesWithoutEmptyValues) > 0 {
		indentedManifest := strings.TrimSpace(strings.Join(manifestLinesWithoutEmptyValues, "\n"))
		return indentedManifest
	}

	// Note that it is empty otherwise
	return "Empty Spec"
}
