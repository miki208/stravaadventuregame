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

	Templates  *template.Template
	SessionMgr *helper.SessionManager

	SqlDb  *sql.DB
	FileDb *database.FileDatabase

	StravaSvc *strava.Strava
	OrsSvc    *openrouteservice.OpenRouteService
}

func (app *App) GetFullAuthorizationCallbackUrl() string {
	return "http://" + app.Hostname + app.StravaSvc.GetAuthorizationCallback()
}

func MakeApp() *App {
	var conf config

	err := conf.loadFromFile("config.ini")
	if err != nil {
		panic(err)
	}

	if !conf.validate() {
		panic("config validation failed")
	}

	templates := getTemplateFileNames(conf.PathToTemplates)

	return &App{
		Hostname:                  conf.Hostname,
		DefaultPageLoggedInUsers:  conf.DefaultPageLoggedInUsers,
		DefaultPageLoggedOutUsers: conf.DefaultPageLoggedOutUsers,

		Templates:  template.Must(template.ParseFiles(templates...)),
		SessionMgr: helper.CreateSessionManager(),

		SqlDb:  database.CreateSQLiteDatabase(conf.SqliteDbPath),
		FileDb: database.CreateFileDatabase(conf.FileDbPath),

		StravaSvc: strava.CreateService(
			conf.StravaConf.ClientId,
			conf.StravaConf.ClientSecret,
			conf.StravaConf.AuthorizationCallback,
			conf.StravaConf.Scope),
		OrsSvc: openrouteservice.CreateService(conf.OrsConf.ApiKey),
	}
}
