package metricsparser

import (
	"bytes"
	"log"
	"text/template"

	"github.com/machadovilaca/operator-observability/pkg/operatormetrics"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/yaml.v3"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

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

func ReadFromPrometheusCR() []Metric {
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
	operatormetrics.Metric

	Name        string
	Description string
	MType       string
}

func (m Metric) getCollector() prometheus.Collector { return nil }

func (m Metric) GetOpts() operatormetrics.MetricOpts {
	return operatormetrics.MetricOpts{
		Name: m.Name,
		Help: m.Description,
	}
}

func (m Metric) GetType() operatormetrics.MetricType {
	return operatormetrics.MetricType(m.MType)
}
