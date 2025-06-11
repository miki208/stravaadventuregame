package helper

import (
	"net/http"
	"time"

	"github.com/google/uuid"
)

func CreateSessionCookie() http.Cookie {
	sessionId := uuid.New().String()
	expires := time.Now().Add(30 * time.Minute)

	return http.Cookie{
		Name:     "session_id",
		Value:    sessionId,
		Expires:  expires,
		Secure:   true,
		HttpOnly: true,
		Path:     "/",
	}
}
