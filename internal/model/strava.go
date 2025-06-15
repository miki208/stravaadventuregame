package model

import (
	"database/sql"
	"errors"
	"time"

	"github.com/miki208/stravaadventuregame/internal/service/strava/externalmodel"
)

type Athlete struct {
	*externalmodel.Athlete
	isAdmin int
}

func (athlete *Athlete) FromExternalModel(externalAthlete *externalmodel.Athlete) {
	athlete.Athlete = externalAthlete
}

func NewAthlete() *Athlete {
	return &Athlete{Athlete: &externalmodel.Athlete{}}
}

func (athl *Athlete) Load(id int64, db *sql.DB, tx *sql.Tx) (bool, error) {
	var err error

	query, params := PrepareQuery("SELECT * FROM Athlete", map[string]any{"id": id})

	var row *sql.Row
	if tx != nil {
		row = tx.QueryRow(query, params...)
	} else {
		row = db.QueryRow(query, params...)
	}

	err = row.Scan(&athl.Id, &athl.FirstName, &athl.LastName, &athl.City, &athl.Country, &athl.Sex, &athl.isAdmin)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = nil
		}

		return false, err
	}

	return true, nil
}

func (athl *Athlete) Save(db *sql.DB, tx *sql.Tx) error {
	var err error

	var found bool
	found, err = AthleteExists(athl.Id, db, tx)
	if err != nil {
		return err
	}

	if found {
		query := "UPDATE Athlete SET first_name=?, last_name=?, city=?, country=?, sex=?, is_admin=? WHERE id=?"

		if tx != nil {
			_, err = tx.Exec(query, athl.FirstName, athl.LastName, athl.City, athl.Country, athl.Sex, athl.isAdmin, athl.Id)
		} else {
			_, err = db.Exec(query, athl.FirstName, athl.LastName, athl.City, athl.Country, athl.Sex, athl.isAdmin, athl.Id)
		}
	} else {
		query := "INSERT INTO Athlete VALUES(?, ?, ?, ?, ?, ?, ?)"

		if tx != nil {
			_, err = tx.Exec(query, athl.Id, athl.FirstName, athl.LastName, athl.City, athl.Country, athl.Sex, athl.isAdmin)
		} else {
			_, err = db.Exec(query, athl.Id, athl.FirstName, athl.LastName, athl.City, athl.Country, athl.Sex, athl.isAdmin)
		}
	}

	return err
}

func (athl *Athlete) Delete(db *sql.DB, tx *sql.Tx) error {
	var err error

	query, params := PrepareQuery("DELETE FROM Athlete", map[string]any{"id": athl.Id})

	if tx != nil {
		_, err = tx.Exec(query, params...)
	} else {
		_, err = db.Exec(query, params...)
	}

	return err
}

func (athl *Athlete) IsAdmin() bool {
	return athl.isAdmin > 0
}

func (athl *Athlete) SetIsAdmin(isAdmin bool) {
	if isAdmin {
		athl.isAdmin = 1
	} else {
		athl.isAdmin = 0
	}
}

func AthleteExists(id int64, db *sql.DB, tx *sql.Tx) (bool, error) {
	temp := NewAthlete()

	return temp.Load(id, db, tx)
}

type StravaCredential struct {
	AthleteId    int64
	AccessToken  string
	RefreshToken string
	ExpiresAt    int
}

func (stravaCredential *StravaCredential) FromExternalModel(externalTokenExchangeResponse *externalmodel.TokenExchangeResponse) {
	stravaCredential.AthleteId = externalTokenExchangeResponse.Athl.Id
	stravaCredential.AccessToken = externalTokenExchangeResponse.AccessToken
	stravaCredential.RefreshToken = externalTokenExchangeResponse.RefreshToken
	stravaCredential.ExpiresAt = externalTokenExchangeResponse.ExpiresAt
}

func (cred *StravaCredential) Load(athlId int64, db *sql.DB, tx *sql.Tx) (bool, error) {
	var err error

	query, params := PrepareQuery("SELECT * FROM StravaCredential", map[string]any{"athlete_id": athlId})

	var row *sql.Row
	if tx != nil {
		row = tx.QueryRow(query, params...)
	} else {
		row = db.QueryRow(query, params...)
	}

	err = row.Scan(&cred.AthleteId, &cred.AccessToken, &cred.RefreshToken, &cred.ExpiresAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = nil
		}

		return false, err
	}

	return true, nil
}

