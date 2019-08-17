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

// Config is cdkbot config. Targets keys are branch name.
type Config struct {
	CDKRoot     string            `yaml:"cdkRoot"`
	Targets     map[string]Target `yaml:"targets"`
	DeployUsers []string          `yaml:"deploy_users"`
}

// Target is cdkbot target
type Target struct {
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

// IsUserAllowedDeploy returns whether user is allowed to deploy
func (c *Config) IsUserAllowedDeploy(userName string) bool {
	if len(c.DeployUsers) == 0 {
		return true
	}
	for _, user := range c.DeployUsers {
		if user == userName {
			return true
		}
	}
	return false
}
