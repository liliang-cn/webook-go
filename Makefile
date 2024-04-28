.PHONY: docker

docker:
	@rm webook || true
	@go mod tidy
	@GOOS=linux GOARCH=arm64 go build -o webook .
	@docker rmi -f leoleecn/webook:v0.0.1 || true
	@docker build -t leoleecn/webook:v0.0.1 .

