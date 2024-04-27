package functions

import (
	"net/http"
	"time"

	"github.com/gofrs/uuid"
)

// function that generate a session token
func GenerateSessionToken() (string, error) {
	sessionID, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	return sessionID.String(), nil
}

// function that set the session token
func SetSessionToken(token string, w http.ResponseWriter) {
	cookie := http.Cookie{
		Name:     "session-token",
		Value:    token,
		HttpOnly: true,
		Secure:   true,
	}
	http.SetCookie(w, &cookie)
}

// function that get the session token
func GetSessionToken(r *http.Request) string {
	cookie, err := r.Cookie("session-token")
	if err != nil {
		return ""
	}
	return cookie.Value
}

// function that clear the session token
func ClearSessionToken(w http.ResponseWriter) {
	cookie := http.Cookie{
		Name:     "session-token",
		Value:    "",
		HttpOnly: true,
		Expires:  time.Now().AddDate(0, 0, -1),
	}
	http.SetCookie(w, &cookie)
}
