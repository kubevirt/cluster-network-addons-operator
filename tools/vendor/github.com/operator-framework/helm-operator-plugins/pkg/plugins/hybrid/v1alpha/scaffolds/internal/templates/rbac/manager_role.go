// Copyright 2019 The Operator-SDK Authors
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

package rbac

import (
	"fmt"
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &ManagerRole{}

var defaultRoleFile = filepath.Join("config", "rbac", "role.yaml")

// ManagerRole scaffolds the role.yaml file
type ManagerRole struct {
	machinery.TemplateMixin
}

// SetTemplateDefaults implements machinery.Template
func (f *ManagerRole) SetTemplateDefaults() error {
	if f.Path == "" {
		f.Path = defaultRoleFile
	}

	f.TemplateBody = fmt.Sprintf(roleTemplate, machinery.NewMarkerFor(f.Path, rulesMarker))

	return nil
}

const (
	rulesMarker = "rules"
)

const roleTemplate = `apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: manager-role
rules:
##
## Base operator rules
##
# We need to get namespaces so the operator can read namespaces to ensure they exist
- apiGroups:
  - ""
  resources:
  - namespaces
  verbs:
  - get
# We need to manage Helm release secrets
- apiGroups:
  - ""
  resources:
  - secrets
  verbs:
  - "*"
# We need to create events on CRs about things happening during reconciliation
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - create

%s
`
