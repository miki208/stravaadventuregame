package auth

import (
	"net/http"

	"github.com/miki208/stravaadventuregame/internal/application"
	"github.com/miki208/stravaadventuregame/internal/handler"
)

func Deauthorize(resp *handler.ResponseWithSession, req *http.Request, app *application.App) error {
	// deauthorize with the Strava API first
	err := app.StravaSvc.Deauthorize(resp.Session().UserId, true, app.SqlDb, nil)
	if err != nil {
		handler.NewHandlerError(http.StatusInternalServerError, err)
	}

	// then log out the user
	Logout(resp, req, app)

	return nil
}
