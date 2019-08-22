package github

import (
	"context"
	"fmt"
	"github.com/sambaiz/cdkbot/functions/operation/eventhandler"
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
	logger *zap.Logger,
) (events.APIGatewayProxyResponse, error) {
	if err := goGitHub.ValidateSignature(
		req.Headers["X-Hub-Signature"],
		[]byte(req.Body),
		[]byte(os.Getenv("GITHUB_WEBHOOK_SECRET")),
	); err != nil {
		logger.Info("Signature is invalid", zap.Error(err))
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
		}, nil
	}
	hook, err := goGitHub.ParseWebHook(req.Headers["X-GitHub-Event"], []byte(req.Body))
	if err != nil {
		logger.Error("Failed to parse hook", zap.Error(err))
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
		}, nil
	}

	switch ev := hook.(type) {
	case *goGitHub.PullRequestEvent:
		eventHandler := eventhandler.New(
			ctx,
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
		)
		switch ev.GetAction() {
		case "opened":
			err = eventHandler.PullRequestOpened(ctx)
		}
	case *goGitHub.IssueCommentEvent:
		eventHandler := eventhandler.New(
			ctx,
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
		)
		switch ev.GetAction() {
		case "created":
			err = eventHandler.CommentCreated(
				ctx,
				ev.GetSender().GetLogin(),
				ev.GetComment().GetBody())
		}
	}
	if err != nil {
		logger.Error("failed to handle an event", zap.Error(err))
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, err
	}
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
	}, nil
}
