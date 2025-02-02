package api

import (
	"net/http"
	"time"
)

const (
	AccessTokenCookieExpiryHours  = 2
	RefreshTokenCookieExpiryHours = 24 * 7 // 7 days
)

// and refresh_token HTTP-only cookies on the frontend.
func RemoveCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		MaxAge:   0, // Expire the cookie
		Path:     "/",
		HttpOnly: true,
		// TODO: Change to true when https is configured
		Secure: false,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		MaxAge:   0, // Expire the cookie
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
	})
}

func CreateCookies(w http.ResponseWriter, accessToken, refreshToken string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(time.Hour * AccessTokenCookieExpiryHours),
	})

	// Set Refresh Token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",      // Cookie name for refresh token
		Value:    refreshToken,         // Refresh token value
		Path:     "/",                  // Make the cookie available for the entire site
		HttpOnly: true,                 // Prevent client-side access to refresh token (mitigates XSS attacks)
		Secure:   false,                // Use "Secure" flag for HTTPS-only cookies
		SameSite: http.SameSiteLaxMode, // Restrict cookie to first-party context
		Expires:  time.Now().Add(time.Hour * RefreshTokenCookieExpiryHours),
	})
}
