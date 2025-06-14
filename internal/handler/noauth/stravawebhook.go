package noauth

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/miki208/stravaadventuregame/internal/application"
	"github.com/miki208/stravaadventuregame/internal/model"
	"github.com/miki208/stravaadventuregame/internal/service/strava"
)

func StravaWebhookCallback(resp http.ResponseWriter, req *http.Request, app *application.App) {
	if req.Method == http.MethodGet {
		handleWebhookForSubscriptionValidation(resp, req, app)
	} else if req.Method == http.MethodPost {
		requestBody, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(resp, "Failed to read request body", http.StatusInternalServerError)

			return
		}

		var webhookEvent model.StravaWebhookEvent
		err = json.Unmarshal(requestBody, &webhookEvent)
		if err != nil {
			http.Error(resp, "Failed to unmarshal webhook event", http.StatusInternalServerError)

			return
		}

		if webhookEvent.ObjectType == "athlete" {
			handleWebhookForAthlete(&webhookEvent, app)
		} else if webhookEvent.ObjectType == "activity" {
			handleWebhookForActivity(&webhookEvent, app)
		}

		resp.WriteHeader(http.StatusOK)
	}
}

func handleWebhookForSubscriptionValidation(resp http.ResponseWriter, req *http.Request, app *application.App) {
	query := req.URL.Query()

	if !query.Has("hub.challenge") || !query.Has("hub.mode") || !query.Has("hub.verify_token") {
		http.Error(resp, "Missing required query parameters", http.StatusBadRequest)

		return
	}

	if query.Get("hub.mode") != "subscribe" {
		http.Error(resp, "Invalid hub.mode", http.StatusBadRequest)

		return
	}

	if query.Get("hub.verify_token") != app.StravaSvc.GetVerifyToken() {
		http.Error(resp, "Invalid hub.verify_token", http.StatusBadRequest)

		return
	}

	challengeResponse := strava.CallbackValidationResponse{HubChallenge: query.Get("hub.challenge")}
	challengeResponseJson, err := json.Marshal(&challengeResponse)
	if err != nil {
		http.Error(resp, "Failed to marshal challenge response", http.StatusInternalServerError)

		return
	}

	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(http.StatusOK)

	_, err = resp.Write(challengeResponseJson)
	if err != nil {
		http.Error(resp, "Failed to write challenge response", http.StatusInternalServerError)

		return
	}
}

func handleWebhookForAthlete(webhookEvent *model.StravaWebhookEvent, app *application.App) {
	authorizedUpdate, ok := webhookEvent.Updates["authorized"]
	if ok && authorizedUpdate == "false" {
		// atlete is revoking access, we're going to delete the athlete from the database
		app.StravaSvc.Deauthorize(webhookEvent.ObjectId, false, app.SqlDb, nil)

		session := app.SessionMgr.GetSessionByUserId(webhookEvent.ObjectId)
		if session != nil {
			app.SessionMgr.DestroySession(*session)
		}
	}
}

func handleWebhookForActivity(webhookEvent *model.StravaWebhookEvent, app *application.App) {
	// we need to be quick here, we're just going to queue the activity for processing

	webhookEventSummary := webhookEvent.GetSummary()
	webhookEventSummary.Save(app.SqlDb, nil)
}
