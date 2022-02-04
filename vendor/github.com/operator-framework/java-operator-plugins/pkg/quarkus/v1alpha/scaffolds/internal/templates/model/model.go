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

package model

import (
	"fmt"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"

	"github.com/operator-framework/java-operator-plugins/pkg/quarkus/v1alpha/scaffolds/internal/templates/util"
)

var _ machinery.Template = &Model{}

type Model struct {
	machinery.TemplateMixin
	machinery.ResourceMixin

	// Package is the source files package
	Package string

	// Name of the operator used for the main file.
	ClassName string
}

func (f *Model) SetTemplateDefaults() error {
	if f.ClassName == "" {
		return fmt.Errorf("invalid model name")
	}

	if f.Path == "" {
		f.Path = util.PrependJavaPath(f.ClassName+".java", util.AsPath(f.Package))
	}

	f.TemplateBody = modelTemplate

	return nil
}

// TODO: pass in the name of the operator i.e. replace Memcached
const modelTemplate = `package {{ .Package }};

{{if .Resource.API.Namespaced}}import io.fabric8.kubernetes.api.model.Namespaced;{{end}}
import io.fabric8.kubernetes.client.CustomResource;
import io.fabric8.kubernetes.model.annotation.Group;
import io.fabric8.kubernetes.model.annotation.Kind;
import io.fabric8.kubernetes.model.annotation.Plural;
import io.fabric8.kubernetes.model.annotation.Version;

@Version("{{ .Resource.Version }}")
@Group("{{ .Resource.QualifiedGroup }}")
@Kind("{{ .Resource.Kind }}")
@Plural("{{ .Resource.Plural }}")
public class {{ .ClassName }} extends CustomResource<{{ .ClassName }}Spec, {{ .ClassName }}Status> {{if .Resource.API.Namespaced}}implements Namespaced {{end}}{}

`
