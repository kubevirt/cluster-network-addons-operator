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
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &Dockerfile{}

// Dockerfile scaffolds a file that defines the containerized build process
type Dockerfile struct {
	machinery.TemplateMixin
}

// SetTemplateDefaults implements file.Template
func (f *Dockerfile) SetTemplateDefaults() error {
	if f.Path == "" {
		f.Path = "Dockerfile"
	}

	f.TemplateBody = dockerfileTemplate

	return nil
}

// The current template scaffolds copying of go dependencies and building
// main.go. If there are any other depencies or folders to be copied like
// `api/` and `controller/` they would have to be added.

const dockerfileTemplate = `# Build the manager binary
FROM golang:1.20 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY cmd/ cmd/
COPY api/ api/
COPY controllers/ controllers/

# Build
RUN GOOS=linux GOARCH=amd64 go build -a -o manager cmd/main.go

FROM registry.access.redhat.com/ubi8/ubi-micro:8.7

ENV HOME=/opt/helm \
    USER_NAME=helm \
    USER_UID=1001

RUN echo "${USER_NAME}:x:${USER_UID}:0:${USER_NAME} user:${HOME}:/sbin/nologin" >> /etc/passwd

# Copy necessary files with the right permissions
COPY --chown=${USER_UID}:0 watches.yaml ${HOME}/watches.yaml
COPY --chown=${USER_UID}:0 helm-charts  ${HOME}/helm-charts

# Copy manager binary
COPY --from=builder /workspace/manager .

USER ${USER_UID}

WORKDIR ${HOME}

ENTRYPOINT ["/manager"]
`
