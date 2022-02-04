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
	"strings"

	"github.com/operator-framework/java-operator-plugins/pkg/quarkus/v1alpha/scaffolds/internal/templates/util"
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &OperatorFile{}

type OperatorFile struct {
	machinery.TemplateMixin

	// Package is the source files package
	Package string

	// Name of the operator used for the main file.
	OperatorName string
}

func (f *OperatorFile) SetTemplateDefaults() error {
	if f.OperatorName == "" {
		return fmt.Errorf("invalid operator name")
	}

	if strings.HasSuffix(f.OperatorName, "Operator") {
		f.OperatorName = strings.TrimSuffix(f.OperatorName, "Operator")
	}

	if f.Path == "" {
		f.Path = util.PrependJavaPath(f.OperatorName+"Operator.java", util.AsPath(f.Package))
	}

	f.TemplateBody = operatorTemplate

	return nil
}

// TODO: pass in the name of the operator i.e. replace Memcached
const operatorTemplate = `
package {{ .Package }};

import io.javaoperatorsdk.operator.Operator;
import io.quarkus.runtime.Quarkus;
import io.quarkus.runtime.QuarkusApplication;
import io.quarkus.runtime.annotations.QuarkusMain;
import javax.inject.Inject;

@QuarkusMain
public class {{ .OperatorName }}Operator implements QuarkusApplication {

  @Inject Operator operator;

  public static void main(String... args) {
    Quarkus.run({{ .OperatorName }}Operator.class, args);
  }

  @Override
  public int run(String... args) throws Exception {
    operator.start();

    Quarkus.waitForExit();
    return 0;
  }
}
`
