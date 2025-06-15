package scheduledjobs

import (
	"fmt"
	"slices"

	"github.com/miki208/stravaadventuregame/internal/application"
	"github.com/miki208/stravaadventuregame/internal/database"
	"github.com/miki208/stravaadventuregame/internal/model"
)

func StravaPendingActivityProcessor(app *application.App) {
	evs, err := model.AllStravaWebhookEvents(app.SqlDb, nil, nil)
	if err != nil {
		return
	}

	for _, ev := range evs {
		processOneActivity(app, &ev)
	}
}

type ActivityProcessingResult int

const (
	ActivityCreated ActivityProcessingResult = iota
	ActivityUpdated
	ActivityDeleted
	ActivityNotProcessed
)

func processOneActivity(app *application.App, ev *model.StravaWebhookEvent) {
	tx, err := app.SqlDb.Begin()
	if err != nil {
		return
	}

	defer tx.Rollback()

	processingResult := ActivityNotProcessed

	// in all cases we need this event to be deleted from the database
	if err = ev.Delete(app.SqlDb, tx); err != nil {
		return
	}

	// check if the athlete exists
	var athleteExists bool
	athlete := model.NewAthlete()
	athleteExists, err = athlete.Load(ev.OwnerId, app.SqlDb, tx)
	if err != nil {
		return
	}

	var existingActivity model.Activity
	var newActivity *model.Activity

	if athleteExists {
		var foundOld bool
		foundOld, err = existingActivity.Load(ev.ObjectId, app.SqlDb, tx)
		if err != nil {
			return
		}

		if ev.AspectType == "delete" {
			// if event is delete, just delete the activity if it exists

			if foundOld {
				processingResult = ActivityDeleted

				err = existingActivity.Delete(app.SqlDb, tx)
				if err != nil {
					return
				}
			}
		} else {
			// if event is update or create, we need to fetch the activity and check if it should be accepted/modified in db
			newActivity, err = app.StravaSvc.GetActivity(athlete.Id, ev.ObjectId, app.SqlDb, tx)
			if err != nil {
				return
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
				return
			}
		}
	}

	err = database.CommitOrRollbackSQLiteTransaction(tx)
	if err != nil {
		return
	}

	if processingResult == ActivityDeleted {
		onActivityDeleted(app, &existingActivity)
	} else if processingResult == ActivityCreated {
		onActivityCreated(app, newActivity)
	} else if processingResult == ActivityUpdated {
		onActivityUpdated(app, &existingActivity, newActivity)
	}
}

func onActivityDeleted(app *application.App, activity *model.Activity) {
	fmt.Printf("Activity %d deleted\n", activity.Id)
}

func onActivityCreated(app *application.App, activity *model.Activity) {
	fmt.Printf("Activity %d created (distance %f)\n", activity.Id, activity.Distance)
}

func onActivityUpdated(app *application.App, oldActivity *model.Activity, newActivity *model.Activity) {
	fmt.Printf("Activity %d updated (distance %f -> %f)\n", oldActivity.Id, oldActivity.Distance, newActivity.Distance)
}
