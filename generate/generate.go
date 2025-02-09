package generate

//go:generate go run -modfile=../tools/go.mod github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -config oapi_codegen_cfg.yaml ../shared/openapi/openapi.yaml
//go:generate go run -modfile=../tools/go.mod go.uber.org/mock/mockgen -source=../database/notification.go -destination=../mocks/mock_notification_db.go -package mocks
//go:generate go run -modfile=../tools/go.mod go.uber.org/mock/mockgen -source=../notification/notification.go -destination=../mocks/mock_notification_service.go -package mocks
