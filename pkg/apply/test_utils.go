package apply_test

import (
	"bytes"
	"fmt"

	unstructured "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// unstructuredFromYaml creates an unstructured object from a raw yaml string
func unstructuredFromYaml(obj string) *unstructured.Unstructured {
	buf := bytes.NewBufferString(obj)
	decoder := yaml.NewYAMLOrJSONDecoder(buf, 4096)

	u := unstructured.Unstructured{}
	err := decoder.Decode(&u)
	if err != nil {
		panic(fmt.Sprintf("failed to parse test yaml: %v", err))
	}

	return &u
}
