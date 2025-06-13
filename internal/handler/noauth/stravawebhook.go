package noauth

import (
	"encoding/json"
	"net/http"

	"github.com/miki208/stravaadventuregame/internal/application"
	"github.com/miki208/stravaadventuregame/internal/service/strava"
)

func StravaWebhookCallback(resp http.ResponseWriter, req *http.Request, app *application.App) {
	if req.Method == http.MethodGet {
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
			http.Error(resp, "Invalid hub.verify_token", http.StatusForbidden)

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
}
