package api

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type ctxKey string

const (
	ReqUUIDKey ctxKey = "reqUUID"
	ReqHeader  string = "X-Request-ID"
)

func WriteReqUUID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		reqUUID := uuid.New().String()
		ctx := context.WithValue(req.Context(), ReqUUIDKey, reqUUID)
		writer.Header().Set(string(ReqHeader), reqUUID)
		next.ServeHTTP(writer, req.WithContext(ctx))
	})
}

func ReadReqUUID(req *http.Request) string {
	if reqUUID, ok := req.Context().Value(ReqUUIDKey).(string); ok {
		return reqUUID
	} else {
		return ""
	}
}
