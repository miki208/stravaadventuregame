package strava

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"

	"github.com/miki208/stravaadventuregame/internal/database"
	"github.com/miki208/stravaadventuregame/internal/helper"
	"github.com/miki208/stravaadventuregame/internal/model"
	"github.com/miki208/stravaadventuregame/internal/service/strava/externalmodel"
)

func (svc *Strava) StravaWebhookCallback(resp http.ResponseWriter, req *http.Request, db *sql.DB, sessionManager *helper.SessionManager) {
	if req.Method == http.MethodGet {
		svc.handleWebhookForSubscriptionValidation(resp, req, db)
	} else if req.Method == http.MethodPost {
		requestBody, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(resp, "Failed to read request body", http.StatusInternalServerError)

			return
		}

		var webhookEvent externalmodel.StravaWebhookEvent
		err = json.Unmarshal(requestBody, &webhookEvent)
		if err != nil {
			http.Error(resp, "Failed to unmarshal webhook event", http.StatusInternalServerError)

			return
		}

		if webhookEvent.ObjectType == "athlete" {
			svc.handleWebhookForAthlete(&webhookEvent, db, sessionManager)
		} else if webhookEvent.ObjectType == "activity" {
			svc.handleWebhookForActivity(&webhookEvent, db)
		}

		resp.WriteHeader(http.StatusOK)
	}
}

func (svc *Strava) handleWebhookForSubscriptionValidation(resp http.ResponseWriter, req *http.Request, db *sql.DB) {
	query := req.URL.Query()

	if !query.Has("hub.challenge") || !query.Has("hub.mode") || !query.Has("hub.verify_token") {
		http.Error(resp, "Missing required query parameters", http.StatusBadRequest)

		return
	}

	if query.Get("hub.mode") != "subscribe" {
		http.Error(resp, "Invalid hub.mode", http.StatusBadRequest)

		return
	}

	if query.Get("hub.verify_token") != svc.GetVerifyToken() {
		http.Error(resp, "Invalid hub.verify_token", http.StatusBadRequest)

		return
	}

	err := SendCallbackValidationResponse(query.Get("hub.challenge"), resp)
	if err != nil {
		http.Error(resp, "Failed to send callback validation response", http.StatusInternalServerError)

		return
	}
}

func (svc *Strava) handleWebhookForAthlete(webhookEvent *externalmodel.StravaWebhookEvent, db *sql.DB, sessionManager *helper.SessionManager) {
	authorizedUpdate, ok := webhookEvent.Updates["authorized"]
	if ok && authorizedUpdate == "false" {
		// atlete is revoking access, we're going to delete the athlete from the database
		svc.Deauthorize(webhookEvent.ObjectId, false, db, nil)

		session := sessionManager.GetSessionByUserId(webhookEvent.ObjectId)
		if session != nil {
			sessionManager.DestroySession(*session)
		}
	}
}

func (svc *Strava) handleWebhookForActivity(webhookEvent *externalmodel.StravaWebhookEvent, db *sql.DB) {
	// we need to be quick here, we're just going to queue the activity for processing

	// we have a special logic for delete and update events in case there is a pending activity in the database
	if webhookEvent.AspectType == "delete" || webhookEvent.AspectType == "update" {
		tx, err := db.Begin()
		if err != nil {
			return
		}

		defer tx.Rollback()

		var webhookEventInDb model.StravaWebhookEvent
		found, err := webhookEventInDb.Load(webhookEvent.ObjectId, db, tx)
		if err != nil {
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
						return
					}
				} else {
					err = webhookEventInDb.Delete(db, tx)
					if err != nil {
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
				return
			}
		}

		database.CommitOrRollbackSQLiteTransaction(tx)
	} else if webhookEvent.AspectType == "create" {
		var internalStravaWebhookEvent model.StravaWebhookEvent
		internalStravaWebhookEvent.FromExternalModel(webhookEvent)

		internalStravaWebhookEvent.Save(db, nil)
	}
}
