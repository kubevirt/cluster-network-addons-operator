package check

import (
	"fmt"
	"io/ioutil"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type Component struct {
	ComponentName                string
	ClusterRole                  string
	ClusterRoleBinding           string
	SecurityContextConstraints   string
	DaemonSets                   []string
	Deployments                  []string
	Secret                       string
	MutatingWebhookConfiguration string
}

var (
	KubeMacPoolComponent = Component{
		ComponentName:                "KubeMacPool",
		ClusterRole:                  "kubemacpool-manager-role",
		ClusterRoleBinding:           "kubemacpool-manager-rolebinding",
		Deployments:                  []string{"kubemacpool-mac-controller-manager", "kubemacpool-cert-manager"},
		MutatingWebhookConfiguration: "kubemacpool-mutator",
	}
	LinuxBridgeComponent = Component{
		ComponentName:              "Linux Bridge",
		ClusterRole:                "bridge-marker-cr",
		ClusterRoleBinding:         "bridge-marker-crb",
		SecurityContextConstraints: "linux-bridge",
		DaemonSets: []string{
			"bridge-marker",
			"kube-cni-linux-bridge-plugin",
		},
	}
	MultusComponent = Component{
		ComponentName:              "Multus",
		ClusterRole:                "multus",
		ClusterRoleBinding:         "multus",
		SecurityContextConstraints: "multus",
		DaemonSets:                 []string{"multus"},
	}
	NMStateComponent = Component{
		ComponentName:              "NMState",
		ClusterRoleBinding:         "nmstate-handler",
		ClusterRole:                "nmstate-handler",
		SecurityContextConstraints: "nmstate",
		DaemonSets: []string{
			"nmstate-handler",
		},
		Deployments: []string{
			"nmstate-webhook",
			"nmstate-cert-manager",
		},
		MutatingWebhookConfiguration: "nmstate",
	}
	OvsComponent = Component{
		ComponentName:              "Ovs",
		ClusterRole:                "ovs-cni-marker-cr",
		ClusterRoleBinding:         "ovs-cni-marker-crb",
		SecurityContextConstraints: "ovs-cni-marker",
		DaemonSets: []string{
			"ovs-cni-amd64",
		},
	}
	MacvtapComponent = Component{
		ComponentName:              "Macvtap",
		ClusterRole:                "",
		ClusterRoleBinding:         "",
		SecurityContextConstraints: "macvtap-cni",
		DaemonSets: []string{
			"macvtap-cni",
		},
	}
	AllComponents = []Component{
		KubeMacPoolComponent,
		LinuxBridgeComponent,
		MultusComponent,
		NMStateComponent,
		OvsComponent,
		MacvtapComponent,
	}
)

type ComponentUpdatePolicy string

const (
	Tagged ComponentUpdatePolicy = "tagged"
	Static ComponentUpdatePolicy = "static"
	Latest ComponentUpdatePolicy = "latest"
)

type ComponentsConfig struct {
	Components map[string]ComponentSource `yaml:"components"`
}

type ComponentSource struct {
	Url          string                `yaml:"url"`
	Commit       string                `yaml:"commit"`
	Branch       string                `yaml:"branch"`
	UpdatePolicy ComponentUpdatePolicy `yaml:"update-policy"`
	Metadata     string                `yaml:"metadata"`
}

func GetComponentSource(component string) (ComponentSource, error) {
	componentsConfig, err := parseComponentsYaml("components.yaml")
	if err != nil {
		return ComponentSource{}, errors.Wrapf(err, "Failed to get components config")
	}

	componentSource, ok := componentsConfig.Components[component]
	if !ok {
		return ComponentSource{}, errors.Wrapf(err, "Failed to get component %s", component)
	}

	return componentSource, nil
}

func parseComponentsYaml(componentsConfigPath string) (ComponentsConfig, error) {
	config := ComponentsConfig{}

	componentsData, err := ioutil.ReadFile(componentsConfigPath)
	if err != nil {
		return ComponentsConfig{}, errors.Wrapf(err, "Failed to open file %s", componentsConfigPath)
	}

	err = yaml.Unmarshal(componentsData, &config)
	if err != nil {
		return ComponentsConfig{}, errors.Wrapf(err, "Failed to Unmarshal %s", componentsConfigPath)
	}

	if len(config.Components) == 0 {
		return ComponentsConfig{}, fmt.Errorf("Failed to Unmarshal %s. Output is empty", componentsConfigPath)
	}

	return config, nil
}
