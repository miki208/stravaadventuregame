package auth

import (
	"net/http"

	"github.com/miki208/stravaadventuregame/internal/application"
	"github.com/miki208/stravaadventuregame/internal/handler"
	"github.com/miki208/stravaadventuregame/internal/helper"
)

func Deauthorize(resp http.ResponseWriter, req *http.Request, app *application.App, session helper.Session) error {
	// deauthorize with the Strava API first
	err := app.StravaSvc.Deauthorize(session.UserId, true, app.SqlDb, nil)
	if err != nil {
		handler.NewHandlerError(http.StatusInternalServerError, err)
	}

	// then log out the user
	Logout(resp, req, app, session)

	return nil
}
