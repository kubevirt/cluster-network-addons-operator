package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	osv1 "github.com/openshift/api/operator/v1"
	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/components"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/network"
	k8sutil "github.com/kubevirt/cluster-network-addons-operator/pkg/util/k8s"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type config struct {
	Platform         string                       `json:"platform"`
	Namespace        string                       `json:"namespace"`
	Cluster          clusterConfig                `json:"cluster,omitempty"`
	Images           map[string]string            `json:"images,omitempty"`
	Spec             cnao.NetworkAddonsConfigSpec `json:"spec"`
	OpenShiftNetwork *osv1.Network                `json:"openShiftNetwork,omitempty"`
}

type clusterConfig struct {
	SCCAvailable        bool `json:"sccAvailable,omitempty"`
	MonitoringAvailable bool `json:"monitoringAvailable,omitempty"`
	IsSingleReplica     bool `json:"isSingleReplica,omitempty"`
}

var imageEnvironment = map[string]struct {
	environment  string
	defaultImage string
}{
	"multus":                          {"MULTUS_IMAGE", components.MultusImageDefault},
	"multusDynamicNetworksController": {"MULTUS_DYNAMIC_NETWORKS_CONTROLLER_IMAGE", components.MultusDynamicNetworksImageDefault},
	"linuxBridge":                     {"LINUX_BRIDGE_IMAGE", components.LinuxBridgeCniImageDefault},
	"linuxBridgeMarker":               {"LINUX_BRIDGE_MARKER_IMAGE", components.LinuxBridgeMarkerImageDefault},
	"ovs":                             {"OVS_CNI_IMAGE", components.OvsCniImageDefault},
	"kubeMacPool":                     {"KUBEMACPOOL_IMAGE", components.KubeMacPoolImageDefault},
	"macvtap":                         {"MACVTAP_CNI_IMAGE", components.MacvtapCniImageDefault},
	"kubevirtIpamController":          {"KUBEVIRT_IPAM_CONTROLLER_IMAGE", components.KubevirtIpamControllerImageDefault},
	"passtBindingCni":                 {"PASST_BINDING_CNI_IMAGE", ""},
}

func main() {
	input := flag.String("input", "", "YAML or JSON render configuration")
	outputDir := flag.String("output-dir", "rendered", "directory for per-component output")
	hco := flag.Bool("hco", false, "also generate Go struct literals for HCO")
	manifestDir := flag.String("manifest-dir", "data", "CNAO component manifest directory")
	flag.Parse()

	if *input == "" {
		fatal(errors.New("-input is required"))
	}
	if err := run(*input, *outputDir, *hco, *manifestDir); err != nil {
		fatal(err)
	}
}

