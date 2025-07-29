package noauth

import (
	"net/http"

	"github.com/miki208/stravaadventuregame/internal/application"
)

func StravaWebhookCallback(resp http.ResponseWriter, req *http.Request, app *application.App) error {
	app.StravaSvc.StravaWebhookCallback(resp, req, app.SqlDb, app.SessionMgr)

	return nil
}
