package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/SlotifyApp/slotify-backend/api"
	"github.com/SlotifyApp/slotify-backend/cron"
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

	// Clear out the servers array in the swagger spec, that skips validating
	// that server names match.
	swagger.Servers = nil

	db, err := database.NewDatabaseWithContext(ctx)
	if err != nil {
		log.Fatalf("error creating db: %s", err.Error())
	}

	server, err := api.NewServerWithContext(ctx, db)
	if err != nil {
		log.Fatalf("error creating server: %s", err.Error())
	}
	if err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter()

	api.ApplyMiddlewares(r, swagger)

	h := api.HandlerFromMux(server, r)
	s := &http.Server{
		Handler:           h,
		Addr:              "0.0.0.0:8080",
		ReadHeaderTimeout: readHeaderTimeout * time.Second,
		ReadTimeout:       0, // No timeout (important for long-lived SSE)
		WriteTimeout:      0, // No timeout (important for long-lived SSE)
		IdleTimeout:       0, // No timeout
	}
	s.RegisterOnShutdown(func() {
		if err = db.DB.Close(); err != nil {
			log.Printf("failed to close database: %+v", err)
		}
		log.Printf("Shutting down server gracefully...")
	})

	server.Logger.Info("http server starting up")

	if err = cron.RegisterDBCronJobs(ctx, server.DB, server.Logger); err != nil {
		log.Fatalf("failed to register db cron jobs: %s", err.Error())
	}

	log.Fatal(s.ListenAndServe())
}
