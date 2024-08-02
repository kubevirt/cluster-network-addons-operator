/*
Copyright 2019 The Kubernetes Authors.
Modifications copyright 2020 The Operator-SDK Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package scaffolds

import (
	"fmt"
	"os"

	"github.com/spf13/afero"
	"sigs.k8s.io/kubebuilder/v3/pkg/config"
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugin/util"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugins"

	"github.com/operator-framework/helm-operator-plugins/pkg/plugins/helm/v1/chartutil"
	"github.com/operator-framework/helm-operator-plugins/pkg/plugins/hybrid/v1alpha/scaffolds/internal/templates"
	"github.com/operator-framework/helm-operator-plugins/pkg/plugins/hybrid/v1alpha/scaffolds/internal/templates/hack"
	"github.com/operator-framework/helm-operator-plugins/pkg/plugins/hybrid/v1alpha/scaffolds/internal/templates/rbac"
	utils "github.com/operator-framework/helm-operator-plugins/pkg/plugins/util"

	kustomizev2 "sigs.k8s.io/kubebuilder/v3/pkg/plugins/common/kustomize/v2"
	golangv4 "sigs.k8s.io/kubebuilder/v3/pkg/plugins/golang/v4/scaffolds"
)

const (
	imageName = "controller:latest"

	// TODO: This is a placeholder for now. This would probably be the operator-sdk version
	hybridOperatorVersion = "0.2.2"

	// helmPluginVersion is the operator-framework/helm-operator-plugin version to be used in the project
	helmPluginVersion = "v0.2.2"
)

var _ plugins.Scaffolder = &initScaffolder{}

type initScaffolder struct {
	// fs is the filesystem that will be used by the scaffolder
	fs machinery.Filesystem

	config          config.Config
	boilerplatePath string
	license         string
	owner           string
}

// NewInitScaffolder returns a new plugins.Scaffolder for project initialization operations
func NewInitScaffolder(config config.Config, license, owner string) plugins.Scaffolder {
	return &initScaffolder{
		config:          config,
		boilerplatePath: hack.DefaultBoilerplatePath,
		license:         license,
		owner:           owner,
	}
}

// InjectFS implements Scaffolder
func (s *initScaffolder) InjectFS(fs machinery.Filesystem) {
	s.fs = fs
}

// Scaffold implements scaffolder
func (s *initScaffolder) Scaffold() error {
	fmt.Println("Writing scaffolds for you to edit...")

	if err := utils.UpdateKustomizationsInit(); err != nil {
		return fmt.Errorf("error updating kustomization.yaml files: %v", err)
	}

	// Initialize the machinery.Scaffold that will write the files to disk
	scaffold := machinery.NewScaffold(s.fs,
		machinery.WithDirectoryPermissions(0755),
		machinery.WithFilePermissions(0644),
		machinery.WithConfig(s.config),
	)

	// The boilerplate file needs to be scaffolded as a separate step as it is going to be used
	// by rest of the files, even those scaffolded in this command call.
	bpFile := &hack.Boilerplate{
		License: s.license,
		Owner:   s.owner,
	}

	bpFile.Path = s.boilerplatePath
	if err := scaffold.Execute(bpFile); err != nil {
		return err
	}

	boilerplate, err := afero.ReadFile(s.fs.FS, s.boilerplatePath)
	if err != nil {
		return err
	}

	// Initialize the machinery.Scaffold that will write the files to disk
	scaffold = machinery.NewScaffold(s.fs,
		machinery.WithConfig(s.config),
		machinery.WithBoilerplate(string(boilerplate)),
	)

	// create placeholder directories for helm charts and go apis
	err = createDirectories([]string{chartutil.HelmChartsDir, "api", "controllers"})
	if err != nil {
		return err
	}

	err = scaffold.Execute(
		&templates.Main{},
		&templates.GoMod{ControllerRuntimeVersion: golangv4.ControllerRuntimeVersion},
		&templates.GitIgnore{},
		&templates.Watches{},
		&rbac.ManagerRole{},
		&templates.Makefile{
			Image:                    imageName,
			KustomizeVersion:         kustomizev2.KustomizeVersion,
			HybridOperatorVersion:    hybridOperatorVersion,
			ControllerToolsVersion:   golangv4.ControllerToolsVersion,
			ControllerRuntimeVersion: golangv4.ControllerRuntimeVersion,
		},
		&templates.Dockerfile{},
		&templates.DockerIgnore{},
	)

	if err != nil {
		return err
	}

	err = util.RunCmd("Get helm-operator-plugins", "go", "get",
		"github.com/operator-framework/helm-operator-plugins@"+helmPluginVersion)
	if err != nil {
		return err
	}

	return nil
}

func createDirectories(directories []string) error {
	for _, dir := range directories {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("unable to create directory %q : %v", dir, err)
		}
	}
	return nil
}
