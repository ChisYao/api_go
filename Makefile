.PHONY: docker
docker:
	@rm webook || true
	@go mod tidy
	@GOOS=linux GOARCH=arm go build -o webook .
	@docker rmi -f chis/webook
	@docker build -t chis/webook .
