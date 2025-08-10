package helper

import (
	"net/http"
	"time"

	"github.com/google/uuid"
)

func CreateSessionCookie(cookieDuration time.Duration, proxyPathPrefix string) http.Cookie {
	sessionId := uuid.New().String()
	expires := time.Now().Add(cookieDuration)

	path := "/"
	if proxyPathPrefix != "" {
		path = proxyPathPrefix
	}

	return http.Cookie{
		Name:     "session_id",
		Value:    sessionId,
		Expires:  expires,
		Secure:   true,
		HttpOnly: true,
		Path:     path,
	}
}

func RefreshSessionCookie(sessionCookie *http.Cookie, cookieDuration time.Duration) {
	if sessionCookie == nil {
		return
	}

	sessionCookie.Expires = time.Now().Add(cookieDuration)
}
