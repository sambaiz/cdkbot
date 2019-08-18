package cdk

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Clienter is interface of CDK client
type Clienter interface {
	Setup(repoPath string) error
	List(repoPath string, contexts map[string]string) ([]string, error)
	Diff(repoPath string, stacks string, contexts map[string]string) (string, bool)
	Deploy(repoPath string, stacks string, contexts map[string]string) (string, error)
}

// Client is CDK client
type Client struct{}

// Setup env to run cdk commands
func (*Client) Setup(repoPath string) error {
	if err := os.Setenv("NPM_CONFIG_USERCONFIG", "/opt/nodejs/.npmrc"); err != nil {
		return err
	}
	// Currently, CDK writes cache at $HOME so it needs to change it.
	// https://github.com/aws/aws-cdk/blob/a357bdef775ad30d726090150d496bcb24d576be/packages/aws-cdk/lib/api/util/account-cache.ts#L24
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
func (*Client) List(repoPath string, contexts map[string]string) ([]string, error) {
	args := []string{"run", "cdk", "--", "list"}
	for k, v := range contexts {
		args = append(args, fmt.Sprintf("-c %s=%s", k, v))
	}
	cmd := exec.Command("npm", args...)
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	lists := strings.Split(strings.Trim(string(out), "\n"), "\n")[3:]
	return lists, nil
}

// Diff stack and returns (diff, hasDiff)
func (*Client) Diff(repoPath string, stacks string, contexts map[string]string) (string, bool) {
	args := []string{"run", "cdk", "--", "diff", stacks}
	for k, v := range contexts {
		args = append(args, "-c", fmt.Sprintf("%s=%s", k, v))
	}
	cmd := exec.Command("npm", args...)
	cmd.Dir = repoPath
	out, _ := cmd.CombinedOutput()
	lines := []string{}
	for _, line := range strings.Split(strings.Trim(string(out), "\n"), "\n")[3:] {
		if !strings.HasPrefix(line, "npm ERR!") {
			lines = append(lines, line)
		}
	}
	// If exit code is 0, there are no diffs.
	return strings.Trim(strings.Join(lines, "\n"), "\n"), cmd.ProcessState.ExitCode() != 0
}

// Deploy stack
func (*Client) Deploy(repoPath string, stacks string, contexts map[string]string) (string, error) {
	args := []string{"run", "cdk", "--", "deploy", "--require-approval", "never", stacks, "-c"}
	for k, v := range contexts {
		args = append(args, fmt.Sprintf("%s=%s", k, v))
	}
	cmd := exec.Command("npm", args...)
	cmd.Dir = repoPath
	out, _ := cmd.CombinedOutput()
	result := strings.Join(strings.Split(strings.Trim(string(out), "\n"), "\n")[3:], "\n")
	return result, nil
}
