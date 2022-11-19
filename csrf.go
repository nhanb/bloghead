package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

const csrfCookieName = "csrf"
const csrfInputName = "csrf-token"

// CsrfCheck performs a [Double Submit Cookie] check against CSRF.
//
// On success, it returns html markup for a hidden <input> tag,
// ready to be put inside a <form>.
//
// On failure, it writes a 403 response and returns "".
//
// [Double Submit Cookie]: https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html#double-submit-cookie
func CsrfCheck(w http.ResponseWriter, r *http.Request) (csrfToken template.HTML) {
	csrfCookie, err := r.Cookie(csrfCookieName)
	if err == http.ErrNoCookie {
		// If previous csrf cookie expired or this is the first request,
		// then generate & set the cookie
		csrfCookie = &http.Cookie{
			Name:     csrfCookieName,
			Value:    generateSecureToken(32),
			MaxAge:   60 * 60 * 24 * 7,
			SameSite: http.SameSiteStrictMode,
		}
		http.SetCookie(w, csrfCookie)
	} else if err != nil {
		// Unexpected error, who knows!
		log.Fatalf("Error reading csrf cookie: %v", err)
	} else if r.Method == "POST" {
		// All POST requests must provide a hidden form input that matches
		// the csrf token from cookie.
		csrfTokenFromForm := r.FormValue(csrfInputName)
		if csrfTokenFromForm == "" || csrfTokenFromForm != csrfCookie.Value {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("Failed CSRF token check."))
			return ""
		}
	}

	return template.HTML(fmt.Sprintf(
		`<input class="csrf-input" type="hidden" name="%s" value="%s" />`,
		csrfInputName, csrfCookie.Value,
	))
}

// Does what it says on the tin.
// Panics if unable to get crytographically secure bytes.
func generateSecureToken(numBytes int) string {
	b := make([]byte, numBytes)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}
