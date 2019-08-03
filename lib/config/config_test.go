package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRead(t *testing.T) {
	tests := []struct {
		title    string
		path     string
		expected *Config
		isError  bool
	}{
		{
			title: "ssuccess",
			path:  "./fixture/cdkbot.yml",
			expected: &Config{
				CDKRoot: ".",
				Targets: []Target{
					{
						Branch: "develop",
						Contexts: map[string]string{
							"env": "stg",
						},
					},
					{
						Branch: "master",
						Contexts: map[string]string{
							"env": "prd",
						},
					},
				},
			},
		},
		{
			title:   "file is not found",
			path:    "./fixture/notfound.yml",
			isError: true,
		},
		{
			title:   "invalid yaml",
			path:    "./fixture/invalid_yaml.yml",
			isError: true,
		},
	}
	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			config, err := Read(test.path)
			if test.isError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, test.expected, config)
		})
	}
}
