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
	"fmt"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"

	"github.com/operator-framework/java-operator-plugins/pkg/quarkus/v1alpha/scaffolds/internal/templates/util"
)

var _ machinery.Template = &ApplicationPropertiesFile{}

type ApplicationPropertiesFile struct {
	machinery.TemplateMixin
	OrgName     string
	ProjectName string
}

func (f *ApplicationPropertiesFile) SetTemplateDefaults() error {
	if f.ProjectName == "" {
		return fmt.Errorf("invalid Application Properties name")
	}

	if f.Path == "" {
		f.Path = util.PrependResourcePath("application.properties")
	}

	f.TemplateBody = ApplicationPropertiesTemplate

	return nil
}

// TODO: pass in the name of the operator i.e. replace Memcached
const ApplicationPropertiesTemplate = `quarkus.container-image.build=true
#quarkus.container-image.group=
quarkus.container-image.name={{ .ProjectName }}-operator
`
