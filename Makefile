
build_image:
	# Build API image, don't need to run this unless pushing to ECR
	docker build -f shared/docker/api.Dockerfile --tag slotify-api .

db_shell:
	# After 'make run', this will give you a shell to the mariadb
	docker exec -it mariadb-container mariadb -u root -prootpassword slotify

.PHONY: generate
generate:
	# Generate sqlc files
	docker run --rm -v $(shell pwd):/src -w /src sqlc/sqlc generate
	# Generate server Go code based on openapi spec
	# Generate mocks
	go generate ./... 
	go mod tidy

lint:
	# Run Golangci-lint for Go linting
	docker run -t --rm -v $(shell pwd):/app -w /app golangci/golangci-lint:v1.64.5 golangci-lint run -v --fix ./...

run:
	# First docker compose down, some data can be persisted in the db
	docker compose -f ./shared/docker/compose.local.yml down || true
	# Docker compose up- start up API and DB in containers
	go generate ./... && go mod tidy && docker compose -f ./shared/docker/compose.local.yml up --build

stop:
	# Docker compose down- stop API and DB containers
	docker compose -f ./shared/docker/compose.local.yml down

test:
	# First docker compose down, some data can be persisted in the db which can
	# cause test fails
	# Docker compose up- start DB and run API tests
	# Also shuts down after tests are finished (--abort-on-container-exit)
	docker compose -f ./shared/docker/compose.test.yml down || true
	go generate ./... && go mod tidy
	docker compose -f ./shared/docker/compose.test.yml up --build --abort-on-container-exit || exit 1
