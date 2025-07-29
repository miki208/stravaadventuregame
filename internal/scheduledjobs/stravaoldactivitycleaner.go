package scheduledjobs

import (
	"log/slog"
	"time"

	"github.com/miki208/stravaadventuregame/internal/application"
	"github.com/miki208/stravaadventuregame/internal/model"
)

func StravaOldActivityCleaner(app *application.App) {
	deleteActivitiesOlderThan := int(time.Now().Unix()) - app.StravaSvc.GetDeleteOldActivitiesAfterDays()*24*60*60

	slog.Info("StravaOldActivityCleaner started.", "deleteActivitiesOlderThan", deleteActivitiesOlderThan)

	activitiesForDeletion, err := model.AllActivities(app.SqlDb, nil, map[string]any{
		"start_date": model.ComparationOperation{Operation: "<=", FieldValue: deleteActivitiesOlderThan},
	})

	if err != nil {
		slog.Error("Failed to retrieve activities for deletion.", "error", err)

		return
	}

	for _, activity := range activitiesForDeletion {
		err = activity.Delete(app.SqlDb, nil)
		if err != nil {
			slog.Error("Failed to delete activity.", "activity_id", activity, "error", err)

			continue
		}

		slog.Info("Deleted old activity.", "activity_id", activity.Id, "start_date", activity.StartDate)
	}

	slog.Info("StravaOldActivityCleaner finished.")
}
