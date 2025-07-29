package application

import (
	"database/sql"
	"fmt"
	"html/template"
	"log/slog"
	"os"

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

	PathToCertCache string

	SqlDb  *sql.DB
	FileDb *database.FileDatabase

	StravaSvc *strava.Strava
	OrsSvc    *openrouteservice.OpenRouteService

	CronSvc *Cron

	logFile *os.File
}

func (app *App) GetFullAuthorizationCallbackUrl() string {
	return "https://" + app.Hostname + app.StravaSvc.GetAuthorizationCallback()
}

func (app *App) GetFullWebhookCallbackUrl() string {
	return "https://" + app.Hostname + app.StravaSvc.GetWebhookCallback()
}

func (app *App) Close() error {
	if app.logFile != nil {
		app.logFile.Close()
	}

	return nil
}

func MakeApp(configFileName string) *App {
	// first initalize the logger
	logFile, err := os.OpenFile("application.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}

	defer func() {
		if r := recover(); r != nil {
			slog.Error("Initialization failed.", "error", r)

			logFile.Close()

			panic(r)
		}
	}()

	logger := slog.New(slog.NewTextHandler(logFile, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	slog.SetDefault(logger)

	// then load the configuration
	var conf config

	err = conf.loadFromFile(configFileName)
	if err != nil {
		panic(fmt.Errorf("failed to load configuration from file %s: %w", configFileName, err))
	}

	if err = conf.validate(); err != nil {
		panic(fmt.Errorf("config validation failed: %w", err))
	}

	templates := getTemplateFileNames(conf.PathToTemplates)

	app := &App{
		Hostname:                  conf.Hostname,
		DefaultPageLoggedInUsers:  conf.DefaultPageLoggedInUsers,
		DefaultPageLoggedOutUsers: conf.DefaultPageLoggedOutUsers,
		AdminPanelPage:            conf.AdminPanelPage,

		Templates:  template.Must(template.ParseFiles(templates...)),
		SessionMgr: helper.CreateSessionManager(),

		PathToCertCache: conf.PathToCertCache,

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

		logFile: logFile,
	}

	app.CronSvc = NewCron(app, conf.ScheduledJobIntervalSec)

	return app
}
