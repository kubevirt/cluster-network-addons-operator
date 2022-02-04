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

var _ machinery.Template = &ModelSpec{}

type ModelSpec struct {
	machinery.TemplateMixin

	// Package is the source files package
	Package string

	// Name of the operator used for the main file.
	ClassName string
}

func (f *ModelSpec) SetTemplateDefaults() error {
	if f.ClassName == "" {
		return fmt.Errorf("invalid operator name")
	}

	if f.Path == "" {
		f.Path = util.PrependJavaPath(f.ClassName+"Spec.java", util.AsPath(f.Package))
	}

	f.TemplateBody = modelSpecTemplate

	return nil
}

// TODO: pass in the name of the operator i.e. replace Memcached
const modelSpecTemplate = `package {{ .Package }};

public class {{ .ClassName }}Spec {

    // Add Spec information here
}
`
