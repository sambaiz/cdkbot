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
	Diff(repoPath string, stacks []string, contexts map[string]string) (string, bool, error)
	Deploy(repoPath string, stacks []string, contexts map[string]string) (string, error)
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
	if err != nil || cmd.ProcessState.ExitCode() != 0 {
		return fmt.Errorf("npm install failed: %s %v", string(out), err)
	}
	return nil
}

// List stack
func (*Client) List(repoPath string, contexts map[string]string) ([]string, error) {
	args := []string{"run", "cdk", "--", "list"}
	for k, v := range contexts {
		args = append(args, "-c", fmt.Sprintf("%s=%s", k, v))
	}
	cmd := exec.Command("npm", args...)
	cmd.Dir = repoPath
	out, err := cmd.CombinedOutput()
	if err != nil || cmd.ProcessState.ExitCode() != 0 {
		return nil, fmt.Errorf("cdk list failed: %s %v", string(out), err)
	}
	lists := strings.Split(strings.Trim(string(out), "\n"), "\n")[3:]
	return lists, nil
}

// Diff stack and returns (diff, hasDiff, error)
func (*Client) Diff(repoPath string, stacks []string, contexts map[string]string) (string, bool, error) {
	args := []string{"run", "cdk", "--", "diff"}
	for _, stack := range stacks {
		args = append(args, stack)
	}
	for k, v := range contexts {
		args = append(args, "-c", fmt.Sprintf("%s=%s", k, v))
	}
	cmd := exec.Command("npm", args...)
	cmd.Dir = repoPath
	out, err := cmd.CombinedOutput()
	// If the error code is 0, there is no diff, if it is 1, there is diff, otherwise it is an error
	if cmd.ProcessState.ExitCode() != 0 && cmd.ProcessState.ExitCode() != 1 {
		return "failed!", true, fmt.Errorf("cdk diff failed: %s %v", string(out), err)
	}
	lines := []string{}
	for _, line := range strings.Split(strings.Trim(string(out), "\n"), "\n")[3:] {
		if !strings.HasPrefix(line, "npm ERR!") {
			lines = append(lines, line)
		}
	}
	return strings.Trim(strings.Join(lines, "\n"), "\n"), cmd.ProcessState.ExitCode() != 0, nil
}

// Deploy stack
func (*Client) Deploy(repoPath string, stacks []string, contexts map[string]string) (string, error) {
	args := []string{"run", "cdk", "--", "deploy"}
	for _, stack := range stacks {
		args = append(args, stack)
	}
	args = append(args, []string{"--require-approval", "never"}...)
	for k, v := range contexts {
		args = append(args, "-c", fmt.Sprintf("%s=%s", k, v))
	}
	cmd := exec.Command("npm", args...)
	cmd.Dir = repoPath
	out, err := cmd.CombinedOutput()
	if err != nil || cmd.ProcessState.ExitCode() != 0 {
		return "failed!", fmt.Errorf("cdk deploy failed: %s %v", string(out), err)
	}
	lines := []string{}
	for _, line := range strings.Split(strings.Trim(string(out), "\n"), "\n")[3:] {
		if !strings.HasPrefix(line, "npm ERR!") {
			lines = append(lines, line)
		}
	}
	return strings.Trim(strings.Join(lines, "\n"), "\n"), err
}
