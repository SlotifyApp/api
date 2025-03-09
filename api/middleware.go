package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/SlotifyApp/slotify-backend/jwt"
	"github.com/getkin/kin-openapi/openapi3"

	chi_middleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/gorilla/mux"
	oapi_middleware "github.com/oapi-codegen/nethttp-middleware"
)

const (
	requestLimit = 100
	ReqHeader    = "X-Request-ID"
)

// RefreshTokenCtxKey is the key in context value for the refresh token value.
type RefreshTokenCtxKey struct{}

// UserIDCtxKey is the key in context value for the user id value.
type UserIDCtxKey struct{}

// RequestIDCtxKey is the key in context value for the request id.
type RequestIDCtxKey struct{}

// AuthMiddleware takes the http-only cookies and sets the access token under the Authorization header.
// The refresh token is set in the request context.
func AuthMiddleware(next http.Handler) http.Handler {
	// Paths to ignore this
	excludedPaths := map[string]bool{
		"/api/auth/callback": true, // http cookie is not set before logging in ie. during OAuth flow
		"/api/healthcheck":   true, // http cookie doesn't need to present for a healthcheck
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if excludedPaths[r.URL.Path] {
			next.ServeHTTP(w, r)
			return
		}

		// Get access_token cookie.
		accessTokenCookie, err := r.Cookie("access_token")
		if err != nil {
			log.Printf("error fetching access_token cookie: route: %s, err: %s", r.URL.Path, err.Error())
			http.Error(w, "Missing or invalid token", http.StatusUnauthorized)
			return
		}
		if accessTokenCookie == nil || accessTokenCookie.Value == "" {
			log.Printf("access_token was nil/empty: route: %s", r.URL.Path)
			http.Error(w, "Missing or invalid token", http.StatusUnauthorized)
			return
		}

		r.Header.Set("Authorization", fmt.Sprintf("Bearer: %s", accessTokenCookie.Value))

		// Some methods require refresh token and it is sent automatically
		// anyway in any request from the frontend
		// set refresh token in request context
		refreshTokenCookie, err := r.Cookie("refresh_token")
		if err != nil {
			log.Printf("error fetching refresh_token cookie: route: %s, err: %s", r.URL.Path, err.Error())
			http.Error(w, "Missing or invalid token", http.StatusUnauthorized)
			return
		}
		if refreshTokenCookie == nil || refreshTokenCookie.Value == "" {
			log.Printf("refresh_token was nil/empty: route: %s", r.URL.Path)
			http.Error(w, "Missing or invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), RefreshTokenCtxKey{}, refreshTokenCookie)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get(ReqHeader)
		if reqID == "" {
			sendError(w, http.StatusInternalServerError,
				fmt.Sprintf("failed to get %s header", ReqHeader))
			return
		}

		ctx := context.WithValue(r.Context(), RequestIDCtxKey{}, reqID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// JWTMiddleware parses and validates the access token, and stores the userID in the request context.
func JWTMiddleware(next http.Handler) http.Handler {
	excludedPaths := map[string]bool{
		"/api/auth/callback": true,
		"/api/healthcheck":   true,
		"/api/users/logout":  true,
		"/api/refresh":       true,
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if excludedPaths[r.URL.Path] {
			next.ServeHTTP(w, r)
			return
		}

		userID, err := jwt.GetUserIDFromReq(r)
		if err != nil {
			log.Printf("failed to get userid from jwt tokens: %+v", err)
			sendError(w, http.StatusUnauthorized, "failed to get userid from jwt access token")
			return
		}

		// set userID in context so it's available in our requests
		// and the access token doesn't need to be parsed again
		ctx := context.WithValue(r.Context(), UserIDCtxKey{}, userID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ApplyMiddlewares applies all the middleware functions for the server.
func ApplyMiddlewares(r *mux.Router, swagger *openapi3.T) {
	middlewares := []mux.MiddlewareFunc{
		// Adds request id to the request context
		RequestIDMiddleware,

		// makes sure that requests and responses follow openapischema
		oapi_middleware.OapiRequestValidator(swagger),

		AuthMiddleware,

		JWTMiddleware,

		chi_middleware.AllowContentType("application/json", "text/event-stream"),

		// rate limitter
		httprate.LimitByIP(requestLimit, 1*time.Minute),

		// returns 500 in case of panics instead of stopping API.
		chi_middleware.Recoverer,
	}

	for _, middleware := range middlewares {
		r.Use(middleware)
	}
}
