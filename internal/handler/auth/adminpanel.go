package auth

import (
	"fmt"
	"net/http"

	"github.com/miki208/stravaadventuregame/internal/application"
	"github.com/miki208/stravaadventuregame/internal/handler"
	"github.com/miki208/stravaadventuregame/internal/helper"
	"github.com/miki208/stravaadventuregame/internal/model"
)

func AdminPanel(resp http.ResponseWriter, req *http.Request, app *application.App, session helper.Session) error {
	athlete := model.NewAthlete()
	found, err := athlete.Load(session.UserId, app.SqlDb, nil)
	if err != nil {
		return handler.NewHandlerError(http.StatusInternalServerError, err)
	}

	if !found {
		return handler.NewHandlerError(http.StatusInternalServerError, fmt.Errorf("athlete not found"))
	}

	if !athlete.IsAdmin() {
		http.Redirect(resp, req, app.DefaultPageLoggedInUsers, http.StatusFound)

		return nil
	}

	// check if subscription ID is already available
	exists, err := app.FileDb.Exists("strava", "webhooksubscription")
	if err != nil {
		return handler.NewHandlerError(http.StatusInternalServerError, err)
	}

	var webhooksubscription model.StravaWebhookSubscription
	if exists {
		err = app.FileDb.Read("strava", "webhooksubscription", &webhooksubscription)
		if err != nil {
			return handler.NewHandlerError(http.StatusInternalServerError, err)
		}
	}

	// render the admin panel page
	err = app.Templates.ExecuteTemplate(resp, "adminpanel.html", webhooksubscription)
	if err != nil {
		return handler.NewHandlerError(http.StatusInternalServerError, err)
	}

	return nil
}
