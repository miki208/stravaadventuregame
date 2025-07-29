package auth

import (
	"net/http"
	"time"

	"github.com/miki208/stravaadventuregame/internal/application"
	"github.com/miki208/stravaadventuregame/internal/helper"
)

func Logout(resp http.ResponseWriter, req *http.Request, app *application.App, session helper.Session) error {
	// destroy the session in the session manager
	app.SessionMgr.DestroySession(session)

	// make the cookie invalid by setting an expiration time in the past and set it in the response
	session.SessionCookie.Expires = time.Now().Add(-time.Hour)
	http.SetCookie(resp, &session.SessionCookie)

	http.Redirect(resp, req, app.DefaultPageLoggedOutUsers, http.StatusFound)

	return nil
}
