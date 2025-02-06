package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/SlotifyApp/slotify-backend/api"
	"github.com/SlotifyApp/slotify-backend/database"
	"github.com/gorilla/mux"
)

const (
	readHeaderTimeout = 5
)

func main() {
	swagger, err := api.GetSwagger()
	if err != nil {
		log.Fatalf("error getting swagger/openapi spec: %s", err.Error())
	}
	ctx := context.Background()

	db, err := database.NewDatabaseWithContext(ctx)
	if err != nil {
		log.Fatalf("error creating db: %s", err.Error())
	}

	server, err := api.NewServerWithContext(ctx, db)
	if err != nil {
		log.Fatalf("error creating server: %s", err.Error())
	}

	r := mux.NewRouter()

	api.ApplyMiddlewares(r, swagger)

	h := api.HandlerFromMux(server, r)

	s := &http.Server{
		Handler:           h,
		Addr:              "0.0.0.0:8080",
		ReadHeaderTimeout: readHeaderTimeout * time.Second,
	}

	server.Logger.Info("http server starting up")

	log.Fatal(s.ListenAndServe())
}
