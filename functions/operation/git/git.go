package git

import (
	"os/exec"

	goGit "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

// Clienter is interface of git client
type Clienter interface {
	Clone(path string, hash *string) (*goGit.Worktree, error)
	Merge(workTree *goGit.Worktree, branch string) error
}

// Client is git client
type Client struct {
	cloneOptions *goGit.CloneOptions
}

// NewClient creates git client
func NewClient(cloneOptions *goGit.CloneOptions) *Client {
	return &Client{
		cloneOptions: cloneOptions,
	}
}

// Clone a git repository
func (c *Client) Clone(path string, hash *string) (*goGit.Worktree, error) {
	if err := exec.Command("rm", "-rf", path).Run(); err != nil {
		return nil, err
	}
	if err := exec.Command("mkdir", path).Run(); err != nil {
		return nil, err
	}
	repo, err := goGit.PlainClone(path, false, c.cloneOptions)
	if err != nil {
		return nil, err
	}
	workTree, err := repo.Worktree()
	if err != nil {
		return nil, err
	}
	if hash != nil {
		workTree, err := repo.Worktree()
		if err != nil {
			return nil, err
		}
		if err := workTree.Checkout(&goGit.CheckoutOptions{
			Hash: plumbing.NewHash(*hash),
		}); err != nil {
			return nil, err
		}
	}
	return workTree, nil
}

// Merge a branch. Noted it fails if not fast-forward merge
func (c *Client) Merge(workTree *goGit.Worktree, branch string) error {
	// go-git doesn't support Merge() so use Pull() instead. It fails if not fast-forward merge.
	if err := workTree.Pull(&goGit.PullOptions{
		ReferenceName: plumbing.NewBranchReferenceName(branch),
	}); err != nil {
		return err
	}
	return nil
}

