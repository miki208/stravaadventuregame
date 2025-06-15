package scheduledjobs

import (
	"time"

	"github.com/miki208/stravaadventuregame/internal/application"
	"github.com/miki208/stravaadventuregame/internal/model"
)

func StravaOldActivityCleaner(app *application.App) {
	deleteActivitiesOlderThan := int(time.Now().Unix()) - app.StravaSvc.GetDeleteOldActivitiesAfterDays()*24*60*60

	activitiesForDeletion, err := model.AllActivities(app.SqlDb, nil, map[string]any{
		"start_date": model.ComparationOperation{Operation: "<=", FieldValue: deleteActivitiesOlderThan},
	})

	if err != nil {
		return
	}

	for _, activity := range activitiesForDeletion {
		err = activity.Delete(app.SqlDb, nil)
		if err != nil {
			// log something here
		}
	}
}
