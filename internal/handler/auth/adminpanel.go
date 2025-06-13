package auth

import (
	"net/http"

	"github.com/miki208/stravaadventuregame/internal/application"
	"github.com/miki208/stravaadventuregame/internal/helper"
	"github.com/miki208/stravaadventuregame/internal/model"
)

func AdminPanel(resp http.ResponseWriter, req *http.Request, app *application.App, session helper.Session) {
	var athlete model.Athlete
	found, err := athlete.LoadById(session.UserId, app.SqlDb, nil)
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)

		return
	}

	if !found {
		http.Error(resp, "Athlete not found.", http.StatusInternalServerError)

		return
	}

	if !athlete.IsAdmin() {
		http.Redirect(resp, req, app.DefaultPageLoggedInUsers, http.StatusFound)

		return
	}

	// check if subscription ID is already available
	exists, err := app.FileDb.Exists("strava", "webhooksubscription")
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)

		return
	}

	var webhooksubscription model.StravaWebhookSubscription
	if exists {
		err = app.FileDb.Read("strava", "webhooksubscription", &webhooksubscription)
		if err != nil {
			http.Error(resp, err.Error(), http.StatusInternalServerError)

			return
		}
	}

	// render the admin panel page
	err = app.Templates.ExecuteTemplate(resp, "adminpanel.html", webhooksubscription)
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)

		return
	}
}
