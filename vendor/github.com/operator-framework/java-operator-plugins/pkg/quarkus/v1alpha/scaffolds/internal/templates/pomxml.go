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

var _ machinery.Template = &PomXmlFile{}

type PomXmlFile struct {
	machinery.TemplateMixin

	// Package is the source files package
	Package         string
	ProjectName     string
	OperatorVersion string
}

func (f *PomXmlFile) SetTemplateDefaults() error {
	if f.Path == "" {
		f.Path = "pom.xml"
	}

	f.TemplateBody = pomxmlTemplate

	return nil
}

// TODO: pass in the name of the operator i.e. replace Memcached
const pomxmlTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
  xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
  xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
  <modelVersion>4.0.0</modelVersion>
  <groupId>{{ .Package }}</groupId>
  <artifactId>{{ .ProjectName }}</artifactId>
  <name>{{ .ProjectName }}</name>
  <version>{{ .OperatorVersion }}-SNAPSHOT</version>
  <packaging>jar</packaging>
  <properties>
    <compiler-plugin.version>3.8.1</compiler-plugin.version>
    <maven.compiler.parameters>true</maven.compiler.parameters>
    <maven.compiler.source>11</maven.compiler.source>
    <maven.compiler.target>11</maven.compiler.target>
    <project.build.sourceEncoding>UTF-8</project.build.sourceEncoding>
    <project.reporting.outputEncoding>UTF-8</project.reporting.outputEncoding>
    <fabric8-client.version>5.4.1</fabric8-client.version>
    <quarkus-sdk.version>1.9.4</quarkus-sdk.version>
    <quarkus.version>1.13.7.Final</quarkus.version>
  </properties>

  <dependencyManagement>
    <dependencies>
      <dependency>
        <groupId>io.fabric8</groupId>
        <artifactId>kubernetes-client-bom</artifactId>
        <version>${fabric8-client.version}</version>
        <type>pom</type>
        <scope>import</scope>
      </dependency>
      <dependency>
        <groupId>io.quarkus</groupId>
        <artifactId>quarkus-bom</artifactId>
        <version>${quarkus.version}</version>
        <type>pom</type>
        <scope>import</scope>
      </dependency>
    </dependencies>
  </dependencyManagement>
  <dependencies>
    <dependency>
      <groupId>io.quarkiverse.operatorsdk</groupId>
      <artifactId>quarkus-operator-sdk</artifactId>
      <version>${quarkus-sdk.version}</version>
    </dependency>
  </dependencies>

  <build>
    <plugins>
      <plugin>
        <groupId>io.quarkus</groupId>
        <artifactId>quarkus-maven-plugin</artifactId>
        <version>${quarkus.version}</version>
        <executions>
          <execution>
            <goals>
              <goal>build</goal>
            </goals>
          </execution>
        </executions>
    </plugin>
    <plugin>
      <artifactId>maven-compiler-plugin</artifactId>
      <version>${compiler-plugin.version}</version>
    </plugin>
    </plugins>
  </build>

  <profiles>
    <profile>
      <id>native</id>
      <properties>
        <quarkus.package.type>native</quarkus.package.type>
      </properties>
    </profile>
  </profiles>

</project>
`
