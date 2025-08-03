package auth

import (
	"fmt"
	"net/http"

	"github.com/miki208/stravaadventuregame/internal/application"
	"github.com/miki208/stravaadventuregame/internal/handler"
	"github.com/miki208/stravaadventuregame/internal/helper"
	"github.com/miki208/stravaadventuregame/internal/model"
)

func AdminPanel(resp *handler.ResponseWithSession, req *http.Request, app *application.App) error {
	athlete := model.NewAthlete()
	found, err := athlete.Load(resp.Session().UserId, app.SqlDb, nil)
	if err != nil {
		return handler.NewHandlerError(http.StatusInternalServerError, err)
	}

	if !found {
		return handler.NewHandlerError(http.StatusInternalServerError, fmt.Errorf("athlete not found"))
	}

	isAdmin, err := helper.IsAthleteAdmin(athlete.Id, app.SqlDb, nil)
	if err != nil {
		return handler.NewHandlerError(http.StatusInternalServerError, err)
	}

	if !isAdmin {
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
