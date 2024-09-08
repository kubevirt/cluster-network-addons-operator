// Copyright 2021 The Operator-SDK Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/operator-framework/java-operator-plugins/pkg/quarkus/v1beta/scaffolds"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/spf13/pflag"

	"sigs.k8s.io/kubebuilder/v3/pkg/config"
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/model/resource"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugin"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugin/util"
	pluginutil "sigs.k8s.io/kubebuilder/v3/pkg/plugin/util"
)

const filePath = "Makefile"

type createAPIOptions struct {
	CRDVersion string
	Namespaced bool
}

type createAPISubcommand struct {
	config   config.Config
	resource *resource.Resource
	options  createAPIOptions
}

func (opts createAPIOptions) UpdateResource(res *resource.Resource) {

	res.API = &resource.API{
		CRDVersion: opts.CRDVersion,
		Namespaced: opts.Namespaced,
	}

	// Ensure that Path is empty and Controller false as this is not a Go project
	res.Path = ""
	res.Controller = false
}

var (
	_ plugin.CreateAPISubcommand = &createAPISubcommand{}
)

func (p *createAPISubcommand) BindFlags(fs *pflag.FlagSet) {
	fs.SortFlags = false
	fs.StringVar(&p.options.CRDVersion, "crd-version", "v1", "crd version to generate")
	fs.BoolVar(&p.options.Namespaced, "namespaced", true, "resource is namespaced")
}

func (p *createAPISubcommand) InjectConfig(c config.Config) error {
	p.config = c

	return nil
}

func (p *createAPISubcommand) Run(fs machinery.Filesystem) error {
	return nil
}

func (p *createAPISubcommand) Validate() error {
	return nil
}

func (p *createAPISubcommand) PostScaffold() error {
	return nil
}

func (p *createAPISubcommand) Scaffold(fs machinery.Filesystem) error {
	scaffolder := scaffolds.NewCreateAPIScaffolder(p.config, *p.resource)

	var s = fmt.Sprintf(makefileBundleCRDFile, p.resource.Plural, p.resource.QualifiedGroup())
	foundLine := findOldFilesForReplacement(filePath, s)

	if !foundLine {
		makefileBytes, err := afero.ReadFile(fs.FS, filePath)
		if err != nil {
			return err
		}

		projectName := p.config.GetProjectName()
		if projectName == "" {
			dir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("error getting current directory: %w", err)
			}
			projectName = strings.ToLower(filepath.Base(dir))
		}

		makefileBytes = append(makefileBytes, []byte(fmt.Sprintf(makefileBundleVarFragment, p.resource.Plural, p.resource.QualifiedGroup()))...)

		makefileBytes = append([]byte(fmt.Sprintf(makefileBundleImageFragement, p.config.GetDomain(), projectName)), makefileBytes...)

		var mode os.FileMode = 0644
		if info, err := fs.FS.Stat(filePath); err == nil {
			mode = info.Mode()
		}
		if err := afero.WriteFile(fs.FS, filePath, makefileBytes, mode); err != nil {
			return fmt.Errorf("error updating Makefile: %w", err)
		}
	}

	scaffolder.InjectFS(fs)

	if err := scaffolder.Scaffold(); err != nil {
		return err
	}

	return nil
}

func (p *createAPISubcommand) InjectResource(res *resource.Resource) error {
	p.resource = res

	// RESOURCE: &{{cache zeusville.com v1 Joke} jokes  0xc00082a640 false 0xc00082a680}
	p.options.UpdateResource(p.resource)

	if err := p.resource.Validate(); err != nil {
		return err
	}

	// Check that resource doesn't have the API scaffolded
	if res, err := p.config.GetResource(p.resource.GVK); err == nil && res.HasAPI() {
		return errors.New("the API resource already exists")
	}

	// Check that the provided group can be added to the project
	if !p.config.IsMultiGroup() && p.config.ResourcesLength() != 0 && !p.config.HasGroup(p.resource.Group) {
		return fmt.Errorf("multiple groups are not allowed by default, to enable multi-group set 'multigroup: true' in your PROJECT file")
	}

	// Selected CRD version must match existing CRD versions.
	if pluginutil.HasDifferentCRDVersion(p.config, p.resource.API.CRDVersion) {
		return fmt.Errorf("only one CRD version can be used for all resources, cannot add %q", p.resource.API.CRDVersion)
	}

	return nil
}

