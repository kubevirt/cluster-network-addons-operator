package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type component struct {
	Url      string `yaml:"url"`
	Commit   string `yaml:"commit"`
	Metadata string `yaml:"metadata"`
}

type componentsConfig struct {
	Components map[string]component `yaml:"components"`
}

func parseComponentsYaml(componentsConfigPath string) (componentsConfig, error) {
	config := componentsConfig{}

	logger.Printf("Opening components config %s", componentsConfigPath)
	componentsData, err := ioutil.ReadFile(componentsConfigPath)
	if err != nil {
		return componentsConfig{}, errors.Wrapf(err, "Failed to open file %s", componentsConfigPath)
	}

	logger.Printf("Unmarshaling components config %s", componentsConfigPath)
	err = yaml.Unmarshal(componentsData, &config)
	if err != nil {
		return componentsConfig{}, errors.Wrapf(err, "Failed to Unmarshal %s", componentsConfigPath)
	}

	if len(config.Components) == 0 {
		return componentsConfig{}, fmt.Errorf("Failed to Unmarshal %s. Output is empty", componentsConfigPath)
	}

	return config, nil
}

func updateComponentsYaml(componentsConfigPath string, config componentsConfig) error {
	logger.Printf("Marshaling components config")
	yamlConfig, err := yaml.Marshal(&config)
	if err != nil {
		return errors.Wrap(err, "Failed to marshal updated components config")
	}

	err = ioutil.WriteFile(componentsConfigPath, yamlConfig, 0644)
	if err != nil {
		return errors.Wrapf(err, "Failed to write to file %s", componentsConfigPath)
	}

	return nil
}

func printCurrentComponentParams(component component) error {
	componentPrettyJson, err := json.MarshalIndent(component, "", "\t")
	if err != nil {
		return errors.Wrap(err, "Failed to print component params")
	}

	logger.Printf("Current component config:\n%s", string(componentPrettyJson))
	return nil
}
