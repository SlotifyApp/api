package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
)

type uuidCtxKey struct {
	uuid string
}

type ctxKey string

const (
	ReqUUIDKey ctxKey = "reqUUID"
	ReqHeader  string = "X-Request-ID"
)

func WriteReqUUID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		reqUUID := uuid.New().String()
		reqCtxKey := uuidCtxKey{uuid: reqUUID}
		ctx := context.WithValue(req.Context(), ReqUUIDKey, reqCtxKey)
		writer.Header().Set(ReqHeader, reqUUID)
		next.ServeHTTP(writer, req.WithContext(ctx))
	})
}

func ReadReqUUID(req *http.Request) (string, error) {
	if reqCtxKey, ok := req.Context().Value(ReqUUIDKey).(uuidCtxKey); ok {
		return reqCtxKey.uuid, nil
	}
	return "", errors.New("error: unable to fetch id of this request")
}