// findOldFilesForReplacement verifies marker (## marker) and if it found then merge new api CRD file to the odler logic
func findOldFilesForReplacement(path, newfile string) bool {

	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	// remember to close the file at the end of the program
	defer f.Close()

	// read the file line by line using scanner
	scanner := bufio.NewScanner(f)
	var foundMarker bool
	for scanner.Scan() {
		// do something with a line
		if scanner.Text() == "## marker" {
			foundMarker = true
			break
		}
	}

	if foundMarker {
		scanner.Scan()
		catLine := scanner.Text()

		splitByPipe := strings.Split(catLine, "|")

		finalString := strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(splitByPipe[0]), "cat"), "target/kubernetes/kubernetes.yml")

		updatedLine := "	" + "cat" + finalString + newfile + " target/kubernetes/kubernetes.yml" + " |" + splitByPipe[1]

		if err := scanner.Err(); err != nil {
			log.Error(err, "Unable to scan existing bundle target command from the Makefile. New bundle target command being created. This may overwrite any existing commands.")
			return false
		}

		// ReplaceInFile replaces all instances of old with new in the file at path.
		err = util.ReplaceInFile(path, catLine, updatedLine)
		if err != nil {
			log.Error(err, "Unable to replace existing bundle target command from the Makefile. New bundle target command being created. This may overwrite any existing commands.")
			return false
		}
	}

	return foundMarker
}

const (
	makefileBundleCRDFile = `target/kubernetes/%[1]s.%[2]s-v1.yml`
)

const (
	makefileBundleVarFragment = `
##@Bundle
.PHONY: bundle
bundle:  ## Generate bundle manifests and metadata, then validate generated files.
## marker
	cat target/kubernetes/%[1]s.%[2]s-v1.yml target/kubernetes/kubernetes.yml | operator-sdk generate bundle -q --overwrite --version $(VERSION) $(BUNDLE_METADATA_OPTS)
	operator-sdk bundle validate ./bundle
	
.PHONY: bundle-build
bundle-build: ## Build the bundle image.
	docker build -f bundle.Dockerfile -t $(BUNDLE_IMG) .
	
.PHONY: bundle-push
bundle-push: ## Push the bundle image.
	docker push $(BUNDLE_IMG)
`
)

const (
	makefileBundleImageFragement = `
VERSION ?= 0.0.1

# CHANNELS define the bundle channels used in the bundle.
# Add a new line here if you would like to change its default config. (E.g CHANNELS = "candidate,fast,stable")
# To re-generate a bundle for other specific channels without changing the standard setup, you can:
# - use the CHANNELS as arg of the bundle target (e.g make bundle CHANNELS=candidate,fast,stable)
# - use environment variables to overwrite this value (e.g export CHANNELS="candidate,fast,stable")
ifneq ($(origin CHANNELS), undefined)
BUNDLE_CHANNELS := --channels=$(CHANNELS)
endif

# DEFAULT_CHANNEL defines the default channel used in the bundle.
# Add a new line here if you would like to change its default config. (E.g DEFAULT_CHANNEL = "stable")
# To re-generate a bundle for any other default channel without changing the default setup, you can:
# - use the DEFAULT_CHANNEL as arg of the bundle target (e.g make bundle DEFAULT_CHANNEL=stable)
# - use environment variables to overwrite this value (e.g export DEFAULT_CHANNEL="stable")
ifneq ($(origin DEFAULT_CHANNEL), undefined)
BUNDLE_DEFAULT_CHANNEL := --default-channel=$(DEFAULT_CHANNEL)
endif
BUNDLE_METADATA_OPTS ?= $(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)

# IMAGE_TAG_BASE defines the docker.io namespace and part of the image name for remote images.
# This variable is used to construct full image tags for bundle and catalog images.
#
# For example, running 'make bundle-build bundle-push catalog-build catalog-push' will build and push both
# %[1]s/%[2]s-bundle:$VERSION and %[1]s/%[2]s-catalog:$VERSION.
IMAGE_TAG_BASE ?= %[1]s/%[2]s

# BUNDLE_IMG defines the image:tag used for the bundle.
# You can use it as an arg. (E.g make bundle-build BUNDLE_IMG=<some-registry>/<project-name-bundle>:<tag>)
BUNDLE_IMG ?= $(IMAGE_TAG_BASE)-bundle:v$(VERSION)
`
)
