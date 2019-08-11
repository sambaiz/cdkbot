package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReaderRead(t *testing.T) {
	tests := []struct {
		title   string
		in      string
		out     *Config
		isError bool
	}{
		{
			title: "success",
			in:    "./test_config/cdkbot.yml",
			out: &Config{
				CDKRoot: ".",
				Targets: map[string]Target{
					"develop": {
						Contexts: map[string]string{
							"env": "stg",
						},
					},
					"master": {
						Contexts: map[string]string{
							"env": "prd",
						},
					},
				},
			},
		},
		{
			title:   "file is not found",
			in:      "./test_config/notfound.yml",
			isError: true,
		},
		{
			title:   "invalid yaml",
			in:      "./test_config/invalid_yaml.yml",
			isError: true,
		},
	}
	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			config, err := new(Reader).Read(test.in)
			if test.isError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, test.out, config)
		})
	}
}
