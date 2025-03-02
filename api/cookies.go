package api

import (
	"net/http"
	"time"
)

const (
	AccessTokenCookieExpiryHours  = 2
	RefreshTokenCookieExpiryHours = 24 * 7 // 7 days
)

// RemoveCookies will expire and remove the access_token and refresh_token HTTP-only cookies on the frontend.
func RemoveCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
		Secure:   true,
		Expires:  time.Unix(0, 0),
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
		Secure:   true,
		Expires:  time.Unix(0, 0),
	})
}

// CreateCookies will set the access_token and refresh_token HTTP-only cookies on the frontend.
func CreateCookies(w http.ResponseWriter, accessToken, refreshToken string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		Expires:  time.Now().Add(time.Hour * AccessTokenCookieExpiryHours),
	})

	// Set Refresh Token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		Expires:  time.Now().Add(time.Hour * RefreshTokenCookieExpiryHours),
	})
}