func (cred *StravaCredential) Save(db *sql.DB, tx *sql.Tx) error {
	var err error

	var found bool
	found, err = StravaCredentialExists(cred.AthleteId, db, tx)
	if err != nil {
		return err
	}

	if found {
		query := "UPDATE StravaCredential SET access_token=?, refresh_token=?, expires_at=? WHERE athlete_id=?"

		if tx != nil {
			_, err = tx.Exec(query, cred.AccessToken, cred.RefreshToken, cred.ExpiresAt, cred.AthleteId)
		} else {
			_, err = db.Exec(query, cred.AccessToken, cred.RefreshToken, cred.ExpiresAt, cred.AthleteId)
		}
	} else {
		query := "INSERT INTO StravaCredential VALUES(?, ?, ?, ?)"

		if tx != nil {
			_, err = tx.Exec(query, cred.AthleteId, cred.AccessToken, cred.RefreshToken, cred.ExpiresAt)
		} else {
			_, err = db.Exec(query, cred.AthleteId, cred.AccessToken, cred.RefreshToken, cred.ExpiresAt)
		}
	}

	return err
}

func StravaCredentialExists(athlId int64, db *sql.DB, tx *sql.Tx) (bool, error) {
	var temp StravaCredential

	return temp.Load(athlId, db, tx)
}

type StravaWebhookSubscription struct {
	Id int `json:"id"`
}

type StravaWebhookEvent struct {
	ObjectId   int64
	OwnerId    int64
	AspectType string
}

func (stravaWebhookEvent *StravaWebhookEvent) FromExternalModel(externalStravaWebhookEvent *externalmodel.StravaWebhookEvent) {
	stravaWebhookEvent.ObjectId = externalStravaWebhookEvent.ObjectId
	stravaWebhookEvent.OwnerId = externalStravaWebhookEvent.OwnerId
	stravaWebhookEvent.AspectType = externalStravaWebhookEvent.AspectType
}

func (ev *StravaWebhookEvent) Load(activityId int64, db *sql.DB, tx *sql.Tx) (bool, error) {
	var err error

	query, params := PrepareQuery("SELECT * FROM PendingActivity", map[string]any{"id": activityId})

	var row *sql.Row
	if tx != nil {
		row = tx.QueryRow(query, params...)
	} else {
		row = db.QueryRow(query, params...)
	}

	err = row.Scan(&ev.ObjectId, &ev.OwnerId, &ev.AspectType)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = nil
		}

		return false, err
	}

	return true, nil
}

func (ev *StravaWebhookEvent) Save(db *sql.DB, tx *sql.Tx) error {
	var err error

	var found bool
	found, err = StravaWebhookEventExists(ev.ObjectId, db, tx)
	if err != nil {
		return err
	}

	if found {
		query := "UPDATE PendingActivity SET aspect_type=? WHERE id=?"

		if tx != nil {
			_, err = tx.Exec(query, ev.AspectType, ev.ObjectId)
		} else {
			_, err = db.Exec(query, ev.AspectType, ev.ObjectId)
		}
	} else {
		query := "INSERT INTO PendingActivity VALUES(?, ?, ?)"

		if tx != nil {
			_, err = tx.Exec(query, ev.ObjectId, ev.OwnerId, ev.AspectType)
		} else {
			_, err = db.Exec(query, ev.ObjectId, ev.OwnerId, ev.AspectType)
		}
	}

	return err
}

func (ev *StravaWebhookEvent) Delete(db *sql.DB, tx *sql.Tx) error {
	var err error

	query, params := PrepareQuery("DELETE FROM PendingActivity", map[string]any{
		"id": ev.ObjectId,
	})

	if tx != nil {
		_, err = tx.Exec(query, params...)
	} else {
		_, err = db.Exec(query, params...)
	}

	return err
}

func StravaWebhookEventExists(activityId int64, db *sql.DB, tx *sql.Tx) (bool, error) {
	var temp StravaWebhookEvent

	return temp.Load(activityId, db, tx)
}

