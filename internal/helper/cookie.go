package helper

import (
	"net/http"
	"time"

	"github.com/google/uuid"
)

func CreateSessionCookie(cookieDuration time.Duration) http.Cookie {
	sessionId := uuid.New().String()
	expires := time.Now().Add(cookieDuration)

	return http.Cookie{
		Name:     "session_id",
		Value:    sessionId,
		Expires:  expires,
		Secure:   true,
		HttpOnly: true,
		Path:     "/",
	}
}

func RefreshSessionCookie(sessionCookie *http.Cookie, cookieDuration time.Duration) {
	if sessionCookie == nil {
		return
	}

	sessionCookie.Expires = time.Now().Add(cookieDuration)
}
