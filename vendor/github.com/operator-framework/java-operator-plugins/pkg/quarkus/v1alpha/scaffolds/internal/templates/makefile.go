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

package templates

import (
	"errors"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &Makefile{}

// Makefile scaffolds the Makefile
type Makefile struct {
	machinery.TemplateMixin

	// Image is controller manager image name
	Image string

	// Kustomize version to use in the project
	KustomizeVersion string

	// // AnsibleOperatorVersion is the version of the ansible-operator binary downloaded by the Makefile.
	// AnsibleOperatorVersion string
}

// SetTemplateDefaults implements machinery.Template
func (f *Makefile) SetTemplateDefaults() error {
	if f.Path == "" {
		f.Path = "Makefile"
	}

	f.TemplateBody = makefileTemplate

	f.IfExistsAction = machinery.Error

	if f.Image == "" {
		f.Image = "controller:latest"
	}

	if f.KustomizeVersion == "" {
		return errors.New("kustomize version is required in scaffold")
	}

	// if f.AnsibleOperatorVersion == "" {
	//     return errors.New("ansible-operator version is required in scaffold")
	// }

	return nil
}

const makefileTemplate = `
# Image URL to use all building/pushing image targets
IMG ?= {{ .Image }}

all: docker-build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Build

docker-build: ## Build docker image with the manager.
	mvn package -Dquarkus.container-image.build=true -Dquarkus.container-image.image=${IMG}

docker-push: ## Push docker image with the manager.
	mvn package -Dquarkus.container-image.push=true -Dquarkus.container-image.image=${IMG}

##@ Deployment

install: ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	@$(foreach file, $(wildcard target/kubernetes/*-v1.yml), kubectl apply -f $(file);)

uninstall: ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config.
	@$(foreach file, $(wildcard target/kubernetes/*-v1.yml), kubectl delete -f $(file);)

deploy: ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	kubectl apply -f target/kubernetes/kubernetes.yml

undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config.
	kubectl delete -f target/kubernetes/kubernetes.yml
`
