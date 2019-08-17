package github

import (
	"context"
	"github.com/sambaiz/cdkbot/functions/operation/constant"
	"github.com/sambaiz/cdkbot/functions/operation/eventhandler"
	"github.com/sambaiz/cdkbot/functions/operation/platform/github/client"
	"gopkg.in/src-d/go-git.v4"
	"net/http"
	"os"

	"go.uber.org/zap"

	"github.com/aws/aws-lambda-go/events"
	goGitHub "github.com/google/go-github/v26/github"
	goGitHttp "gopkg.in/src-d/go-git.v4/plumbing/transport/http"
)

var (
	gitHubWebhookSecret = os.Getenv("GITHUB_WEBHOOK_SECRET")
)

// Handler handles GitHub webhook
func Handler(
	ctx context.Context,
	req events.APIGatewayProxyRequest,
	logger *zap.Logger,
) (events.APIGatewayProxyResponse, error) {
	if err := goGitHub.ValidateSignature(req.Headers["X-Hub-Signature"], []byte(req.Body), []byte(gitHubWebhookSecret)); err != nil {
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

	cloneAuth := &goGitHttp.BasicAuth{
		Username: os.Getenv("GITHUB_USER_NAME"),
		Password: os.Getenv("GITHUB_ACCESS_TOKEN"),
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
			&git.CloneOptions{
				URL: ev.GetRepo().GetCloneURL(),
				Auth: cloneAuth,
			},
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
			&git.CloneOptions{
				URL: ev.GetRepo().GetCloneURL(),
				Auth: cloneAuth,
			},
		)
		switch ev.GetAction() {
		case "created":
			nameToLabel := map[string]constant.Label{}
			for _, label := range ev.GetIssue().Labels {
				if label.Name == nil {
					continue
				}
				if lb, ok := constant.NameToLabel[*label.Name]; ok {
					nameToLabel[lb.Name] = constant.NameToLabel[lb.Name]
				}
			}
			err = eventHandler.CommentCreated(
				ctx,
				ev.GetComment().GetBody(),
				nameToLabel)
		}
	}
	if err != nil {
		logger.Error("Failed to event an event", zap.Error(err))
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, err
	}
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
	}, nil
}