func AllStravaWebhookEvents(db *sql.DB, tx *sql.Tx, filter map[string]any) ([]StravaWebhookEvent, error) {
	var err error

	var rows *sql.Rows
	query, params := PrepareQuery("SELECT * FROM PendingActivity", filter)
	if tx != nil {
		rows, err = tx.Query(query, params...)
	} else {
		rows, err = db.Query(query, params...)
	}
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var events []StravaWebhookEvent
	for rows.Next() {
		events = append(events, StravaWebhookEvent{})

		eventToEdit := &events[len(events)-1]
		if err = rows.Scan(&eventToEdit.ObjectId, &eventToEdit.OwnerId, &eventToEdit.AspectType); err != nil {
			return nil, err
		}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

type Activity struct {
	Id                 int64
	Distance           float32
	MovingTime         int
	TotalElevationGain float32
	SportType          string
	StartDate          int

	AthleteId int64
}

func (internalActivity *Activity) FromExternalModel(externalActivity *externalmodel.Activity) {
	internalActivity.Id = externalActivity.Id
	internalActivity.Distance = externalActivity.Distance
	internalActivity.MovingTime = externalActivity.MovingTime
	internalActivity.TotalElevationGain = externalActivity.TotalElevationGain
	internalActivity.SportType = externalActivity.SportType

	parsedTime, err := time.Parse(time.RFC3339, externalActivity.StartDate)
	if err != nil {
		internalActivity.StartDate = 0
	} else {
		internalActivity.StartDate = int(parsedTime.Unix())
	}
}

func (a *Activity) Load(id int64, db *sql.DB, tx *sql.Tx) (bool, error) {
	var err error

	query, params := PrepareQuery("SELECT id, athlete_id, type, distance, start_date, moving_time, elevation_gain FROM Activity", map[string]any{
		"id": id,
	})

	var row *sql.Row
	if tx != nil {
		row = tx.QueryRow(query, params...)
	} else {
		row = db.QueryRow(query, params...)
	}

	err = row.Scan(&a.Id, &a.AthleteId, &a.SportType, &a.Distance, &a.StartDate, &a.MovingTime, &a.TotalElevationGain)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = nil
		}

		return false, err
	}

	return true, nil
}

func (a *Activity) Save(db *sql.DB, tx *sql.Tx) error {
	var err error

	var found bool
	found, err = ActivityExists(a.Id, db, tx)
	if err != nil {
		return err
	}

	if found {
		query := "UPDATE Activity SET type=?, distance=?, start_date=?, moving_time=?, elevation_gain=? WHERE id=?"

		if tx != nil {
			_, err = tx.Exec(query, a.SportType, a.Distance, a.StartDate, a.MovingTime, a.TotalElevationGain, a.Id)
		} else {
			_, err = db.Exec(query, a.SportType, a.Distance, a.StartDate, a.MovingTime, a.TotalElevationGain, a.Id)
		}
	} else {
		query := "INSERT INTO Activity VALUES(?, ?, ?, ?, ?, ?, ?)"

		if tx != nil {
			_, err = tx.Exec(query, a.Id, a.AthleteId, a.SportType, a.Distance, a.StartDate, a.MovingTime, a.TotalElevationGain)
		} else {
			_, err = db.Exec(query, a.Id, a.AthleteId, a.SportType, a.Distance, a.StartDate, a.MovingTime, a.TotalElevationGain)
		}
	}

	return err
}

func (a *Activity) Delete(db *sql.DB, tx *sql.Tx) error {
	var err error

	query, params := PrepareQuery("DELETE FROM Activity", map[string]any{
		"id": a.Id,
	})

	if tx != nil {
		_, err = tx.Exec(query, params...)
	} else {
		_, err = db.Exec(query, params...)
	}

	return err
}

func ActivityExists(id int64, db *sql.DB, tx *sql.Tx) (bool, error) {
	var temp Activity

	return temp.Load(id, db, tx)
}

func AllActivities(db *sql.DB, tx *sql.Tx, filter map[string]any) ([]Activity, error) {
	var err error

	var rows *sql.Rows
	query, params := PrepareQuery("SELECT id, athlete_id, type, distance, start_date, moving_time, elevation_gain FROM Activity", filter)
	if tx != nil {
		rows, err = tx.Query(query, params...)
	} else {
		rows, err = db.Query(query, params...)
	}
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var activities []Activity
	for rows.Next() {
		activities = append(activities, Activity{})

		activityToEdit := &activities[len(activities)-1]
		if err = rows.Scan(&activityToEdit.Id, &activityToEdit.AthleteId, &activityToEdit.SportType, &activityToEdit.Distance,
			&activityToEdit.StartDate, &activityToEdit.MovingTime, &activityToEdit.TotalElevationGain); err != nil {
			return nil, err
		}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return activities, nil
}
