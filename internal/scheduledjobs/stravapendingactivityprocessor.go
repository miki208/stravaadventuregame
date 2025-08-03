package scheduledjobs

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"time"

	"github.com/miki208/stravaadventuregame/internal/application"
	"github.com/miki208/stravaadventuregame/internal/database"
	"github.com/miki208/stravaadventuregame/internal/helper"
	"github.com/miki208/stravaadventuregame/internal/model"
	"github.com/miki208/stravaadventuregame/internal/service/strava"
)

func StravaPendingActivityProcessor(app *application.App) {
	slog.Info("StravaPendingActivityProcessor > StravaPendingActivityProcessor started.")

	evs, err := model.AllStravaWebhookEvents(app.SqlDb, nil, nil)
	if err != nil {
		slog.Error("StravaPendingActivityProcessor > Failed to load pending Strava webhook events.", "error", err)

		return
	}

	processingTimeUnix := time.Now().Unix()
	for _, ev := range evs {
		if ev.EventTime+int64(app.StravaSvc.GetProcessWebhookEventsAfterSec()) >= processingTimeUnix {
			continue
		}

		slog.Info("StravaPendingActivityProcessor > Processing webhook event.", "activity_id", ev.ObjectId, "event_time", ev.EventTime, "aspect_type", ev.AspectType)

		if !processOneActivity(app, &ev) {
			break // if we got a rate limit error, we stop processing
		}
	}

	slog.Info("StravaPendingActivityProcessor > StravaPendingActivityProcessor finished.")
}

type ActivityProcessingResult int

const (
	ActivityCreated ActivityProcessingResult = iota
	ActivityUpdated
	ActivityDeleted
	ActivityNotProcessed
)

func processOneActivity(app *application.App, ev *model.StravaWebhookEvent) bool {
	tx, err := app.SqlDb.Begin()
	if err != nil {
		slog.Error("StravaPendingActivityProcessor > Failed to begin main transaction.", "error", err)

		return true
	}

	defer tx.Rollback()

	processingResult := ActivityNotProcessed

	// in all cases we need this event to be deleted from the database
	if err = ev.Delete(app.SqlDb, tx); err != nil {
		slog.Error("StravaPendingActivityProcessor > Failed to delete webhook event.", "activity_id", ev.ObjectId, "error", err)

		return true
	}

	// check if the athlete exists
	var athleteExists bool
	athlete := model.NewAthlete()
	athleteExists, err = athlete.Load(ev.OwnerId, app.SqlDb, tx)
	if err != nil {
		slog.Error("StravaPendingActivityProcessor > Failed to load athlete.", "athlete_id", ev.OwnerId, "error", err)

		return true
	}

	var existingActivity model.Activity
	var newActivity *model.Activity

	if athleteExists {
		var foundOld bool
		foundOld, err = existingActivity.Load(ev.ObjectId, app.SqlDb, tx)
		if err != nil {
			slog.Error("StravaPendingActivityProcessor > Failed to load existing activity.", "activity_id", ev.ObjectId, "error", err)

			return true
		}

		if ev.AspectType == "delete" {
			// if event is delete, just delete the activity if it exists

			if foundOld {
				processingResult = ActivityDeleted

				err = existingActivity.Delete(app.SqlDb, tx)
				if err != nil {
					slog.Error("StravaPendingActivityProcessor > Failed to delete activity.", "activity_id", existingActivity.Id, "error", err)

					return true
				}
			}
		} else {
			// if event is update or create, we need to fetch the activity and check if it should be accepted/modified in db
			newActivity, err = app.StravaSvc.GetActivity(athlete.Id, ev.ObjectId, app.SqlDb, tx)
			if err != nil {
				stravaErr, ok := err.(*strava.StravaError)
				if ok && stravaErr.StatusCode() == http.StatusTooManyRequests {
					slog.Error("StravaPendingActivityProcessor > Rate limit error encountered, stopping processing.", "error", stravaErr)

					return false
				}

				slog.Error("StravaPendingActivityProcessor > Failed to fetch activity from Strava.", "activity_id", ev.ObjectId, "error", err)

				return true
			}

			shouldAcceptNew := slices.Contains(app.SupportedActivityTypes, newActivity.SportType)

			if shouldAcceptNew && foundOld && ev.AspectType == "update" {
				processingResult = ActivityUpdated

				err = newActivity.Save(app.SqlDb, tx)
			} else if shouldAcceptNew && !foundOld && ev.AspectType == "create" {
				processingResult = ActivityCreated

				err = newActivity.Save(app.SqlDb, tx)
			} else if !shouldAcceptNew && foundOld && ev.AspectType == "update" {
				processingResult = ActivityDeleted

				err = existingActivity.Delete(app.SqlDb, tx)
			}

			if err != nil {
				slog.Error("StravaPendingActivityProcessor > Failed to process activity.", "processing_result", processingResult, "activity_id", newActivity.Id, "aspect_type", ev.AspectType, "error", err)

				return true
			}
		}
	}

	err = database.CommitOrRollbackSQLiteTransaction(tx)
	if err != nil {
		slog.Error("StravaPendingActivityProcessor > Failed to commit transaction.", "error", err)

		return true
	}

	switch processingResult {
	case ActivityDeleted:
		onActivityDeleted(app, &existingActivity)
	case ActivityCreated:
		onActivityCreated(app, newActivity)
	case ActivityUpdated:
		onActivityUpdated(app, &existingActivity, newActivity)
	}

	return true
}

