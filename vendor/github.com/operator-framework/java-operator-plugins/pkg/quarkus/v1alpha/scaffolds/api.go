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

package scaffolds

import (
	"sigs.k8s.io/kubebuilder/v3/pkg/config"
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/model/resource"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugins"

	"github.com/operator-framework/java-operator-plugins/pkg/quarkus/v1alpha/scaffolds/internal/templates/controller"
	"github.com/operator-framework/java-operator-plugins/pkg/quarkus/v1alpha/scaffolds/internal/templates/model"
	"github.com/operator-framework/java-operator-plugins/pkg/quarkus/v1alpha/util"
)

type apiScaffolder struct {
	fs machinery.Filesystem

	config   config.Config
	resource resource.Resource
}

// NewCreateAPIScaffolder returns a new plugins.Scaffolder for project initialization operations
func NewCreateAPIScaffolder(cfg config.Config, res resource.Resource) plugins.Scaffolder {
	return &apiScaffolder{
		config:   cfg,
		resource: res,
	}
}

func (s *apiScaffolder) InjectFS(fs machinery.Filesystem) {
	s.fs = fs
}

func (s *apiScaffolder) Scaffold() error {

	if err := s.config.UpdateResource(s.resource); err != nil {
		return err
	}

	// Initialize the machinery.Scaffold that will write the files to disk
	scaffold := machinery.NewScaffold(s.fs,
		// NOTE: kubebuilder's default permissions are only for root users
		machinery.WithDirectoryPermissions(0755),
		machinery.WithFilePermissions(0644),
		machinery.WithConfig(s.config),
		machinery.WithResource(&s.resource),
	)

	var createAPITemplates []machinery.Builder
	createAPITemplates = append(createAPITemplates,
		&model.Model{
			Package:   util.ReverseDomain(s.config.GetDomain()),
			ClassName: util.ToClassname(s.resource.Kind),
		},
		&model.ModelSpec{
			Package:   util.ReverseDomain(s.config.GetDomain()),
			ClassName: util.ToClassname(s.resource.Kind),
		},
		&model.ModelStatus{
			Package:   util.ReverseDomain(s.config.GetDomain()),
			ClassName: util.ToClassname(s.resource.Kind),
		},
		&controller.Controller{
			Package:   util.ReverseDomain(s.config.GetDomain()),
			ClassName: util.ToClassname(s.resource.Kind),
		},
	)

	return scaffold.Execute(createAPITemplates...)
}
