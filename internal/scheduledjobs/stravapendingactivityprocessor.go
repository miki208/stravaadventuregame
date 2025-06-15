package scheduledjobs

import (
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
		tx, err := app.SqlDb.Begin()
		if err != nil {
			continue
		}

		var exists bool
		athlete := model.NewAthlete()
		exists, err = athlete.Load(ev.OwnerId, app.SqlDb, tx)

		if err != nil {
			tx.Rollback()

			continue
		}

		if !exists {
			ev.Delete(app.SqlDb, tx)

			database.CommitOrRollbackSQLiteTransaction(tx)

			continue
		}

		activity, err := app.StravaSvc.GetActivity(athlete.Id, ev.ObjectId, app.SqlDb, tx)
		if err != nil {
			tx.Rollback()

			continue
		}

		allowedSportTypes := []string{"Hike", "Run", "TrailRun", "VirtualRun", "Walk", "Wheelchair"}
		shouldAccept := slices.Contains(allowedSportTypes, activity.SportType)

		if shouldAccept {
			if err = activity.Save(app.SqlDb, tx); err != nil {
				tx.Rollback()

				continue
			}
		}

		if err = ev.Delete(app.SqlDb, tx); err != nil {
			tx.Rollback()

			continue
		}

		database.CommitOrRollbackSQLiteTransaction(tx)
	}
}
