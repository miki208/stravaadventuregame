package auth

import (
	"net/http"

	"github.com/miki208/stravaadventuregame/internal/application"
	"github.com/miki208/stravaadventuregame/internal/helper"
)

func Deauthorize(resp http.ResponseWriter, req *http.Request, app *application.App, session helper.Session) {
	// deauthorize with the Strava API first
	err := app.StravaSvc.Deauthorize(session.UserId, app.SqlDb, nil)
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)

		return
	}

	// then log out the user
	Logout(resp, req, app, session)
}
