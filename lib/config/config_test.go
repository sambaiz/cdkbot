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
			in:    "./test/cdkbot.yml",
			out: &Config{
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
			in:      "./test/notfound.yml",
			isError: true,
		},
		{
			title:   "invalid yaml",
			in:      "./test/invalid_yaml.yml",
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
