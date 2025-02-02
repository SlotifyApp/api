package api

import (
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
)

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000") // Allow frontend origin
		w.Header().Set("Access-Control-Allow-Credentials", "true")             // Allow cookies to be sent
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, UPDATE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// set it under the Authorization header.
func AuthMiddleware(next http.Handler) http.Handler {
	excludedPaths := map[string]bool{
		"/api/auth/callback": true,
		"/healthcheck":       true,
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if excludedPaths[r.URL.Path] {
			next.ServeHTTP(w, r)
			return
		}

		log.Printf("AuthMiddleware executed")

		accessTokenCookie, err := r.Cookie("access_token")
		if err != nil {
			log.Printf("failed to get access_token cookie: %s", err.Error())
			http.Error(w, "Missing or invalid token", http.StatusUnauthorized)
			return
		}
		if accessTokenCookie == nil || accessTokenCookie.Value == "" {
			log.Printf("cookie value of access token was not present")
			http.Error(w, "Missing or invalid token", http.StatusUnauthorized)
			return
		}

		r.Header.Set("Authorization", "Bearer "+accessTokenCookie.Value)

		// Some methods require refresh token and it is sent automatically
		// anyway in any request from the frontend
		refreshTokenCookie, err := r.Cookie("refresh_token")
		if err != nil {
			log.Printf("failed to get refresh_token cookie: %s", err.Error())
			http.Error(w, "Missing or invalid token", http.StatusUnauthorized)
			return
		}
		if refreshTokenCookie == nil || refreshTokenCookie.Value == "" {
			log.Printf("cookie value of refresh token was not present")
			http.Error(w, "Missing or invalid token", http.StatusUnauthorized)
			return
		}

		r.Header.Set("Refreshtoken", refreshTokenCookie.Value)

		next.ServeHTTP(w, r)
	})
}

func JWTMiddleware(next http.Handler) http.Handler {
	excludedPaths := map[string]bool{
		"/api/auth/callback": true,
		"/healthcheck":       true,
		"/user/logout":       true,
		"/refresh":           true,
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if excludedPaths[r.URL.Path] {
			next.ServeHTTP(w, r)
			return
		}

		log.Printf("JWTMiddleware executed")
		accessToken, err := jwt.GetJWTFromRequest(r)
		if err != nil {
			log.Printf("jwt middleware error: %s", err.Error())
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		if _, err = jwt.ParseJWT(accessToken, jwt.AccessTokenJWTSecretEnv); err != nil {
			log.Printf("jwt middleware error: %s", err.Error())
			http.Error(w, "Failed to parse JWT", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func ApplyMiddlewares(r *mux.Router, swagger *openapi3.T) {
	middlewares := []mux.MiddlewareFunc{
		CORSMiddleware,
		AuthMiddleware,
		oapi_middleware.OapiRequestValidator(swagger),
		JWTMiddleware,
		chi_middleware.Logger,
		chi_middleware.AllowContentType("application/json"),
		httprate.LimitByIP(requestLimit, 1*time.Minute),
		chi_middleware.Recoverer,
	}

	for _, middleware := range middlewares {
		r.Use(middleware)
	}
}
