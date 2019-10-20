package git

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClientClone(t *testing.T) {
	hash := "334706a61eb25c944efbf76074e7d48ea9948b9a"
	client := NewClient("https://github.com/sambaiz/cdkbot")
	err := client.Clone("/tmp/cdkbot", &hash)
	assert.Nil(t, err)
	out, err := exec.Command("ls", "/tmp/cdkbot/README.md").Output()
	assert.Nil(t, err)
	assert.Equal(t, "/tmp/cdkbot/README.md\n", string(out))
}
