package auth

import (
	"net/http"

	"github.com/miki208/stravaadventuregame/internal/application"
	"github.com/miki208/stravaadventuregame/internal/handler"
)

func Logout(resp *handler.ResponseWithSession, req *http.Request, app *application.App) error {
	// destroy the session in the session manager
	app.SessionMgr.DestroySession(resp.Session())

	resp.InvalidateSession()

	http.Redirect(resp, req, app.DefaultPageLoggedOutUsers, http.StatusFound)

	return nil
}
