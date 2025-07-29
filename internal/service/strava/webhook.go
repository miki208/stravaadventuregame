package strava

import (
	"database/sql"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/miki208/stravaadventuregame/internal/database"
	"github.com/miki208/stravaadventuregame/internal/helper"
	"github.com/miki208/stravaadventuregame/internal/model"
	"github.com/miki208/stravaadventuregame/internal/service/strava/externalmodel"
)

func (svc *Strava) StravaWebhookCallback(resp http.ResponseWriter, req *http.Request, db *sql.DB, sessionManager *helper.SessionManager) {
	switch req.Method {
	case http.MethodGet:
		svc.handleWebhookForSubscriptionValidation(resp, req, db)
	case http.MethodPost:
		requestBody, err := io.ReadAll(req.Body)
		if err != nil {
			slog.Error("strava_webhook > Failed to read request body.", "error", err)

			resp.WriteHeader(http.StatusInternalServerError)

			return
		}

		var webhookEvent externalmodel.StravaWebhookEvent
		err = json.Unmarshal(requestBody, &webhookEvent)
		if err != nil {
			slog.Error("strava_webhook > Failed to unmarshal webhook event.", "error", err)

			resp.WriteHeader(http.StatusInternalServerError)

			return
		}

		switch webhookEvent.ObjectType {
		case "athlete":
			svc.handleWebhookForAthlete(&webhookEvent, db, sessionManager)
		case "activity":
			svc.handleWebhookForActivity(&webhookEvent, db)
		}

		resp.WriteHeader(http.StatusOK)
	}
}

func (svc *Strava) handleWebhookForSubscriptionValidation(resp http.ResponseWriter, req *http.Request, db *sql.DB) {
	query := req.URL.Query()

	if !query.Has("hub.challenge") || !query.Has("hub.mode") || !query.Has("hub.verify_token") {
		slog.Error("strava_webhook > Missing required query parameters for Strava webhook validation")

		resp.WriteHeader(http.StatusBadRequest)

		return
	}

	if query.Get("hub.mode") != "subscribe" {
		slog.Error("strava_webhook > Invalid hub.mode for Strava webhook validation", "mode", query.Get("hub.mode"))

		resp.WriteHeader(http.StatusBadRequest)

		return
	}

	if query.Get("hub.verify_token") != svc.GetVerifyToken() {
		slog.Error("strava_webhook > Invalid hub.verify_token for Strava webhook validation", "token", query.Get("hub.verify_token"))

		resp.WriteHeader(http.StatusBadRequest)

		return
	}

	err := SendCallbackValidationResponse(query.Get("hub.challenge"), resp)
	if err != nil {
		slog.Error("strava_webhook > Failed to send callback validation response.", "error", err)

		resp.WriteHeader(http.StatusInternalServerError)

		return
	}

	resp.WriteHeader(http.StatusOK)
}

func (svc *Strava) handleWebhookForAthlete(webhookEvent *externalmodel.StravaWebhookEvent, db *sql.DB, sessionManager *helper.SessionManager) {
	authorizedUpdate, ok := webhookEvent.Updates["authorized"]
	if ok && authorizedUpdate == "false" {
		// atlete is revoking access, we're going to delete the athlete from the database
		if err := svc.Deauthorize(webhookEvent.ObjectId, false, db, nil); err != nil {
			slog.Error("strava_webhook > Failed to deauthorize athlete.", "error", err, "athlete_id", webhookEvent.ObjectId)

			return
		}

		session := sessionManager.GetSessionByUserId(webhookEvent.ObjectId)
		if session != nil {
			sessionManager.DestroySession(*session)
		}

		slog.Info("strava_webhook > Athlete deauthorized.", "athlete_id", webhookEvent.ObjectId)
	}
}

func (svc *Strava) handleWebhookForActivity(webhookEvent *externalmodel.StravaWebhookEvent, db *sql.DB) {
	// we need to be quick here, we're just going to queue the activity for processing

	// we have a special logic for delete and update events in case there is a pending activity in the database
	switch webhookEvent.AspectType {
	case "delete", "update":
		tx, err := db.Begin()
		if err != nil {
			slog.Error("strava_webhook > Failed to begin transaction for webhook event.", "error", err)

			return
		}

		defer tx.Rollback()

		var webhookEventInDb model.StravaWebhookEvent
		found, err := webhookEventInDb.Load(webhookEvent.ObjectId, db, tx)
		if err != nil {
			slog.Error("strava_webhook > Failed to load webhook event from database.", "error", err, "event_id", webhookEvent.ObjectId)

			return
		}

		if found {
			// new event = delete
			// 	old event = create -> delete old
			//	old event = update -> update old to delete
			// new event = update
			//	old event = create -> do nothing
			// 	old event = update -> do nothing

			if webhookEvent.AspectType == "delete" {
				if webhookEventInDb.AspectType == "update" {
					webhookEventInDb.AspectType = "delete"

					err = webhookEventInDb.Save(db, tx)
					if err != nil {
						slog.Error("strava_webhook > Failed to update webhook event to delete.", "error", err, "event_id", webhookEvent.ObjectId)

						return
					}
				} else {
					err = webhookEventInDb.Delete(db, tx)
					if err != nil {
						slog.Error("strava_webhook > Failed to delete webhook event.", "error", err, "event_id", webhookEvent.ObjectId)

						return
					}
				}
			}
		} else {
			// new event = delete -> save
			// new event = update -> save

			var internalStravaWebhookEvent model.StravaWebhookEvent
			internalStravaWebhookEvent.FromExternalModel(webhookEvent)

			err = internalStravaWebhookEvent.Save(db, tx)
			if err != nil {
				slog.Error("strava_webhook > Failed to save webhook event.", "error", err, "event_id", webhookEvent.ObjectId)

				return
			}
		}

		if err = database.CommitOrRollbackSQLiteTransaction(tx); err != nil {
			slog.Error("strava_webhook > Failed to commit transaction for webhook event.", "error", err, "event_id", webhookEvent.ObjectId)

			return
		}
	case "create":
		var internalStravaWebhookEvent model.StravaWebhookEvent
		internalStravaWebhookEvent.FromExternalModel(webhookEvent)

		if err := internalStravaWebhookEvent.Save(db, nil); err != nil {
			slog.Error("strava_webhook > Failed to save webhook event.", "error", err, "event_id", webhookEvent.ObjectId)

			return
		}
	}
}