func onActivityDeleted(app *application.App, activity *model.Activity) {
	slog.Info("StravaPendingActivityProcessor > Activity deleted.", "activity_id", activity)

	tx, err := app.SqlDb.Begin()
	if err != nil {
		slog.Error("StravaPendingActivityProcessor > (onActivityDeleted) Failed to begin transaction.", "activity_id", activity.Id, "error", err)

		return
	}

	defer tx.Rollback()

	startedAdventure, err := model.AllAdventures(app.SqlDb, tx, map[string]any{
		"athlete_id": activity.AthleteId,
		"completed":  0,
	})
	if err != nil {
		slog.Error("StravaPendingActivityProcessor > (onActivityDeleted) Failed to load started adventures.", "athlete_id", activity.AthleteId, "error", err)

		return
	}

	progressIsMade := false

	var oldTotalDistance float32
	if len(startedAdventure) > 0 && startedAdventure[0].StartDate < activity.StartDate {
		oldTotalDistance = startedAdventure[0].CurrentDistance

		if startedAdventure[0].CurrentDistance > activity.Distance {
			startedAdventure[0].CurrentDistance -= activity.Distance
		} else {
			startedAdventure[0].CurrentDistance = 0
		}

		if startedAdventure[0].CurrentDistance != oldTotalDistance {
			progressIsMade = true

			err = onTotalDistanceUpdated(&startedAdventure[0], activity, oldTotalDistance, app, tx)
			if err != nil {
				slog.Error("StravaPendingActivityProcessor > (onActivityDeleted) Failed to update state on total distance updated.", "error", err)

				return
			}
		}
	}

	if err = database.CommitOrRollbackSQLiteTransaction(tx); err != nil {
		slog.Error("StravaPendingActivityProcessor > (onActivityDeleted) Failed to commit transaction with updated progress.", "error", err)

		return
	}

	if progressIsMade {
		if err = onProgressCommited(&startedAdventure[0], activity, app, "delete"); err != nil {
			slog.Error("StravaPendingActivityProcessor > (onActivityDeleted) Error occurred on trigger for commited progress.", "error", err)

			return
		}
	}
}

