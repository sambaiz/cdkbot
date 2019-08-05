package config

import (
	"io/ioutil"

	"github.com/go-yaml/yaml"
)

// Readerer is interface of config reader
type Readerer interface {
	Read(path string) (*Config, error)
}

// Reader is config reader
type Reader struct{}

// Config is cdkbot config
type Config struct {
	CDKRoot string   `yaml:"cdkRoot"`
	Targets []Target `yaml:"targets"`
}

// Target is cdkbot target
type Target struct {
	Branch   string            `yaml:"branch"`
	Contexts map[string]string `yaml:"contexts"`
}

// Read config
func (*Reader) Read(path string) (*Config, error) {
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
