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
	UseTls                    bool
	InsecurePort              int
	Hostname                  string
	ProxyPathPrefix           string
	defaultPageLoggedInUsers  string
	defaultPageLoggedOutUsers string
	adminPanelPage            string

	Templates  *template.Template
	SessionMgr *helper.SessionManager

	PathToCertCache string

	SqlDb  *sql.DB
	FileDb *database.FileDatabase

	StravaSvc *strava.Strava
	OrsSvc    *openrouteservice.OpenRouteService

	CronSvc *Cron

	logFile *os.File

	SupportedActivityTypes []string
}

func (app *App) GetDefaultPageLoggedInUsers() string {
	return app.ProxyPathPrefix + app.defaultPageLoggedInUsers
}

func (app *App) GetDefaultPageLoggedInUsersWithoutProxyPathPrefix() string {
	return app.defaultPageLoggedInUsers
}

func (app *App) GetDefaultPageLoggedOutUsers() string {
	return app.ProxyPathPrefix + app.defaultPageLoggedOutUsers
}

func (app *App) GetDefaultPageLoggedOutUsersWithoutProxyPathPrefix() string {
	return app.defaultPageLoggedOutUsers
}

func (app *App) GetAdminPanelPage() string {
	return app.ProxyPathPrefix + app.adminPanelPage
}

func (app *App) GetAdminPanelPageWithoutProxyPathPrefix() string {
	return app.adminPanelPage
}

func (app *App) GetFullAuthorizationCallbackUrl() string {
	return "https://" + app.Hostname + app.ProxyPathPrefix + app.StravaSvc.GetAuthorizationCallback()
}

func (app *App) GetFullWebhookCallbackUrl() string {
	return "https://" + app.Hostname + app.ProxyPathPrefix + app.StravaSvc.GetWebhookCallback()
}

func (app *App) Close() error {
	if app.logFile != nil {
		app.logFile.Close()
	}

	return nil
}

func MakeApp(configFileName string) *App {
	// first load the configuration
	var conf config

	err := conf.loadFromFile(configFileName)
	if err != nil {
		panic(fmt.Errorf("failed to load configuration from file %s: %w", configFileName, err))
	}

	if err = conf.validate(); err != nil {
		panic(fmt.Errorf("config validation failed: %w", err))
	}

	// then initalize the logger
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
		Level: conf.getLoggingLevel(),
	}))

	slog.SetDefault(logger)

	templates := getTemplateFileNames(conf.PathToTemplates)

	app := &App{
		UseTls:                    conf.UseTls,
		InsecurePort:              conf.InsecurePort,
		Hostname:                  conf.Hostname,
		ProxyPathPrefix:           conf.ProxyPathPrefix,
		defaultPageLoggedInUsers:  conf.DefaultPageLoggedInUsers,
		defaultPageLoggedOutUsers: conf.DefaultPageLoggedOutUsers,
		adminPanelPage:            conf.AdminPanelPage,

		Templates:  template.Must(template.ParseFiles(templates...)),
		SessionMgr: helper.CreateSessionManager(conf.SessionDurationInMinutes),

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

		SupportedActivityTypes: conf.SupportedActivityTypes,
	}

	app.CronSvc = NewCron(app, conf.ScheduledJobIntervalSec)

	return app
}
