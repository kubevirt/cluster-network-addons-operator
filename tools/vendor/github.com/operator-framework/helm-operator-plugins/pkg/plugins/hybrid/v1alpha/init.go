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

package v1alpha

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/pflag"
	"sigs.k8s.io/kubebuilder/v3/pkg/config"
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugin"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugin/util"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugins/golang"

	"github.com/operator-framework/helm-operator-plugins/pkg/plugins/hybrid/v1alpha/scaffolds"
	projutil "github.com/operator-framework/helm-operator-plugins/pkg/plugins/util"
)

type initSubcommand struct {
	config config.Config

	// For help text
	commandName string

	// boilerplate options
	license string
	owner   string

	// go config options
	repo string
}

var _ plugin.InitSubcommand = &initSubcommand{}

// UpdateMetadata defines plugin context
func (p *initSubcommand) UpdateMetadata(cliMeta plugin.CLIMetadata, subcmdMeta *plugin.SubcommandMetadata) {
	subcmdMeta.Description = `Initialize a new project including the following files:
	- a "go.mod" with project dependencies
	- a "PROJECT" file that stores project configuration
	- a "Makefile" with several useful make targets for the project
	- several YAML files for project deployment under the "config" directory
	- a "main.go" file that creates the manager that will run the project controllers
  `
	subcmdMeta.Examples = fmt.Sprintf(`  # Initialize a new project with your domain and name in copyright
	$ %[1]s init --plugins=%[2]s --owner "Your Name"

	# Initialize a new project defining a specific project version
	%[1]s init --plugins=%[2]s --project-version 3
`, cliMeta.CommandName, pluginKey)

	p.commandName = cliMeta.CommandName
}

func (p *initSubcommand) BindFlags(fs *pflag.FlagSet) {
	fs.SortFlags = false

	// project args
	fs.StringVar(&p.repo, "repo", "", "name to use for go module (e.g., github.com/user/repo), "+
		"defaults to the go package of the current working directory.")

	// boilerplate args
	fs.StringVar(&p.license, "license", "apache2",
		"license to use to boilerplate, may be one of 'apache2', 'none'")
	fs.StringVar(&p.owner, "owner", "", "owner to add to the copyright")
}

func (p *initSubcommand) InjectConfig(c config.Config) error {
	p.config = c

	// Try to guess repository if flag is not set
	if p.repo == "" {
		repoPath, err := golang.FindCurrentRepo()
		if err != nil {
			return fmt.Errorf("error finding current repository: %v", err)
		}
		p.repo = repoPath
	}

	if err := p.config.SetRepository(p.repo); err != nil {
		return err
	}
	return nil
}

// TODO:
// 1. Pre-scaffold check to verify if the right Go version and directory is used.
// This needs to be added from Kubebuilder.
// 2. Verify if additional customizations are needed in config files as done in
// helm operator.
func (p *initSubcommand) Scaffold(fs machinery.Filesystem) error {
	if err := addInitCustomizations(p.config.GetProjectName()); err != nil {
		return fmt.Errorf("error updating init manifests: %q", err)
	}

	scaffolder := scaffolds.NewInitScaffolder(p.config, p.license, p.owner)
	scaffolder.InjectFS(fs)
	return scaffolder.Scaffold()
}

func (p *initSubcommand) PostScaffold() error {
	err := util.RunCmd("Update dependencies", "go", "mod", "tidy")
	if err != nil {
		return err
	}

	return nil
}

// addInitCustomizations will perform the required customizations for this plugin on the common base
func addInitCustomizations(projectName string) error {
	managerFile := filepath.Join("config", "manager", "manager.yaml")

	// todo: we ought to use afero instead. Replace this methods to insert/update
	// by https://github.com/kubernetes-sigs/kubebuilder/pull/2119

	// Add leader election arg in config/manager/manager.yaml and in config/default/manager_auth_proxy_patch.yaml
	err := util.InsertCode(managerFile,
		"--leader-elect",
		fmt.Sprintf("\n        - --leader-election-id=%s", projectName))
	if err != nil {
		return err
	}
	err = util.InsertCode(filepath.Join("config", "default", "manager_auth_proxy_patch.yaml"),
		"- \"--leader-elect\"",
		fmt.Sprintf("\n        - \"--leader-election-id=%s\"", projectName))
	if err != nil {
		return err
	}

	// Remove the call to the command as manager. Helm has not been exposing this entrypoint
	// todo: provide the manager entrypoint for helm and then remove it
	const command = `command:
        - /manager
        `
	err = projutil.ReplaceInFile(managerFile, command, "")
	if err != nil {
		return err
	}

	return nil
}
