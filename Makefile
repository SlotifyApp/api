build_image:
	# Build API image, don't need to run this unless pushing to ECR
	docker build -f shared/docker/api.Dockerfile --tag slotify-api .

db_shell:
	# After 'make run', this will give you a shell to the mariadb
	docker exec -it mariadb-container mariadb -u root -prootpassword slotify

generate_api_docs:
	# Generate API markdown documentation using openapi spec
	docker run --rm  -v $(shell pwd):/local openapitools/openapi-generator-cli generate -i /local/shared/openapi/openapi.yaml -g markdown     -o /local/api_docs

generate:
	# Generate server Go code based on openapi spec
	go generate ./... && go mod tidy

lint:
	# Run Golangci-lint for Go linting
	docker run -t --rm -v $(shell pwd):/app -w /app golangci/golangci-lint:v1.63.4 golangci-lint run -v --fix ./...

run:
	# Docker compose up- start up API and DB in containers
	go generate ./... && go mod tidy && docker compose -f ./shared/docker/compose.local.yml up --build

stop:
	# Docker compose down- stop API and DB containers
	docker compose -f ./shared/docker/compose.local.yml down

stop_test:
	# Unless down is called, some data can be persisted in the db which can 
	# cause test fails
	docker compose -f ./shared/docker/compose.test.yml down

test:
	# Docker compose up- start DB and run API tests
	# Also shuts down after tests are finished (--abort-on-container-exit)
	go generate ./... && go mod tidy && docker compose -f ./shared/docker/compose.test.yml up --build --abort-on-container-exit 
