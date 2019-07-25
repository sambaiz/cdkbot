package cdk

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Setup env to run cdk commands
func Setup(repoPath string) error {
	if err := os.Setenv("NPM_CONFIG_USERCONFIG", "/opt/nodejs/.npmrc"); err != nil {
		return err
	}
	// avoid cdk error https://github.com/aws/aws-cdk/blob/a357bdef775ad30d726090150d496bcb24d576be/packages/aws-cdk/lib/api/util/account-cache.ts#L24
	if err := os.Setenv("HOME", "/tmp"); err != nil {
		return err
	}
	cmd := exec.Command("npm", "install")
	cmd.Dir = repoPath
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to run npm install: %s %s", string(out), err.Error())
	}
	return nil
}

// List stack
func List(repoPath string) ([]string, error) {
	cmd := exec.Command("npm", "run", "cdk", "--", "list")
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return strings.Split(string(out), "\n"), nil
}

// Diff stack
func Diff(repoPath string) (string, bool) {
	cmd := exec.Command("npm", "run", "cdk", "--", "diff")
	cmd.Dir = repoPath
	out, _ := cmd.CombinedOutput()
	return string(out), cmd.ProcessState.ExitCode() != 0
}

// Deploy stack
func Deploy(repoPath string, stacks string) (string, error) {
	cmd := exec.Command("npm", "run", "cdk", "--", "deploy", "--require-approval", "never", stacks)
	cmd.Dir = repoPath
	out, _ := cmd.CombinedOutput()
	return string(out), nil
}
