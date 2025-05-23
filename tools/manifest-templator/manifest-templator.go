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
 * Copyright 2018-2020 Red Hat, Inc.
 *
 */

package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"

	"github.com/ghodss/yaml"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	components "github.com/kubevirt/cluster-network-addons-operator/pkg/components"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

type operatorData struct {
	Deployment        string
	DeploymentSpec    string
	RoleString        string
	Rules             string
	ClusterRoleString string
	ClusterRules      string
	CRD               *extv1.CustomResourceDefinition
	CRDString         string
	CRDVersion        string
	CRString          string
	RelatedImages     components.RelatedImages
}

type templateData struct {
	Version         string
	VersionReplaces string
	OperatorVersion string
	Namespace       string
	ContainerPrefix string
	ImageName       string
	ContainerTag    string
	ImagePullPolicy string
	CNA             *operatorData
	AddonsImages    *components.AddonsImages
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func fixResourceString(in string, indention int) string {
	out := strings.Builder{}
	scanner := bufio.NewScanner(strings.NewReader(in))
	for scanner.Scan() {
		line := scanner.Text()
		// remove separator lines
		if !strings.HasPrefix(line, "---") {
			// indent so that it fits into the manifest
			// spaces is is indention - 2, because we want to have 2 spaces less for being able to start an array
			spaces := strings.Repeat(" ", indention-2)
			if strings.HasPrefix(line, "apiGroups") {
				// spaces + array start
				out.WriteString(spaces + "- " + line + "\n")
			} else {
				// 2 more spaces
				out.WriteString(spaces + "  " + line + "\n")
			}
		}
	}
	return out.String()
}

func marshallObject(obj interface{}, writer io.Writer) error {
	jsonBytes, err := json.Marshal(obj)
	check(err)

	var r unstructured.Unstructured
	if err := json.Unmarshal(jsonBytes, &r.Object); err != nil {
		return err
	}

	// remove status and metadata.creationTimestamp
	unstructured.RemoveNestedField(r.Object, "template", "metadata", "creationTimestamp")
	unstructured.RemoveNestedField(r.Object, "spec", "template", "metadata", "creationTimestamp")
	unstructured.RemoveNestedField(r.Object, "metadata", "creationTimestamp")
	unstructured.RemoveNestedField(r.Object, "status")

	jsonBytes, err = json.Marshal(r.Object)
	if err != nil {
		return err
	}

	yamlBytes, err := yaml.JSONToYAML(jsonBytes)
	if err != nil {
		return err
	}

	// fix templates by removing quotes...
	s := string(yamlBytes)
	s = strings.Replace(s, "'{{", "{{", -1)
	s = strings.Replace(s, "}}'", "}}", -1)
	yamlBytes = []byte(s)

	_, err = writer.Write([]byte("---\n"))
	if err != nil {
		return err
	}

	_, err = writer.Write(yamlBytes)
	if err != nil {
		return err
	}

	return nil
}

func getCNA(data *templateData, allowMultus bool) {
	writer := strings.Builder{}

	// Get CNA Deployment
	cnadeployment := components.GetDeployment(
		data.Version,
		data.OperatorVersion,
		data.Namespace,
		data.ContainerPrefix,
		data.ImageName,
		data.ContainerTag,
		data.ImagePullPolicy,
		data.AddonsImages,
	)
	err := marshallObject(cnadeployment, &writer)
	check(err)
	deployment := writer.String()

	// Get CNA DeploymentSpec for CSV
	writer = strings.Builder{}
	err = marshallObject(cnadeployment.Spec, &writer)
	check(err)
	deploymentSpec := fixResourceString(writer.String(), 12)

	// Get CNA Role
	writer = strings.Builder{}
	role := components.GetRole(data.Namespace)
	marshallObject(role, &writer)
	roleString := writer.String()

	// Get the Rules out of CNA's ClusterRole
	writer = strings.Builder{}
	cnaRules := role.Rules
	for _, rule := range cnaRules {
		err := marshallObject(rule, &writer)
		check(err)
	}
	rules := fixResourceString(writer.String(), 14)

	// Get CNA ClusterRole
	writer = strings.Builder{}
	clusterRole := components.GetClusterRole(allowMultus)
	marshallObject(clusterRole, &writer)
	clusterRoleString := writer.String()

	// Get the Rules out of CNA's ClusterRole
	writer = strings.Builder{}
	cnaClusterRules := clusterRole.Rules
	for _, rule := range cnaClusterRules {
		err := marshallObject(rule, &writer)
		check(err)
	}
	clusterRules := fixResourceString(writer.String(), 14)

	// Get CNA CRD
	writer = strings.Builder{}
	crd := components.GetCrd()
	marshallObject(crd, &writer)
	crdString := writer.String()
	crdString = addPreserveUnknownFields(crdString)
	crdVersion := crd.Spec.Versions[0].Name

	// Get CNA CR
	writer = strings.Builder{}
	cr := components.GetCRV1()
	marshallObject(cr, &writer)
	crString := writer.String()

	// Get related images
	relatedImages := data.AddonsImages.ToRelatedImages()
	selfImageName := fmt.Sprintf("%s/%s:%s", data.ContainerPrefix, data.ImageName, data.ContainerTag)
	relatedImages.Add(selfImageName)

	cnaData := operatorData{
		Deployment:        deployment,
		DeploymentSpec:    deploymentSpec,
		RoleString:        roleString,
		Rules:             rules,
		ClusterRoleString: clusterRoleString,
		ClusterRules:      clusterRules,
		CRD:               crd,
		CRDString:         crdString,
		CRDVersion:        crdVersion,
		CRString:          crString,
		RelatedImages:     relatedImages,
	}
	data.CNA = &cnaData
}

func addPreserveUnknownFields(crdString string) string {
	// TODO replace this solution with a better one once this issue get resolved:
	// https://github.com/kubernetes/kubernetes/issues/95702
	return crdString + "  preserveUnknownFields: false"
}

func main() {
	version := flag.String("version", "", "The csv version")
	versionReplaces := flag.String("version-replaces", "", "The csv version this replaces")
	operatorVersion := flag.String("operator-version", "", "The operator version")
	namespace := flag.String("namespace", components.Namespace, "Namespace used by csv")
	containerPrefix := flag.String("container-prefix", "quay.io/kubevirt", "The container repository used for the operator image")
	imageName := flag.String("image-name", components.Name, "The operator image's name")
	containerTag := flag.String("container-tag", "latest", "The operator image's container tag")
	imagePullPolicy := flag.String("image-pull-policy", "Always", "The pull policy to use on the operator image")
	allowMultus := flag.Bool("allow-multus", true, "Install Multus RBAC, making it possible to deploy Multus through the operator (should be disabled on OpenShift to limit CNAO's RBAC to the necessary minimum)")
	multusImage := flag.String("multus-image", components.MultusImageDefault, "The multus image managed by CNA")
	linuxBridgeCniImage := flag.String("linux-bridge-cni-image", components.LinuxBridgeCniImageDefault, "The linux bridge cni image managed by CNA")
	linuxBridgeMarkerImage := flag.String("linux-bridge-marker-image", components.LinuxBridgeMarkerImageDefault, "The linux bridge marker image managed by CNA")
	kubeMacPoolImage := flag.String("kubemacpool-image", components.KubeMacPoolImageDefault, "The kubemacpool-image managed by CNA")
	ovsCniImage := flag.String("ovs-cni-image", components.OvsCniImageDefault, "The ovs cni image managed by CNA")
	macvtapCniImage := flag.String("macvtap-cni-image", components.MacvtapCniImageDefault, "The macvtap cni image managed by CNA")
	kubeRbacProxyImage := flag.String("kube-rbac-proxy-image", components.KubeRbacProxyImageDefault, "The kube rbac proxy used by CNA")
	coreDNSImage := flag.String("core-dns-image", components.CoreDNSImageDefault, "The coredns image used by CNA")
	multusDynamicNetworksImage := flag.String("multus-dynamic-networks-image", components.MultusDynamicNetworksImageDefault, "The multus dynamic networks controller image managed by CNA")
	kubeSecondaryDNSImage := flag.String("kube-secondary-dns-image", components.KubeSecondaryDNSImageDefault, "The kubesecondarydns-image managed by CNA")
	kubevirtIpamControllerImage := flag.String("kubevirt-ipam-controller-image", components.KubevirtIpamControllerImageDefault, "The kubevirtipamcontroller-image managed by CNA")
	dumpOperatorCRD := flag.Bool("dump-crds", false, "Append operator CRD to bottom of template. Used for csv-generator")
	inputFile := flag.String("input-file", "", "Not used for csv-generator")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.CommandLine.ParseErrorsWhitelist.UnknownFlags = true
	pflag.Parse()

	data := templateData{
		Version:         *version,
		VersionReplaces: *versionReplaces,
		OperatorVersion: *operatorVersion,
		Namespace:       *namespace,
		ContainerPrefix: *containerPrefix,
		ImageName:       *imageName,
		ContainerTag:    *containerTag,
		ImagePullPolicy: *imagePullPolicy,
		AddonsImages: (&components.AddonsImages{
			Multus:                 *multusImage,
			LinuxBridgeCni:         *linuxBridgeCniImage,
			LinuxBridgeMarker:      *linuxBridgeMarkerImage,
			KubeMacPool:            *kubeMacPoolImage,
			OvsCni:                 *ovsCniImage,
			MacvtapCni:             *macvtapCniImage,
			KubeRbacProxy:          *kubeRbacProxyImage,
			MultusDynamicNetworks:  *multusDynamicNetworksImage,
			KubeSecondaryDNS:       *kubeSecondaryDNSImage,
			KubevirtIpamController: *kubevirtIpamControllerImage,
			CoreDNS:                *coreDNSImage,
		}).FillDefaults(),
	}

	// Load in all CNA Resources
	getCNA(&data, *allowMultus)

	if *inputFile == "" {
		panic("Must specify input file")
	}

	manifestTemplate := template.Must(template.ParseFiles(*inputFile))
	err := manifestTemplate.Execute(os.Stdout, data)
	check(err)

	if *dumpOperatorCRD {
		fmt.Printf(data.CNA.CRDString)
	}
}
