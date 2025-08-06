package main

import (
	"flag"

	"log/slog"

	"github.com/miki208/stravaadventuregame/internal/application"
	"github.com/miki208/stravaadventuregame/internal/handler"
	"github.com/miki208/stravaadventuregame/internal/handler/auth"
	"github.com/miki208/stravaadventuregame/internal/handler/noauth"
	"github.com/miki208/stravaadventuregame/internal/handler/other"
	"github.com/miki208/stravaadventuregame/internal/scheduledjobs"
)

func main() {
	configFileName := flag.String("config", "config.ini", "Path to the configuration file")
	flag.Parse()

	app := application.MakeApp(*configFileName)
	defer app.Close()

	slog.Info("Application started.", "hostname", app.Hostname)

	slog.Info("Registering scheduled jobs...")
	for _, job := range scheduledjobs.GetScheduledJobs() {
		app.CronSvc.AddJob(job)
	}
	slog.Info("Scheduled jobs registered.")

	app.CronSvc.Start()
	defer app.CronSvc.Stop()

	srvFactory := application.ServerFactory{
		Hostname:        app.Hostname,
		PathToCertCache: app.PathToCertCache,
	}

	srv := srvFactory.CreateServer(app.ServerType)

	srv.AddRoute(app.DefaultPageLoggedOutUsers, handler.MakeHandlerWoutSession(app, noauth.Authorize))
	srv.AddRoute(app.StravaSvc.GetAuthorizationCallback(), handler.MakeHandlerWoutSession(app, noauth.StravaAuthCallback))
	srv.AddRoute(app.DefaultPageLoggedInUsers, handler.MakeHandlerWSession(app, auth.Welcome))
	srv.AddRoute("/start-adventure", handler.MakeHandlerWSession(app, auth.StartAdventure))
	srv.AddRoute("/logout", handler.MakeHandlerWSession(app, auth.Logout))
	srv.AddRoute("/deauthorize", handler.MakeHandlerWSession(app, auth.Deauthorize))
	srv.AddRoute("/settings", handler.MakeHandlerWSession(app, auth.Settings))
	srv.AddRoute(app.AdminPanelPage, handler.MakeHandlerWSession(app, auth.AdminPanel))
	srv.AddRoute("/stravawebhook/delete", handler.MakeHandlerWSession(app, auth.DeleteStravaWebhookSubscription))
	srv.AddRoute("/stravawebhook/create", handler.MakeHandlerWSession(app, auth.CreateStravaWebhookSubscription))
	srv.AddRoute(app.StravaSvc.GetWebhookCallback(), handler.MakeHandlerWoutSession(app, noauth.StravaWebhookCallback))
	srv.AddRoute("/static/", handler.MakeHandler(app, other.FileServer))

	srv.ListenAndServe()
}
