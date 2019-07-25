.PHONY: clean build install-tools lint

clean: 
	rm -rf ./functions/github/github
	rm -rf ./functions/github-webhook/github-webhook
	rm -rf npm-layer
	
build:
	GOOS=linux GOARCH=amd64 go build -o functions/github/github ./functions/github
	GOOS=linux GOARCH=amd64 go build -o functions/github-webhook/github-webhook ./functions/github-webhook
	docker build -t npmbin ./npm-lambda-layer
	docker run npmbin cat /tmp/npm-layer.zip > npm-layer.zip && unzip npm-layer.zip -d npm-layer && rm npm-layer.zip

install-tools:
	go get -u golang.org/x/lint/golint

lint:
	golint -set_exit_status $$(go list ./...)
	go vet ./...