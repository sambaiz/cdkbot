package git

import (
	"os"
	"os/exec"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
)

// Clienter is interface of git client
type Clienter interface {
	Clone(url, path string, hash *string) error
}

// Client is git client
type Client struct{}

// Clone a git repository
func (*Client) Clone(url, path string, hash *string) error {
	if err := exec.Command("rm", "-rf", path).Run(); err != nil {
		return err
	}
	if err := exec.Command("mkdir", path).Run(); err != nil {
		return err
	}
	repo, err := git.PlainClone(path, false, &git.CloneOptions{
		URL: url,
		Auth: &http.BasicAuth{
			Username: os.Getenv("GITHUB_USER_NAME"),
			Password: os.Getenv("GITHUB_ACCESS_TOKEN"),
		},
	})
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
