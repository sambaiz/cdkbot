package eventhandler

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-github/v26/github"
	"github.com/sambaiz/cdkbot/functions/github/client"
	githubClientMock "github.com/sambaiz/cdkbot/functions/github/client/mock"
	cdkMock "github.com/sambaiz/cdkbot/lib/cdk/mock"
	"github.com/sambaiz/cdkbot/lib/config"
	configMock "github.com/sambaiz/cdkbot/lib/config/mock"
	gitMock "github.com/sambaiz/cdkbot/lib/git/mock"
	"github.com/stretchr/testify/assert"
)

func TestEventHandlerIssueComment(t *testing.T) {
	t.Run("comment diff and has diff", func(t *testing.T) {
		buf, err := ioutil.ReadFile("./test_event/issuecomment_created_deploy.json")
		if err != nil {
			t.Error(err)
			return
		}
		var hook *github.IssueCommentEvent
		if err := json.Unmarshal(buf, &hook); err != nil {
			t.Error(err)
			return
		}
		hook.Comment.Body = &[]string{"/diff"}[0]

		ctx := context.Background()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		githubClient := githubClientMock.NewMockClienter(ctrl)
		gliClient := gitMock.NewMockClienter(ctrl)
		configClient := configMock.NewMockReaderer(ctrl)
		cdkClient := cdkMock.NewMockClienter(ctrl)

		hash := "hash"
		repoOwnerName := hook.GetRepo().GetOwner().GetLogin()
		repoName := hook.GetRepo().GetName()
		issueNumber := hook.GetIssue().GetNumber()
		githubClient.EXPECT().CreateStatusOfLatestCommit(
			ctx, repoOwnerName, repoName, issueNumber, client.StatePending).Return(nil)
		githubClient.EXPECT().GetPullRequestLatestCommitHash(
			ctx, repoOwnerName, repoName, issueNumber).Return(hash, nil)
		gliClient.EXPECT().Clone(
			"https://github.com/test-user/test-repo.git",
			clonePath,
			&hash,
		).Return(nil)
		cfg := &config.Config{
			CDKRoot: ".",
		}
		configClient.EXPECT().Read(fmt.Sprintf("%s/cdkbot.yml", clonePath)).Return(cfg, nil)
		cdkPath := fmt.Sprintf("%s/%s", clonePath, cfg.CDKRoot)
		cdkClient.EXPECT().Setup(cdkPath).Return(nil)

		// doActionDiff()
		result := "result"
		args := ""
		cdkClient.EXPECT().Diff(cdkPath).Return(result, true)
		githubClient.EXPECT().CreateComment(
			ctx,
			repoOwnerName,
			repoName,
			issueNumber,
			fmt.Sprintf("### cdk diff %s\n```%s```", args, result),
		).Return(nil)
		githubClient.EXPECT().CreateStatusOfLatestCommit(
			ctx, repoOwnerName, repoName, issueNumber, client.StateFailure).Return(nil)

		eventHandler := &EventHandler{
			cli:    githubClient,
			git:    gliClient,
			config: configClient,
			cdk:    cdkClient,
		}
		assert.Nil(t, eventHandler.IssueComment(ctx, hook))
	})

	t.Run("comment deploy TestStack and success", func(t *testing.T) {
		buf, err := ioutil.ReadFile("./test_event/issuecomment_created_deploy.json")
		if err != nil {
			t.Error(err)
			return
		}
		var hook *github.IssueCommentEvent
		if err := json.Unmarshal(buf, &hook); err != nil {
			t.Error(err)
			return
		}

		ctx := context.Background()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		githubClient := githubClientMock.NewMockClienter(ctrl)
		gliClient := gitMock.NewMockClienter(ctrl)
		configClient := configMock.NewMockReaderer(ctrl)
		cdkClient := cdkMock.NewMockClienter(ctrl)

		hash := "hash"
		repoOwnerName := hook.GetRepo().GetOwner().GetLogin()
		repoName := hook.GetRepo().GetName()
		issueNumber := hook.GetIssue().GetNumber()
		githubClient.EXPECT().CreateStatusOfLatestCommit(
			ctx, repoOwnerName, repoName, issueNumber, client.StatePending).Return(nil)
		githubClient.EXPECT().GetPullRequestLatestCommitHash(
			ctx, repoOwnerName, repoName, issueNumber).Return(hash, nil)
		gliClient.EXPECT().Clone(
			"https://github.com/test-user/test-repo.git",
			clonePath,
			&hash,
		).Return(nil)
		cfg := &config.Config{
			CDKRoot: ".",
		}
		configClient.EXPECT().Read(fmt.Sprintf("%s/cdkbot.yml", clonePath)).Return(cfg, nil)
		cdkPath := fmt.Sprintf("%s/%s", clonePath, cfg.CDKRoot)
		cdkClient.EXPECT().Setup(cdkPath).Return(nil)

		// doActionDeploy()
		result := "result"
		args := "TestStack"
		cdkClient.EXPECT().Deploy(cdkPath, args).Return(result, nil)
		cdkClient.EXPECT().Diff(cdkPath).Return("", false)
		githubClient.EXPECT().CreateComment(
			ctx,
			repoOwnerName,
			repoName,
			issueNumber,
			fmt.Sprintf("### cdk deploy\n```%s```\n%s", result, "All stacks have been deployed :tada:"),
		).Return(nil)
		githubClient.EXPECT().CreateStatusOfLatestCommit(
			ctx, repoOwnerName, repoName, issueNumber, client.StateSuccess).Return(nil)

		eventHandler := &EventHandler{
			cli:    githubClient,
			git:    gliClient,
			config: configClient,
			cdk:    cdkClient,
		}
		assert.Nil(t, eventHandler.IssueComment(ctx, hook))
	})
}

func TestParseCommand(t *testing.T) {
	tests := []struct {
		title   string
		in      string
		out     *command
		isError bool
	}{
		{
			title: "diff",
			in:    "/diff aaa bbb",
			out: &command{
				action: actionDiff,
				args:   "aaa bbb",
			},
		},
		{
			title: "deploy",
			in:    "/deploy aaa bbb",
			out: &command{
				action: actionDeploy,
				args:   "aaa bbb",
			},
		},
		{
			title: "unknown",
			in:    "/unknown aaa bbb",
			out:   nil,
		},
	}
	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			cmd := parseCommand(test.in)
			assert.Equal(t, test.out, cmd)
		})
	}
}
