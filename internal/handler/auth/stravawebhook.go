package auth

import (
	"fmt"
	"net/http"

	"github.com/miki208/stravaadventuregame/internal/application"
	"github.com/miki208/stravaadventuregame/internal/handler"
	"github.com/miki208/stravaadventuregame/internal/model"
)

func CreateStravaWebhookSubscription(resp *handler.ResponseWithSession, req *http.Request, app *application.App) error {
	athlete := model.NewAthlete()
	found, err := athlete.Load(resp.Session().UserId, app.SqlDb, nil)
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

	subscrExists, err := app.FileDb.Exists("strava", "webhooksubscription")
	if err != nil {
		return handler.NewHandlerError(http.StatusInternalServerError, err)
	}

	if subscrExists {
		return handler.NewHandlerError(http.StatusBadRequest, fmt.Errorf("webhook subscription already exists"))
	}

	subscription, err := app.StravaSvc.CreateSubscription(app.GetFullWebhookCallbackUrl())
	if err != nil {
		return handler.NewHandlerError(http.StatusInternalServerError, err)
	}

	err = app.FileDb.Write("strava", "webhooksubscription", &model.StravaWebhookSubscription{Id: subscription})
	if err != nil {
		return handler.NewHandlerError(http.StatusInternalServerError, err)
	}

	http.Redirect(resp, req, app.AdminPanelPage, http.StatusFound)

	return nil
}

func DeleteStravaWebhookSubscription(resp *handler.ResponseWithSession, req *http.Request, app *application.App) error {
	athlete := model.NewAthlete()
	found, err := athlete.Load(resp.Session().UserId, app.SqlDb, nil)
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

	var webhookSubscription model.StravaWebhookSubscription
	err = app.FileDb.Read("strava", "webhooksubscription", &webhookSubscription)
	if err != nil {
		return handler.NewHandlerError(http.StatusInternalServerError, err)
	}

	err = app.StravaSvc.DeleteSubscription(webhookSubscription.Id)
	if err != nil {
		return handler.NewHandlerError(http.StatusInternalServerError, err)
	}

	err = app.FileDb.Delete("strava", "webhooksubscription")
	if err != nil {
		return handler.NewHandlerError(http.StatusInternalServerError, err)
	}

	http.Redirect(resp, req, app.AdminPanelPage, http.StatusFound)

	return nil
}
