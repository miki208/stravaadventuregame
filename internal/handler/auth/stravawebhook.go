package auth

import (
	"net/http"

	"github.com/miki208/stravaadventuregame/internal/application"
	"github.com/miki208/stravaadventuregame/internal/helper"
	"github.com/miki208/stravaadventuregame/internal/model"
)

func CreateStravaWebhookSubscription(resp http.ResponseWriter, req *http.Request, app *application.App, session helper.Session) {
	athlete := model.NewAthlete()
	found, err := athlete.Load(session.UserId, app.SqlDb, nil)
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

	subscrExists, err := app.FileDb.Exists("strava", "webhooksubscription")
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)

		return
	}

	if subscrExists {
		http.Error(resp, "Webhook subscription already exists.", http.StatusBadRequest)

		return
	}

	subscription, err := app.StravaSvc.CreateSubscription(app.GetFullWebhookCallbackUrl())
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)

		return
	}

	err = app.FileDb.Write("strava", "webhooksubscription", &model.StravaWebhookSubscription{Id: subscription})
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)

		return
	}

	http.Redirect(resp, req, app.AdminPanelPage, http.StatusFound)
}

func DeleteStravaWebhookSubscription(resp http.ResponseWriter, req *http.Request, app *application.App, session helper.Session) {
	athlete := model.NewAthlete()
	found, err := athlete.Load(session.UserId, app.SqlDb, nil)
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

	var webhookSubscription model.StravaWebhookSubscription
	err = app.FileDb.Read("strava", "webhooksubscription", &webhookSubscription)
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)

		return
	}

	err = app.StravaSvc.DeleteSubscription(webhookSubscription.Id)
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)

		return
	}

	err = app.FileDb.Delete("strava", "webhooksubscription")
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)

		return
	}

	http.Redirect(resp, req, app.AdminPanelPage, http.StatusFound)
}
