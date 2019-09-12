package github

import (
	"context"
	"fmt"
	"github.com/sambaiz/cdkbot/functions/operation/command"
	"github.com/sambaiz/cdkbot/functions/operation/logger"
	"github.com/sambaiz/cdkbot/functions/operation/platform/github/client"
	"net/http"
	"os"
	"strings"

	"go.uber.org/zap"

	"github.com/aws/aws-lambda-go/events"
	goGitHub "github.com/google/go-github/v26/github"
)

// Handler handles GitHub webhook
func Handler(
	ctx context.Context,
	req events.APIGatewayProxyRequest,
	logger logger.Loggerer,
) (*events.APIGatewayProxyResponse, error) {
	if err := goGitHub.ValidateSignature(
		req.Headers["X-Hub-Signature"],
		[]byte(req.Body),
		[]byte(os.Getenv("GITHUB_WEBHOOK_SECRET")),
	); err != nil {
		logger.Info("Signature is invalid", zap.Error(err))
		return &events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
		}, nil
	}
	hook, err := goGitHub.ParseWebHook(req.Headers["X-GitHub-Event"], []byte(req.Body))
	if err != nil {
		logger.Error("parse hook error", zap.Error(err))
		return &events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
		}, nil
	}

	switch ev := hook.(type) {
	case *goGitHub.PullRequestEvent:
		runner := command.NewRunner(
			client.New(
				ctx,
				ev.GetRepo().GetOwner().GetLogin(),
				ev.GetRepo().GetName(),
				ev.GetPullRequest().GetNumber(),
			),
			fmt.Sprintf("https://%s:%s@%s",
				os.Getenv("GITHUB_USER_NAME"),
				os.Getenv("GITHUB_ACCESS_TOKEN"),
				strings.Replace(ev.GetRepo().GetCloneURL(), "https://", "", 1)),
				logger,
		)
		switch ev.GetAction() {
		case "opened":
			err = runner.Diff(ctx)
		}
	case *goGitHub.PushEvent:
		client, err := client.NewWithHeadBranch(
			ctx,
			ev.GetRepo().GetOwner().GetLogin(),
			ev.GetRepo().GetName(),
			strings.TrimLeft(ev.GetRef(), "refs/heads/"),
		)
		if err != nil {
			// Push to branch where PR is not created does nothing
			return &events.APIGatewayProxyResponse{
				StatusCode: http.StatusOK,
			}, nil
		}
		err = command.NewRunner(
			client,
			fmt.Sprintf("https://%s:%s@%s",
				os.Getenv("GITHUB_USER_NAME"),
				os.Getenv("GITHUB_ACCESS_TOKEN"),
				strings.Replace(ev.GetRepo().GetCloneURL(), "https://", "", 1)),
				logger,
		).Diff(ctx)
	case *goGitHub.IssueCommentEvent:
		runner := command.NewRunner(
			client.New(
				ctx,
				ev.GetRepo().GetOwner().GetLogin(),
				ev.GetRepo().GetName(),
				ev.GetIssue().GetNumber(),
			),
			fmt.Sprintf("https://%s:%s@%s",
				os.Getenv("GITHUB_USER_NAME"),
				os.Getenv("GITHUB_ACCESS_TOKEN"),
				strings.Replace(ev.GetRepo().GetCloneURL(), "https://", "", 1)),
				logger,
		)
		switch ev.GetAction() {
		case "created":
			err = runner.Run(ctx, ev.GetComment().GetBody(), ev.GetSender().GetLogin())
		}
	}
	if err != nil {
		return nil, err
	}
	return &events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
	}, nil
}
