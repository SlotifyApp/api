package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	chi_middleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/gorilla/mux"
	oapi_middleware "github.com/oapi-codegen/nethttp-middleware"
	"github.com/saths008/slotify-backend/api"
)

const (
	readHeaderTimeout = 5
	requestLimit      = 100
)

func applyMiddlewares(r *mux.Router, swagger *openapi3.T) {
	middlewares := []mux.MiddlewareFunc{
		chi_middleware.Logger,
		chi_middleware.AllowContentType("application/json"),
		httprate.LimitByIP(requestLimit, 1*time.Minute),
		oapi_middleware.OapiRequestValidator(swagger),
		chi_middleware.Recoverer,
	}

	for _, middleware := range middlewares {
		r.Use(middleware)
	}
}

func main() {
	swagger, err := api.GetSwagger()
	if err != nil {
		log.Fatalf("error getting swagger/openapi spec: %s", err.Error())
	}
	ctx := context.Background()
	server, err := api.NewServerWithContext(ctx)
	if err != nil {
		log.Fatalf("error creating new server: %s", err.Error())
	}

	r := mux.NewRouter()

	applyMiddlewares(r, swagger)

	h := api.HandlerFromMux(server, r)

	s := &http.Server{
		Handler:           h,
		Addr:              "0.0.0.0:8080",
		ReadHeaderTimeout: readHeaderTimeout * time.Second,
	}

	server.Logger.Info("http server starting up")

	log.Fatal(s.ListenAndServe())
}
