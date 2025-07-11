package application

import (
	"database/sql"
	"html/template"

	"github.com/miki208/stravaadventuregame/internal/database"
	"github.com/miki208/stravaadventuregame/internal/helper"
	"github.com/miki208/stravaadventuregame/internal/service/openrouteservice"
	"github.com/miki208/stravaadventuregame/internal/service/strava"
)

type App struct {
	Hostname                  string
	DefaultPageLoggedInUsers  string
	DefaultPageLoggedOutUsers string
	AdminPanelPage            string

	Templates  *template.Template
	SessionMgr *helper.SessionManager

	SqlDb  *sql.DB
	FileDb *database.FileDatabase

	StravaSvc *strava.Strava
	OrsSvc    *openrouteservice.OpenRouteService

	CronSvc *Cron
}

func (app *App) GetFullAuthorizationCallbackUrl() string {
	return "https://" + app.Hostname + app.StravaSvc.GetAuthorizationCallback()
}

func (app *App) GetFullWebhookCallbackUrl() string {
	return "https://" + app.Hostname + app.StravaSvc.GetWebhookCallback()
}

func MakeApp(configFileName string) *App {
	var conf config

	err := conf.loadFromFile(configFileName)
	if err != nil {
		panic(err)
	}

	if !conf.validate() {
		panic("config validation failed")
	}

	templates := getTemplateFileNames(conf.PathToTemplates)

	app := &App{
		Hostname:                  conf.Hostname,
		DefaultPageLoggedInUsers:  conf.DefaultPageLoggedInUsers,
		DefaultPageLoggedOutUsers: conf.DefaultPageLoggedOutUsers,
		AdminPanelPage:            conf.AdminPanelPage,

		Templates:  template.Must(template.ParseFiles(templates...)),
		SessionMgr: helper.CreateSessionManager(),

		SqlDb:  database.CreateSQLiteDatabase(conf.SqliteDbPath),
		FileDb: database.CreateFileDatabase(conf.FileDbPath),

		StravaSvc: strava.CreateService(
			conf.StravaConf.ClientId,
			conf.StravaConf.ClientSecret,
			conf.StravaConf.AuthorizationCallback,
			conf.StravaConf.Scope,
			conf.StravaConf.WebhookCallback,
			conf.StravaConf.VerifyToken,
			conf.StravaConf.DeleteOldActivitiesAfterDays,
			conf.StravaConf.ProcessWebhookEventsAfterSec),
		OrsSvc: openrouteservice.CreateService(conf.OrsConf.ApiKey),
	}

	app.CronSvc = NewCron(app, conf.ScheduledJobIntervalSec)

	return app
}
