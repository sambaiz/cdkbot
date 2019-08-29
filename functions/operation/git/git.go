package git

import (
	"fmt"
	"os/exec"
)

// Clienter is interface of git client
type Clienter interface {
	Clone(path string, hash *string) error
	Merge(path, branch string) error
	Checkout(path, fileName, branch string) error
}

// Client is git client
type Client struct {
	cloneURL string
}

// NewClient creates git client
func NewClient(cloneURL string) *Client {
	return &Client{
		cloneURL: cloneURL,
	}
}

// Clone a git repository
func (c *Client) Clone(path string, hash *string) error {
	if err := exec.Command("rm", "-rf", path).Run(); err != nil {
		return err
	}
	if err := exec.Command("mkdir", "-p", path).Run(); err != nil {
		return err
	}
	if err := exec.Command("git", "clone", c.cloneURL, path).Run(); err != nil {
		return fmt.Errorf("git clone failed: %s", err.Error())
	}
	if hash != nil {
		cmd := exec.Command("git", "checkout", *hash)
		cmd.Dir = path
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("git checkout failed: %s", err.Error())
		}
	}
	return nil
}

// Merge a branch
func (c *Client) Merge(path, branch string) error {
	cmd := exec.Command("git", "merge", branch, "-m", `"cdkbot merged"`)
	cmd.Dir = path
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git merge failed: %s", err.Error())
	}
	return nil
}

// Checkout file of branch
func (c *Client) Checkout(path, fileName, branch string) error {
	cmd := exec.Command("git", "checkout", branch, fileName)
	cmd.Dir = path
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git checkout %s of %s failed: %s", fileName, branch, err.Error())
	}
	return nil
}
