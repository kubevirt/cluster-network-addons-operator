// Copyright 2020 The Operator-SDK Authors
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

package ansible

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/pflag"
	"sigs.k8s.io/kubebuilder/v4/pkg/config"
	"sigs.k8s.io/kubebuilder/v4/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v4/pkg/plugin"
	"sigs.k8s.io/kubebuilder/v4/pkg/plugin/util"

	"github.com/operator-framework/ansible-operator-plugins/pkg/plugins/ansible/v1/scaffolds"
	sdkpluginutil "github.com/operator-framework/ansible-operator-plugins/pkg/plugins/util"
)

const (
	groupFlag   = "group"
	versionFlag = "version"
	kindFlag    = "kind"
)

var _ plugin.InitSubcommand = &initSubcommand{}

type initSubcommand struct {
	// Wrapped plugin that we will call at post-scaffold
	apiSubcommand createAPISubcommand

	config config.Config

	// For help text.
	commandName string

	// Flags
	group   string
	version string
	kind    string
}

// UpdateContext injects documentation for the command
func (p *initSubcommand) UpdateMetadata(cliMeta plugin.CLIMetadata, subcmdMeta *plugin.SubcommandMetadata) {
	subcmdMeta.Description = `
Initialize a new Ansible-based operator project.

Writes the following files
- a kubebuilder PROJECT file with the domain and project layout configuration
- a Makefile that provides an interface for building and managing the operator
- Kubernetes manifests and kustomize configuration
- a watches.yaml file that defines the mapping between APIs and Roles/Playbooks

Optionally creates a new API, using the same flags as "create api"
`
	subcmdMeta.Examples = fmt.Sprintf(`
  # Scaffold a project with no API
  $ %[1]s init --plugins=%[2]s --domain=my.domain \

  # Invokes "create api"
  $ %[1]s init --plugins=%[2]s \
      --domain=my.domain \
      --group=apps --version=v1alpha1 --kind=AppService

  $ %[1]s init --plugins=%[2]s \
      --domain=my.domain \
      --group=apps --version=v1alpha1 --kind=AppService \
      --generate-role

  $ %[1]s init --plugins=%[2]s \
      --domain=my.domain \
      --group=apps --version=v1alpha1 --kind=AppService \
      --generate-playbook

  $ %[1]s init --plugins=%[2]s \
      --domain=my.domain \
      --group=apps --version=v1alpha1 --kind=AppService \
      --generate-playbook \
      --generate-role
`, cliMeta.CommandName, pluginKey)

	p.commandName = cliMeta.CommandName
}

func (p *initSubcommand) BindFlags(fs *pflag.FlagSet) {
	fs.SortFlags = false
	fs.StringVar(&p.group, "group", "", "resource Group")
	fs.StringVar(&p.version, "version", "", "resource Version")
	fs.StringVar(&p.kind, "kind", "", "resource Kind")
	p.apiSubcommand.BindFlags(fs)
}

func (p *initSubcommand) InjectConfig(c config.Config) error {
	p.config = c
	return nil
}

func (p *initSubcommand) Scaffold(fs machinery.Filesystem) error {
	if err := addInitCustomizations(p.config.GetProjectName()); err != nil {
		return fmt.Errorf("error updating init manifests: %s", err)
	}

	scaffolder := scaffolds.NewInitScaffolder(p.config)
	scaffolder.InjectFS(fs)
	return scaffolder.Scaffold()
}

