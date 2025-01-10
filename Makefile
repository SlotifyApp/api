lint:
	docker run -t --rm -v $(shell pwd):/app -w /app golangci/golangci-lint:v1.63.4 golangci-lint run -v --fix ./...

generate:
	go generate ./... && go mod tidy

run:
	go generate ./... && go mod tidy && go run cmd/server/main.go

generate_api_docs:
	docker run --rm  -v $(shell pwd):/local openapitools/openapi-generator-cli generate -i /local/openapi.yaml -g markdown     -o /local/api_docs
