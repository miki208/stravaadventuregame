package scheduledjobs

import "github.com/miki208/stravaadventuregame/internal/application"

func GetScheduledJobs() []application.CronJob {
	return []application.CronJob{
		StravaPendingActivityProcessor,
		StravaOldActivityCleaner,
	}
}
