package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strings"
)

const (
	csrfCookieName = "csrf_token"
	csrfHeaderName = "X-CSRF-Token"
	tokenLength    = 32
)

func generateToken() (string, error) {
	b := make([]byte, tokenLength)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func CSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(csrfCookieName)
		var token string

		if err != nil || cookie.Value == "" {
			token, err = generateToken()
			if err != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			http.SetCookie(w, &http.Cookie{
				Name:     csrfCookieName,
				Value:    token,
				Path:     "/",
				HttpOnly: false,
				SameSite: http.SameSiteStrictMode,
				Secure:   r.TLS != nil,
			})
		} else {
			token = cookie.Value
		}

		if r.Method == http.MethodPost || r.Method == http.MethodPut ||
			r.Method == http.MethodPatch || r.Method == http.MethodDelete {

			headerToken := r.Header.Get(csrfHeaderName)
			if headerToken == "" || !strings.EqualFold(headerToken, token) {
				http.Error(w, "invalid csrf token", http.StatusForbidden)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