func onActivityCreated(app *application.App, activity *model.Activity) {
	slog.Info("StravaPendingActivityProcessor > Activity created.", "activity_id", activity.Id)

	tx, err := app.SqlDb.Begin()
	if err != nil {
		slog.Error("StravaPendingActivityProcessor > (onActivityCreated) Failed to begin transaction.", "activity_id", activity.Id, "error", err)

		return
	}

	defer tx.Rollback()

	// check if there is any started adventure
	startedAdventure, err := model.AllAdventures(app.SqlDb, tx, map[string]any{
		"athlete_id": activity.AthleteId,
		"completed":  0,
	})
	if err != nil {
		slog.Error("StravaPendingActivityProcessor > (onActivityCreated) Failed to load started adventures.", "athlete_id", activity.AthleteId, "error", err)

		return
	}

	progressIsMade := false

	var oldTotalDistance float32
	if len(startedAdventure) > 0 && startedAdventure[0].StartDate <= activity.StartDate {
		oldTotalDistance = startedAdventure[0].CurrentDistance

		// if there is an adventure that started before this activity, we can add the activity's distance to it
		startedAdventure[0].CurrentDistance += activity.Distance

		if oldTotalDistance != startedAdventure[0].CurrentDistance {
			progressIsMade = true

			err = onTotalDistanceUpdated(&startedAdventure[0], activity, oldTotalDistance, app, tx)
			if err != nil {
				slog.Error("StravaPendingActivityProcessor > (onActivityCreated) Failed to update state on total distance updated.", "error", err)

				return
			}
		}
	}

	if err = database.CommitOrRollbackSQLiteTransaction(tx); err != nil {
		slog.Error("StravaPendingActivityProcessor > (onActivityCreated) Failed to commit transaction with updated progress.", "error", err)

		return
	}

	if progressIsMade {
		if err = onProgressCommited(&startedAdventure[0], activity, app, "create"); err != nil {
			slog.Error("StravaPendingActivityProcessor > (onActivityCreated) Error occurred on trigger for commited progress.", "error", err)

			return
		}
	}
}

func onActivityUpdated(app *application.App, oldActivity *model.Activity, newActivity *model.Activity) {
	slog.Info("StravaPendingActivityProcessor > Activity updated.", "activity_id", newActivity.Id)

	tx, err := app.SqlDb.Begin()
	if err != nil {
		slog.Error("StravaPendingActivityProcessor > (onActivityUpdated) Failed to begin transaction.", "activity_id", newActivity.Id, "error", err)

		return
	}

	defer tx.Rollback()

	// check if there is any started adventure
	startedAdventure, err := model.AllAdventures(app.SqlDb, tx, map[string]any{
		"athlete_id": newActivity.AthleteId,
		"completed":  0,
	})
	if err != nil {
		slog.Error("StravaPendingActivityProcessor > (onActivityUpdated) Failed to load started adventures.", "athlete_id", newActivity.AthleteId, "error", err)

		return
	}

	progressIsMade := false

	var oldTotalDistance float32
	if len(startedAdventure) > 0 {
		oldTotalDistance = startedAdventure[0].CurrentDistance

		var distanceToAdd float32
		if startedAdventure[0].StartDate <= oldActivity.StartDate && startedAdventure[0].StartDate > newActivity.StartDate {
			distanceToAdd = -oldActivity.Distance
		} else if startedAdventure[0].StartDate > oldActivity.StartDate && startedAdventure[0].StartDate <= newActivity.StartDate {
			distanceToAdd = newActivity.Distance
		} else if startedAdventure[0].StartDate <= oldActivity.StartDate && startedAdventure[0].StartDate <= newActivity.StartDate {
			distanceToAdd = newActivity.Distance - oldActivity.Distance
		}

		startedAdventure[0].CurrentDistance += distanceToAdd
		if startedAdventure[0].CurrentDistance < 0 {
			startedAdventure[0].CurrentDistance = 0
		}

		if oldTotalDistance != startedAdventure[0].CurrentDistance {
			progressIsMade = true

			err = onTotalDistanceUpdated(&startedAdventure[0], newActivity, oldTotalDistance, app, tx)
			if err != nil {
				slog.Error("StravaPendingActivityProcessor > (onActivityUpdated) Failed to update state on total distance updated.", "error", err)

				return
			}
		}
	}

	if err = database.CommitOrRollbackSQLiteTransaction(tx); err != nil {
		slog.Error("StravaPendingActivityProcessor > (onActivityUpdated) Failed to commit transaction with updated progress.", "error", err)

		return
	}

	if progressIsMade {
		if err = onProgressCommited(&startedAdventure[0], newActivity, app, "update"); err != nil {
			slog.Error("StravaPendingActivityProcessor > (onActivityUpdated) Error occurred on trigger for commited progress.", "error", err)

			return
		}
	}
}

