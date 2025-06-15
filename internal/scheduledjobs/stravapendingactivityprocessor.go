package scheduledjobs

import (
	"net/http"
	"slices"

	"github.com/miki208/stravaadventuregame/internal/application"
	"github.com/miki208/stravaadventuregame/internal/database"
	"github.com/miki208/stravaadventuregame/internal/model"
	"github.com/miki208/stravaadventuregame/internal/service/strava"
)

func StravaPendingActivityProcessor(app *application.App) {
	evs, err := model.AllStravaWebhookEvents(app.SqlDb, nil, nil)
	if err != nil {
		return
	}

	for _, ev := range evs {
		if !processOneActivity(app, &ev) {
			break // if we got a rate limit error, we stop processing
		}
	}
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
		return true
	}

	defer tx.Rollback()

	processingResult := ActivityNotProcessed

	// in all cases we need this event to be deleted from the database
	if err = ev.Delete(app.SqlDb, tx); err != nil {
		return true
	}

	// check if the athlete exists
	var athleteExists bool
	athlete := model.NewAthlete()
	athleteExists, err = athlete.Load(ev.OwnerId, app.SqlDb, tx)
	if err != nil {
		return true
	}

	var existingActivity model.Activity
	var newActivity *model.Activity

	if athleteExists {
		var foundOld bool
		foundOld, err = existingActivity.Load(ev.ObjectId, app.SqlDb, tx)
		if err != nil {
			return true
		}

		if ev.AspectType == "delete" {
			// if event is delete, just delete the activity if it exists

			if foundOld {
				processingResult = ActivityDeleted

				err = existingActivity.Delete(app.SqlDb, tx)
				if err != nil {
					return true
				}
			}
		} else {
			// if event is update or create, we need to fetch the activity and check if it should be accepted/modified in db
			newActivity, err = app.StravaSvc.GetActivity(athlete.Id, ev.ObjectId, app.SqlDb, tx)
			if err != nil {
				stravaErr, ok := err.(*strava.StravaError)
				if ok && stravaErr.StatusCode() == http.StatusTooManyRequests {
					return false
				}

				return true
			}

			allowedSportTypes := []string{"Hike", "Run", "TrailRun", "VirtualRun", "Walk", "Wheelchair"}
			shouldAcceptNew := slices.Contains(allowedSportTypes, newActivity.SportType)

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
				return true
			}
		}
	}

	err = database.CommitOrRollbackSQLiteTransaction(tx)
	if err != nil {
		return true
	}

	if processingResult == ActivityDeleted {
		onActivityDeleted(app, &existingActivity)
	} else if processingResult == ActivityCreated {
		onActivityCreated(app, newActivity)
	} else if processingResult == ActivityUpdated {
		onActivityUpdated(app, &existingActivity, newActivity)
	}

	return true
}

func onActivityDeleted(app *application.App, activity *model.Activity) {
	tx, err := app.SqlDb.Begin()
	if err != nil {
		return
	}

	defer tx.Rollback()

	startedAdventure, err := model.AllAdventures(app.SqlDb, tx, map[string]any{
		"athlete_id": activity.AthleteId,
		"completed":  0,
	})
	if err != nil {
		return
	}

	if len(startedAdventure) > 0 && startedAdventure[0].StartDate < activity.StartDate {
		if startedAdventure[0].CurrentDistance > activity.Distance {
			startedAdventure[0].CurrentDistance -= activity.Distance
		} else {
			startedAdventure[0].CurrentDistance = 0
		}

		err = startedAdventure[0].Save(app.SqlDb, tx)
		if err != nil {
			return
		}
	}

	if err = database.CommitOrRollbackSQLiteTransaction(tx); err != nil {
	}
}

func onActivityCreated(app *application.App, activity *model.Activity) {
	tx, err := app.SqlDb.Begin()
	if err != nil {
		return
	}

	defer tx.Rollback()

	// check if there is any started adventure
	startedAdventure, err := model.AllAdventures(app.SqlDb, tx, map[string]any{
		"athlete_id": activity.AthleteId,
		"completed":  0,
	})
	if err != nil {
		return
	}

	if len(startedAdventure) > 0 && startedAdventure[0].StartDate <= activity.StartDate {
		// if there is an adventure that started before this activity, we can add the activity's distance to it
		startedAdventure[0].CurrentDistance += activity.Distance

		if startedAdventure[0].CurrentDistance >= startedAdventure[0].TotalDistance {
			startedAdventure[0].Completed = 1
			startedAdventure[0].EndDate = activity.StartDate + activity.MovingTime
		}

		err = startedAdventure[0].Save(app.SqlDb, tx)
		if err != nil {
			return
		}
	}

	if err = database.CommitOrRollbackSQLiteTransaction(tx); err != nil {
		// update strava activity description
	}
}

func onActivityUpdated(app *application.App, oldActivity *model.Activity, newActivity *model.Activity) {
	tx, err := app.SqlDb.Begin()
	if err != nil {
		return
	}

	defer tx.Rollback()

	// check if there is any started adventure
	startedAdventure, err := model.AllAdventures(app.SqlDb, tx, map[string]any{
		"athlete_id": newActivity.AthleteId,
		"completed":  0,
	})
	if err != nil {
		return
	}

	if len(startedAdventure) > 0 {
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
		} else if startedAdventure[0].CurrentDistance >= startedAdventure[0].TotalDistance {
			startedAdventure[0].Completed = 1
			startedAdventure[0].EndDate = newActivity.StartDate + newActivity.MovingTime
		}

		err = startedAdventure[0].Save(app.SqlDb, tx)
		if err != nil {
			return
		}
	}

	if err = database.CommitOrRollbackSQLiteTransaction(tx); err != nil {
		// update strava activity description
	}
}
