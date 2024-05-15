.PHONY: docker

docker:
	@rm webook || true
	@go mod tidy
	@GOOS=linux GOARCH=arm64 go build -tags=k8s -o webook .
	@docker rmi -f webook:v0.0.1 || true
	@docker build -t webook:v0.0.1 .

