.PHONY: clean build install-tools lint test _test

clean: 
	rm -rf ./functions/github/github
	rm -rf ./functions/github-webhook/github-webhook
	rm -rf npm-layer
	
build:
	GOOS=linux GOARCH=amd64 go build -o functions/github/github ./functions/github
	GOOS=linux GOARCH=amd64 go build -o functions/github-webhook/github-webhook ./functions/github-webhook
	docker build -t cdkbot-npmbin ./npm-lambda-layer
	docker run cdkbot-npmbin cat /tmp/npm-layer.zip > npm-layer.zip && unzip npm-layer.zip -d npm-layer && rm npm-layer.zip

install-tools:
	go get -u golang.org/x/lint/golint
	go get -u github.com/golang/mock/mockgen

lint:
	golint -set_exit_status $$(go list ./...)

test:
	docker build -t cdkbot-npmbin ./npm-lambda-layer
	docker build -t cdkbot-test -f ./test/Dockerfile .
	docker rm -f cdkbot-test || true
	docker run -itd --name cdkbot-test cdkbot-test /bin/sh
	docker cp . cdkbot-test:/root/cdkbot
	docker exec cdkbot-test make _test
	docker rm -f cdkbot-test

_test:
	go test ./...

mock:
	mockgen -package mock -source lib/cdk/cdk.go -destination lib/cdk/mock/cdk_mock.go
	mockgen -package mock -source lib/config/config.go -destination lib/config/mock/config_mock.go
	mockgen -package mock -source lib/git/git.go -destination lib/git/mock/git_mock.go
	mockgen -package mock -source functions/github/client/client.go -destination functions/github/client/mock/client_mock.go
