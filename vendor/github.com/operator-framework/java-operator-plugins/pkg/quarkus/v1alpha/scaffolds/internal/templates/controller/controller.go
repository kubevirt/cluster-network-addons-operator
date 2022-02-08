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

package controller

import (
	"fmt"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"

	"github.com/operator-framework/java-operator-plugins/pkg/quarkus/v1alpha/scaffolds/internal/templates/util"
)

var _ machinery.Template = &Controller{}

type Controller struct {
	machinery.TemplateMixin

	// Package is the source files package
	Package string

	// Name of the operator used for the main file.
	ClassName string
}

func (f *Controller) SetTemplateDefaults() error {
	if f.ClassName == "" {
		return fmt.Errorf("invalid model name")
	}

	if f.Path == "" {
		f.Path = util.PrependJavaPath(f.ClassName+"Controller.java", util.AsPath(f.Package))
	}

	f.TemplateBody = controllerTemplate

	return nil
}

// TODO: pass in the name of the operator i.e. replace Memcached
const controllerTemplate = `package {{ .Package }};

import io.fabric8.kubernetes.client.KubernetesClient;
import io.javaoperatorsdk.operator.api.*;
import io.javaoperatorsdk.operator.api.Context;
import io.javaoperatorsdk.operator.processing.event.EventSourceManager;

@Controller
public class {{ .ClassName }}Controller implements ResourceController<{{ .ClassName }}> {

    private final KubernetesClient client;

    public {{ .ClassName }}Controller(KubernetesClient client) {
        this.client = client;
    }

    // TODO Fill in the rest of the controller

    @Override
    public void init(EventSourceManager eventSourceManager) {
        // TODO: fill in init
    }

    @Override
    public UpdateControl<{{ .ClassName }}> createOrUpdateResource(
        {{ .ClassName }} resource, Context<{{ .ClassName }}> context) {
        // TODO: fill in logic

        return UpdateControl.noUpdate();
    }
}

`
