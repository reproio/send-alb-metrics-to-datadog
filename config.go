package main

import (
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	RequestCountMetricName         string                 `yaml:"request_count_metrics_name"`
	TargetProcessingTimeMetricName string                 `yaml:"target_processing_time_metrics_name"`
	PathTransformingRules          []PathTransformingRule `yaml:"path_transforming_rules"`
	TargetPaths                    []string               `yaml:"target_paths"`
	CustomTags                     []Tag                  `yaml:"custom_tags"`
}

type Tag struct {
	Name   string `yaml:"name"`
	EnvKey string `yaml:"env_key"`
}

func (t *Tag) Key() string {
	return os.Getenv(t.EnvKey)
}

func NewConfigFromFile(path string) (*Config, error) {
	var config Config
	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(buf, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
