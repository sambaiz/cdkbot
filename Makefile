.PHONY: clean build package deploy publish install-tools lint test _test doc

S3Bucket=cdkbot
Platform=github
Region=us-east-1

clean: 
	rm -rf ./functions/operation/operation
	rm -rf ./functions/webhook/webhook
	rm -rf layer
	
build:
	GOOS=linux GOARCH=amd64 go build -o functions/webhook/webhook ./functions/webhook

# make build-tasks-image Version=x.x.x
build-tasks-image:
	GOOS=linux GOARCH=amd64 go build -o tasks/operation/operation ./tasks/operation
	docker build -t sambaiz/cdkbot-operation:${Version} -f ./tasks/operation/Dockerfile .
	# docker push sambaiz/cdkbot-operation:${Version}

package: build
	sam package --output-template-file packaged.yaml --s3-bucket ${S3Bucket} --region ${Region}

deploy: package
	aws cloudformation deploy --parameter-overrides \
	SubnetID=${SubnetID} \
	Platform=${Platform} \
	GitHubUserName=${GitHubUserName} \
	GitHubAccessToken=${GitHubAccessToken} \
	GitHubWebhookSecret=${GitHubWebhookSecret} \
	--template-file packaged.yaml --stack-name cdkbot --capabilities CAPABILITY_IAM --region ${Region}

publish: package
	sam publish -t packaged.yaml --region ${Region}

install-tools:
	go get -u golang.org/x/lint/golint
	go get -u github.com/golang/mock/mockgen

lint:
	golint -set_exit_status $$(go list ./...)

test:
	docker build -t cdkbot-test -f ./test/Dockerfile .
	docker rm -f cdkbot-test || true
	docker run -itd --name cdkbot-test cdkbot-test /bin/sh
	docker cp . cdkbot-test:/root/cdkbot
	go mod download
	docker cp `go env GOPATH`/pkg/mod/cache cdkbot-test:/go/pkg/mod/cache
	docker exec cdkbot-test make _test
	docker rm -f cdkbot-test

_test:
	go test ./...

mock:
	mockgen -package mock -source functions/operation/cdk/cdk.go -destination functions/operation/cdk/mock/cdk_mock.go
	mockgen -package mock -source functions/operation/config/config.go -destination functions/operation/config/mock/config_mock.go
	mockgen -package mock -source functions/operation/git/git.go -destination functions/operation/git/mock/git_mock.go
	mockgen -package mock -source functions/operation/platform/client.go -destination functions/operation/platform/mock/client_mock.go
