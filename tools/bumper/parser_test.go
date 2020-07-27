package main

import (
	"fmt"
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/getlantern/deepcopy"
)

var _ = Describe("Testing internal parser", func() {
	var tmpfile *os.File

	createTempFileWithContent := func(content []byte) {
		var err error
		tmpfile, err = ioutil.TempFile("", "test")
		Expect(err).ToNot(HaveOccurred(), "Should successfully create temp file")

		defer tmpfile.Close()

		By("Writing and closing the temporary file")
		_, err = tmpfile.Write(content)
		Expect(err).ToNot(HaveOccurred(), "Should successfully write content to the temp file")
	}

	AfterEach(func() {
		By("Removing the temporary file")
		err := os.Remove(tmpfile.Name())
		Expect(err).ToNot(HaveOccurred(), "Should successfully remove the temp file")
	})

	expectedValidComponentsConfig := componentsConfig{
		Components: map[string]component{
			"multus": {
				Url:      "https://github.com/intel/multus-cni",
				Commit:   "bfaf22964b51b93b08dcb1f10cbd8f8cb9a3603a",
				Metadata: "v3.4.1"},
			"linux-bridge": {
				Url:      "https://github.com/containernetworking/plugins",
				Commit:   "ad10b6fa91aacd720f1f9ab94341a97a82a24965",
				Metadata: "v0.8.6"},
			"bridge-marker": {
				Url:      "https://github.com/kubevirt/bridge-marker",
				Commit:   "049eb6a796b742767a0be838f6ddb7f79c97dd60",
				Metadata: "0.3.0"},
			"kubemacpool": {
				Url:      "https://github.com/k8snetworkplumbingwg/kubemacpool",
				Commit:   "7728339b6c8dcfc9d53cc18ce8e06a910094ed47",
				Metadata: "v0.16.0"},
			"nmstate": {
				Url:      "https://github.com/nmstate/kubernetes-nmstate",
				Commit:   "6f41747788235c3363a314dfae46a4ed90e1846a",
				Metadata: "v0.25.0"},
			"ovs-cni": {
				Url:      "https://github.com/kubevirt/ovs-cni",
				Commit:   "edb4754d08f49be54c399c938895fe82aab6aa5a",
				Metadata: "v0.12.0"},
			"macvtap-cni": {
				Url:      "https://github.com/kubevirt/macvtap-cni",
				Commit:   "14e5a3c7bf516bd4542366a78b7af02f04230ac5",
				Metadata: "v0.2.0"},
		},
	}

	var componentsDummyYaml = []byte(`
components:
  multus:
    url: "https://github.com/intel/multus-cni"
    commit: "bfaf22964b51b93b08dcb1f10cbd8f8cb9a3603a"
    metadata: "v3.4.1"
  linux-bridge:
    url: "https://github.com/containernetworking/plugins"
    commit: "ad10b6fa91aacd720f1f9ab94341a97a82a24965"
    metadata: "v0.8.6"
  bridge-marker:
    url: "https://github.com/kubevirt/bridge-marker"
    commit: "049eb6a796b742767a0be838f6ddb7f79c97dd60"
    metadata: "0.3.0"
  kubemacpool:
    url: "https://github.com/k8snetworkplumbingwg/kubemacpool"
    commit: "7728339b6c8dcfc9d53cc18ce8e06a910094ed47"
    metadata: "v0.16.0"
  nmstate:
    url: "https://github.com/nmstate/kubernetes-nmstate"
    commit: "6f41747788235c3363a314dfae46a4ed90e1846a"
    metadata: "v0.25.0"
  ovs-cni:
    url: "https://github.com/kubevirt/ovs-cni"
    commit: "edb4754d08f49be54c399c938895fe82aab6aa5a"
    metadata: "v0.12.0"
  macvtap-cni:
    url: "https://github.com/kubevirt/macvtap-cni"
    commit: "14e5a3c7bf516bd4542366a78b7af02f04230ac5"
    metadata: "v0.2.0"
`)

	var corruptedDummyYaml = []byte(`
components:
 Currepted: "data""
    Shouldnt': "work"
`)

	var emptyYaml = []byte(``)

	Describe("ParseComponentsYaml function", func() {
		Context("when yaml file is empty", func() {
			BeforeEach(func() {
				By("Creating a temporary file with empty content")
				createTempFileWithContent(emptyYaml)
			})
			It("should return an error", func() {
				_, err := parseComponentsYaml(tmpfile.Name())
				Expect(err).To(HaveOccurred(), "Should Fail to parse file")
			})
		})
		Context("when yaml file is corrupted", func() {
			BeforeEach(func() {
				By("Creating a temporary file with corrupted content")
				createTempFileWithContent(corruptedDummyYaml)
			})
			It("should return an error because the unmarshaling failed", func() {
				By(fmt.Sprintf("Parsing content of temporary file %s", tmpfile.Name()))
				_, err := parseComponentsYaml(tmpfile.Name())
				Expect(err).To(HaveOccurred(), "Should Fail to parse file")
			})
		})
		Context("when yaml file is valid", func() {
			BeforeEach(func() {
				By("Creating a temporary file with empty content")
				createTempFileWithContent(componentsDummyYaml)
			})
			It("should correctly parse params of all components", func() {
				By(fmt.Sprintf("Parsing content of temporary file %s", tmpfile.Name()))
				componentsConfig, err := parseComponentsYaml(tmpfile.Name())
				Expect(err).ToNot(HaveOccurred(), "Should Succeed to parse file")

				By("Checking map is not empty")
				Expect(componentsConfig.Components).ToNot(BeEmpty(), "Components map Should not be empty")
				By("Checking map output is as expected")
				Expect(componentsConfig).To(Equal(expectedValidComponentsConfig), "componentsConfig should be identical to expected")
			})
		})
	})

	Describe("UpdateComponentsYaml function", func() {
		Context("when yaml file is valid", func() {
			var validComponentsConfig componentsConfig
			var err error

			BeforeEach(func() {
				By("Creating a temporary file with empty content")
				createTempFileWithContent(componentsDummyYaml)

				By("Parsing the yaml")
				validComponentsConfig, err = parseComponentsYaml(tmpfile.Name())
				Expect(err).ToNot(HaveOccurred(), "Should Succeed to parse file")
			})

			It("should correctly update existing component in file", func() {
				By("Changing some params in existing component the yaml")
				updatedComponent := component{Url: "https://github.com/qinqon/kube-admission-webhook", Commit: "f8f7795d05781cdaf869c301e4736adc29a0de28", Metadata: "v0.11.0"}
				validComponentsConfig.Components["macvtap-cni"] = updatedComponent

				// copy expected struct to add the change expected in this test
				expectedValidComponentsConfigCopy := componentsConfig{}
				deepcopy.Copy(&expectedValidComponentsConfigCopy, &expectedValidComponentsConfig)
				expectedValidComponentsConfigCopy.Components["macvtap-cni"] = updatedComponent

				err := updateComponentsYaml(tmpfile.Name(), validComponentsConfig)
				Expect(err).ToNot(HaveOccurred(), "Should Succeed to updating the file")

				By("Checking params updated in the yaml")
				newcomponentsConfig, err := parseComponentsYaml(tmpfile.Name())

				By("Checking map output is as expected")
				Expect(newcomponentsConfig).To(Equal(expectedValidComponentsConfigCopy), "componentsConfig should be identical to expected")
			})

			It("should correctly add new component in file", func() {
				By("add a new component the yaml")
				newComponent := component{Url: "https://github.com/qinqon/kube-admission-webhook", Commit: "f8f7795d05781cdaf869c301e4736adc29a0de28", Metadata: "v0.11.0"}
				validComponentsConfig.Components["kube-admission"] = newComponent

				// copy expected struct to add the change expected in this test
				expectedValidComponentsConfigCopy := componentsConfig{}
				deepcopy.Copy(&expectedValidComponentsConfigCopy, &expectedValidComponentsConfig)
				expectedValidComponentsConfigCopy.Components["kube-admission"] = newComponent

				err := updateComponentsYaml(tmpfile.Name(), validComponentsConfig)
				Expect(err).ToNot(HaveOccurred(), "Should Succeed to updating the file")

				By("Checking params updated in the yaml")
				newcomponentsConfig, err := parseComponentsYaml(tmpfile.Name())
				Expect(err).ToNot(HaveOccurred(), "Should Succeed to parse file after update")

				By("Checking map output is as expected")
				Expect(newcomponentsConfig).To(Equal(expectedValidComponentsConfigCopy), "componentsConfig should be identical to expected")
			})
		})
	})
})
