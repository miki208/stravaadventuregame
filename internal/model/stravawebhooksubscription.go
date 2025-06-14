package model

import (
	"database/sql"
	"errors"

	"github.com/miki208/stravaadventuregame/internal/database"
)

type StravaWebhookSubscription struct {
	Id int `json:"id"`
}

type StravaWebhookEvent struct {
	ObjectType     string            `json:"object_type"`
	ObjectId       int64             `json:"object_id"`
	AspectType     string            `json:"aspect_type"`
	Updates        map[string]string `json:"updates"`
	OwnerId        int64             `json:"owner_id"`
	SubscriptionId int               `json:"subscription_id"`
	EventTime      int64             `json:"event_time"`
}

func (ev *StravaWebhookEvent) GetSummary() *StravaWebhookEventSummary {
	return &StravaWebhookEventSummary{
		ObjectId:   ev.ObjectId,
		OwnerId:    ev.OwnerId,
		AspectType: ev.AspectType,
	}
}

type StravaWebhookEventSummary struct {
	ObjectId   int64
	OwnerId    int64
	AspectType string
}

func (ev *StravaWebhookEventSummary) LoadByActivityId(activityId int64, db *sql.DB, tx *sql.Tx) (bool, error) {
	var row *sql.Row
	query := "SELECT * FROM PendingActivity WHERE id=?"

	if tx != nil {
		row = tx.QueryRow(query, activityId)
	} else {
		row = db.QueryRow(query, activityId)
	}

	err := row.Scan(&ev.ObjectId, &ev.OwnerId, &ev.AspectType)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		if tx != nil {
			tx.Rollback()
		}

		return false, err
	}

	return true, nil
}

func EventExists(activityId int64, db *sql.DB, tx *sql.Tx) (bool, error) {
	var temp StravaWebhookEventSummary

	exists, err := temp.LoadByActivityId(activityId, db, tx)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (ev *StravaWebhookEventSummary) Save(db *sql.DB, tx *sql.Tx) error {
	isExternalTx, tx, err := database.GetOrCreateSQLiteTransaction(db, tx)
	if err != nil {
		return err
	}

	exists, err := EventExists(ev.ObjectId, db, tx)
	if err != nil {
		return err
	}

	if exists {
		_, err = tx.Exec("UPDATE PendingActivity SET aspect_type=? WHERE id=?", ev.AspectType, ev.ObjectId)
		if err != nil {
			tx.Rollback()

			return err
		}
	} else {
		_, err = tx.Exec("INSERT INTO PendingActivity VALUES(?, ?, ?)", ev.ObjectId, ev.OwnerId, ev.AspectType)
		if err != nil {
			tx.Rollback()

			return err
		}
	}

	if !isExternalTx {
		err = database.CommitOrRollbackSQLiteTransaction(tx)
		if err != nil {
			return err
		}
	}

	return nil
}