func onTotalDistanceUpdated(adventure *model.Adventure, activity *model.Activity, oldTotalDistance float32, app *application.App, tx *sql.Tx) error {
	if adventure.CurrentDistance >= adventure.TotalDistance {
		adventure.Completed = 1
		adventure.CurrentDistance = adventure.TotalDistance

		onAdventureCompleted(adventure, activity, app, tx)
	} else {
		if adventure.CurrentDistance == 0 {
			var startLocation model.Location
			found, err := startLocation.Load(adventure.StartLocation, app.SqlDb, tx)
			if err != nil {
				return err
			}

			if !found {
				return errors.New("start location not found")
			}

			adventure.CurrentLocationName = startLocation.Name
		} else {
			courseDbName := fmt.Sprintf("%d-%d", min(adventure.StartLocation, adventure.EndLocation),
				max(adventure.StartLocation, adventure.EndLocation))

			var route *model.DirectionsRoute = model.NewDirectionsRoute()
			err := app.FileDb.Read("course", courseDbName, route)
			if err != nil {
				return err
			}

			lon, lat, err := helper.GetPointFromPolylineAndDistance(route.Geometry, adventure.StartLocation > adventure.EndLocation, adventure.CurrentDistance*1000)
			if err != nil {
				return err
			}

			geocodeResults, err := app.OrsSvc.ReverseGeocode(lon, lat, 10, "country,region,locality,localadmin")
			if err != nil {
				return err
			}

			adventure.CurrentLocationName = helper.GetPreferedLocationName(geocodeResults)
		}
	}

	return adventure.Save(app.SqlDb, tx)
}

func onAdventureCompleted(adventure *model.Adventure, activity *model.Activity, app *application.App, tx *sql.Tx) error {
	adventure.EndDate = activity.StartDate + activity.MovingTime

	var endLocation model.Location
	found, err := endLocation.Load(adventure.EndLocation, app.SqlDb, tx)
	if err != nil {
		return err
	}

	if !found {
		return errors.New("end location not found")
	}

	adventure.CurrentLocationName = endLocation.Name

	return nil
}

func onProgressCommited(adventure *model.Adventure, activity *model.Activity, app *application.App, eventType string) error {
	// for now, just update the activity description with the adventure progress (if enabled)

	var athleteSettings model.AthleteSettings
	found, err := athleteSettings.Load(adventure.AthleteId, app.SqlDb, nil)
	if err != nil {
		return err
	}

	if !found {
		return errors.New("settings not found")
	}

	if athleteSettings.AutoUpdateActivityDescription == 0 {
		return nil
	}

	if eventType != "create" {
		return nil
	}

	var locationStart, locationEnd model.Location
	foundStart, err := locationStart.Load(adventure.StartLocation, app.SqlDb, nil)
	if err != nil {
		return err
	}

	foundEnd, err := locationEnd.Load(adventure.EndLocation, app.SqlDb, nil)
	if err != nil {
		return err
	}

	if !foundStart || !foundEnd {
		return errors.New("start or end location not found")
	}

	var descriptionText string
	if adventure.Completed == 1 {
		descriptionText = fmt.Sprintf("Adventure completed!\nI have reached %s (started from %s, at %s (GMT)).\nTotal distance: %.2f km.",
			locationEnd.Name, locationStart.Name, time.Unix(int64(adventure.StartDate), 0).UTC().Format(time.DateTime), adventure.TotalDistance)
	} else {
		descriptionText = fmt.Sprintf("Adventure in progress!\nI am at %s (started from %s, at %s (GMT), going to %s).\nDistance traveled: %.2f/%.2f km.",
			adventure.CurrentLocationName, locationStart.Name, time.Unix(int64(adventure.StartDate), 0).UTC().Format(time.DateTime), locationEnd.Name,
			adventure.CurrentDistance, adventure.TotalDistance)
	}

	var fullDescription string
	if activity.Description != "" {
		fullDescription = activity.Description + "\n\n" + descriptionText
	} else {
		fullDescription = descriptionText
	}

	activity, err = app.StravaSvc.UpdateActivity(activity.AthleteId, activity.Id, map[string]any{
		"description": fullDescription,
	}, app.SqlDb, nil)

	if err != nil {
		return err
	}

	err = activity.Save(app.SqlDb, nil)
	if err != nil {
		return err
	}

	return nil
}
