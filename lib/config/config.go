package config

import (
	"io/ioutil"

	"github.com/go-yaml/yaml"
)

type Config struct {
	CDKRoot string   `yaml:"cdkRoot"`
	Targets []Target `yaml:"targets"`
}

type Target struct {
	Branch   string            `yaml:"branch"`
	Contexts map[string]string `yaml:"contexts"`
}

func Read(path string) (*Config, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config Config
	if err := yaml.Unmarshal(buf, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
