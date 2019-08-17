package git

import (
	"os/exec"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

// Clienter is interface of git client
type Clienter interface {
	Clone(path string, hash *string) error
}

// Client is git client
type Client struct{
	cloneOptions *git.CloneOptions
}

// NewClient creates git client
func NewClient(cloneOptions *git.CloneOptions) *Client {
	return &Client{
		cloneOptions: cloneOptions,
	}
}


// Clone a git repository
func (c *Client) Clone(path string, hash *string) error {
	if err := exec.Command("rm", "-rf", path).Run(); err != nil {
		return err
	}
	if err := exec.Command("mkdir", path).Run(); err != nil {
		return err
	}
	repo, err := git.PlainClone(path, false, c.cloneOptions)
	if err != nil {
		return err
	}
	if hash != nil {
		workTree, err := repo.Worktree()
		if err != nil {
			return err
		}
		if err := workTree.Checkout(&git.CheckoutOptions{
			Hash: plumbing.NewHash(*hash),
		}); err != nil {
			return err
		}
	}
	return nil
}