func run(inputPath, outputDir string, hco bool, manifestDir string) error {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return errors.Wrap(err, "read input")
	}
	var cfg config
	if err := yaml.UnmarshalStrict(data, &cfg); err != nil {
		return errors.Wrap(err, "parse input")
	}

	if cfg.Platform != "kubernetes" && cfg.Platform != "openshift" {
		return fmt.Errorf("platform must be kubernetes or openshift, got %q", cfg.Platform)
	}
	if cfg.Namespace == "" {
		return errors.New("namespace is required")
	}
	if cfg.Platform == "kubernetes" && cfg.OpenShiftNetwork != nil {
		return errors.New("openShiftNetwork is only valid for the openshift platform")
	}
	if cfg.Platform == "openshift" && cfg.OpenShiftNetwork == nil {
		cfg.OpenShiftNetwork = &osv1.Network{}
	}

	if err := os.Setenv("OPERAND_NAMESPACE", cfg.Namespace); err != nil {
		return errors.Wrap(err, "set operand namespace")
	}
	for _, image := range imageEnvironment {
		if image.defaultImage == "" {
			continue
		}
		if err := os.Setenv(image.environment, image.defaultImage); err != nil {
			return errors.Wrapf(err, "set default image %s", image.environment)
		}
	}
	for name, value := range cfg.Images {
		image, ok := imageEnvironment[name]
		if !ok {
			return fmt.Errorf("unsupported image key %q", name)
		}
		if err := os.Setenv(image.environment, value); err != nil {
			return errors.Wrapf(err, "set image %s", name)
		}
	}

	if cfg.Spec.PlacementConfiguration == nil {
		cfg.Spec.PlacementConfiguration = &cnao.PlacementConfiguration{}
	}
	if err := network.FillDefaults(&cfg.Spec, nil); err != nil {
		return errors.Wrap(err, "fill configuration defaults")
	}
	if err := network.Validate(&cfg.Spec, cfg.OpenShiftNetwork); err != nil {
		return errors.Wrap(err, "validate configuration")
	}

	cluster := &network.ClusterInfo{
		OpenShift4:          cfg.Platform == "openshift",
		SCCAvailable:        cfg.Cluster.SCCAvailable,
		MonitoringAvailable: cfg.Cluster.MonitoringAvailable,
		IsSingleReplica:     cfg.Cluster.IsSingleReplica,
	}
	components, err := network.RenderByComponent(&cfg.Spec, manifestDir, cfg.OpenShiftNetwork, cluster)
	if err != nil {
		return errors.Wrap(err, "render components")
	}
	stripCNAOGeneratedFields(components)
	sharedObjects := components["common"]
	componentNames := []string{"multus", "multusDynamicNetworks", "linuxBridge", "kubeMacPool", "ovs", "macvtap", "kubevirtIpamController"}
	selected := 0
	for _, name := range componentNames {
		if len(components[name]) > 0 {
			selected++
		}
	}
	if selected == 0 {
		return errors.New("at least one component must be enabled in spec")
	}
	for _, name := range componentNames {
		if len(components[name]) == 0 {
			continue
		}
		objects := componentOutputObjects(sharedObjects, components[name])
		if len(objects) == 0 {
			continue
		}
		componentDir := filepath.Join(outputDir, name)
		if err := os.MkdirAll(componentDir, 0755); err != nil {
			return errors.Wrapf(err, "create output directory for %s", name)
		}
		manifestBytes, err := marshalObjects(objects)
		if err != nil {
			return errors.Wrapf(err, "marshal %s manifests", name)
		}
		if err := os.WriteFile(filepath.Join(componentDir, "manifests.yaml"), manifestBytes, 0644); err != nil {
			return errors.Wrapf(err, "write %s manifests", name)
		}
		if !hco {
			continue
		}
		goBytes, err := generateGo(name, objects)
		if err != nil {
			return errors.Wrapf(err, "generate HCO Go for %s", name)
		}
		if err := os.WriteFile(filepath.Join(componentDir, "hco.go"), goBytes, 0644); err != nil {
			return errors.Wrapf(err, "write %s HCO Go output", name)
		}
	}
	return nil
}

func componentOutputObjects(shared, component []*unstructured.Unstructured) []*unstructured.Unstructured {
	objects := make([]*unstructured.Unstructured, 0, len(shared)+len(component))
	objects = append(objects, component...)
	objects = append(objects, shared...)
	return objects
}

func stripCNAOGeneratedFields(components map[string][]*unstructured.Unstructured) {
	for _, objects := range components {
		for _, object := range objects {
			if object == nil {
				continue
			}
			unstructured.RemoveNestedField(object.Object, "metadata", "creationTimestamp")
			if object.GetKind() == "Namespace" {
				labels := object.GetLabels()
				for _, key := range k8sutil.RemovedLabels() {
					delete(labels, key)
				}
				object.SetLabels(labels)
			}
		}
	}
}

func marshalObjects(objects []*unstructured.Unstructured) ([]byte, error) {
	var out bytes.Buffer
	for _, object := range objects {
		if object == nil || len(object.Object) == 0 {
			continue
		}
		encoded, err := yaml.Marshal(object.Object)
		if err != nil {
			return nil, errors.Wrapf(err, "encode %s %s", object.GetKind(), object.GetName())
		}
		if out.Len() > 0 {
			out.WriteString("---\n")
		}
		out.Write(encoded)
	}
	return out.Bytes(), nil
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
