package metricsparser

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"

	"github.com/kubevirt/cluster-network-addons-operator/pkg/monitoring"

	"sigs.k8s.io/controller-tools/pkg/markers"
)

func MetricsOptsToMetricList(Metrics map[monitoring.MetricsKey]monitoring.MetricsOpts, result MetricList) MetricList {
	for _, opts := range Metrics {
		result = append(result, MetricDescriptionToMetric(opts))
	}
	return result
}

type PrometheusCR struct {
	Spec struct {
		Groups []struct {
			Name  string      `yaml:"name"`
			Rules []yaml.Node `yaml:"rules"`
		} `yaml:"groups"`
	} `yaml:"spec"`
}

type Rule struct {
	Record string `yaml:"record,omitempty"`
}

type Comment struct {
	Summary string
	Type    string
}

func ReadFromPrometheusCR() MetricList {
	var cr PrometheusCR
	err := yaml.Unmarshal(ParseTemplateFile(), &cr)
	if err != nil {
		log.Fatal(err)
	}

	reg := &markers.Registry{}
	reg.Define("+help", markers.DescribesPackage, Comment{})
	reg.Define(" +help", markers.DescribesPackage, Comment{})

	ml := make([]Metric, 0)

	for _, group := range cr.Spec.Groups {
		for _, ruleNode := range group.Rules {
			rule := &Rule{}
			_ = ruleNode.Decode(rule)

			if rule.Record != "" {
				s, t := GetComments(reg, &ruleNode)

				ml = append(ml, Metric{
					Name:        rule.Record,
					Description: s,
					MType:       t,
				})
			}
		}
	}

	return ml
}

func GetComments(reg *markers.Registry, ruleNode *yaml.Node) (string, string) {
	if ruleNode.HeadComment == "" {
		return "", ""
	}

	defn := reg.Lookup(ruleNode.HeadComment, markers.DescribesPackage)
	rawComment, _ := defn.Parse(ruleNode.HeadComment)
	comment, ok := rawComment.(Comment)
	if !ok {
		return "", ""
	}

	return comment.Summary, comment.Type
}

func ParseTemplateFile() []byte {
	t, err := template.ParseFiles("data/monitoring/prom-rule.yaml")
	if err != nil {
		log.Fatal(err)
	}

	var doc bytes.Buffer
	test := struct {
		Namespace          string
		RunbookURLTemplate string
	}{
		Namespace:          "testNamespace",
		RunbookURLTemplate: "testRunbookURLTemplate",
	}

	err = t.Execute(&doc, test)
	if err != nil {
		log.Fatal(err)
	}

	return doc.Bytes()
}

type Metric struct {
	Name        string
	Description string
	MType       string
}

func MetricDescriptionToMetric(rrd monitoring.MetricsOpts) Metric {
	return Metric{
		Name:        rrd.Name,
		Description: rrd.Help,
		MType:       rrd.Type,
	}
}

func (m Metric) WriteOut() {
	fmt.Println("###", m.Name)

	writeNewLine := false

	if m.Description != "" {
		fmt.Print(m.Description + ". ")
		writeNewLine = true
	}

	if m.MType != "" {
		fmt.Print("Type: " + m.MType + ".")
		writeNewLine = true
	}

	if writeNewLine {
		fmt.Println()
	}
}

type MetricList []Metric

// Len implements sort.Interface.Len
func (m MetricList) Len() int {
	return len(m)
}

// Less implements sort.Interface.Less
func (m MetricList) Less(i, j int) bool {
	return m[i].Name < m[j].Name
}

// Swap implements sort.Interface.Swap
func (m MetricList) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

func (m *MetricList) add(line string) {
	split := strings.Split(line, " ")
	name := split[2]
	split[3] = strings.Title(split[3])
	description := strings.Join(split[3:], " ")
	*m = append(*m, Metric{Name: name, Description: description})
}

func (m MetricList) WriteOut() {
	for _, met := range m {
		met.WriteOut()
	}
}