func (p *initSubcommand) PostScaffold() error {
	doAPI := p.group != "" || p.version != "" || p.kind != ""
	if !doAPI {
		fmt.Printf("Next: define a resource with:\n$ %s create api\n", p.commandName)
	} else {
		args := []string{"create", "api"}
		// The following three checks should match the default values in sig.k8s.io/kubebuilder/v3/pkg/cli/resource.go
		if p.group != "" {
			args = append(args, fmt.Sprintf("--%s", groupFlag), p.group)
		}
		if p.version != "" {
			args = append(args, fmt.Sprintf("--%s", versionFlag), p.version)
		}
		if p.kind != "" {
			args = append(args, fmt.Sprintf("--%s", kindFlag), p.kind)
		}
		if p.apiSubcommand.options.CRDVersion != defaultCrdVersion {
			args = append(args, fmt.Sprintf("--%s", crdVersionFlag), p.apiSubcommand.options.CRDVersion)
		}
		if p.apiSubcommand.options.DoPlaybook {
			args = append(args, fmt.Sprintf("--%s", generatePlaybookFlag))
		}
		if p.apiSubcommand.options.DoRole {
			args = append(args, fmt.Sprintf("--%s", generateRoleFlag))
		}
		if err := util.RunCmd("Creating the API", os.Args[0], args...); err != nil {
			return err
		}
	}

	return nil
}

// addInitCustomizations will perform the required customizations for this plugin on the common base
func addInitCustomizations(projectName string) error {
	roleFile := filepath.Join("config", "rbac", "role.yaml")

	// We have our own base role file, so we remove the default one
	if err := os.Remove(roleFile); err != nil && !os.IsNotExist(err) {
		return err
	}

	managerFile := filepath.Join("config", "manager", "manager.yaml")
	managerMetricsPatch := filepath.Join("config", "default", "manager_metrics_patch.yaml")

	// todo: we ought to use afero instead. Replace this methods to insert/update
	// by https://github.com/kubernetes-sigs/kubebuilder/pull/2119

	// Add leader election

	err := util.InsertCode(managerFile,
		"--leader-elect",
		fmt.Sprintf("\n          - --leader-election-id=%s", projectName))
	if err != nil {
		return err
	}

	// Enable the proper auth/metrics flags
	err = util.ReplaceInFile(managerMetricsPatch,
		`# This patch adds the args to allow exposing the metrics endpoint using HTTPS
- op: add
  path: /spec/template/spec/containers/0/args/0
  value: --metrics-bind-address=:8443`, `# This patch adds the args to allow exposing the metrics endpoint using HTTPS
- op: add
  path: /spec/template/spec/containers/0/args/0
  value: --metrics-bind-address=:8443
# This patch adds the args to allow securing the metrics endpoint
- op: add
  path: /spec/template/spec/containers/0/args/0
  value: --metrics-secure
# This patch adds the args to allow RBAC-based authn/authz the metrics endpoint
- op: add
  path: /spec/template/spec/containers/0/args/0
  value: --metrics-require-rbac`)
	if err != nil {
		return err
	}

	// update default resource request and limits with bigger values
	const resourcesLimitsFragment = `  resources:
          limits:
            cpu: 500m
            memory: 128Mi
          requests:
            cpu: 10m
            memory: 64Mi
      `

	const resourcesLimitsAnsibleFragment = `  resources:
          limits:
            cpu: 500m
            memory: 768Mi
          requests:
            cpu: 10m
            memory: 256Mi
      `

	err = util.ReplaceInFile(managerFile, resourcesLimitsFragment, resourcesLimitsAnsibleFragment)
	if err != nil {
		return err
	}

	// Add ANSIBLE_GATHERING env var
	const envVar = `
        env:
        - name: ANSIBLE_GATHERING
          value: explicit`
	err = util.InsertCode(managerFile, "name: manager", envVar)
	if err != nil {
		return err
	}

	// replace the default ports because ansible has been using another one
	// todo: remove it when we be able to change the port for the default one
	// issue: https://github.com/operator-framework/ansible-operator-plugins/issues/4331
	err = util.ReplaceInFile(managerFile, "8081", "6789")
	if err != nil {
		return err
	}

	// Remove the call to the command as manager. Helm/Ansible has not been exposing this entrypoint
	// todo: provide the manager entrypoint for helm/ansible and then remove it
	const command = `command:
        - /manager
        `
	err = util.ReplaceInFile(managerFile, command, "")
	if err != nil {
		return err
	}

	if err := sdkpluginutil.UpdateKustomizationsInit(); err != nil {
		return fmt.Errorf("error updating kustomization.yaml files: %v", err)
	}

	return nil
}
