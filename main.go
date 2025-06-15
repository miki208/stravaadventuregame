package main

import (
	"net/http"

	"github.com/miki208/stravaadventuregame/internal/application"
	"github.com/miki208/stravaadventuregame/internal/handler"
	"github.com/miki208/stravaadventuregame/internal/handler/auth"
	"github.com/miki208/stravaadventuregame/internal/handler/noauth"
	"github.com/miki208/stravaadventuregame/internal/scheduledjobs"
)

func main() {
	app := application.MakeApp()

	for _, job := range scheduledjobs.GetScheduledJobs() {
		app.CronSvc.AddJob(job)
	}

	app.CronSvc.Start()
	defer app.CronSvc.Stop() // I'm aware this does not make sense as long as I'm using http.ListenAndServe, but it's a placeholder for future use.

	http.HandleFunc(app.DefaultPageLoggedOutUsers, handler.MakeHandlerWoutSession(app, noauth.Authorize))
	http.HandleFunc(app.StravaSvc.GetAuthorizationCallback(), handler.MakeHandlerWoutSession(app, noauth.StravaAuthCallback))
	http.HandleFunc(app.DefaultPageLoggedInUsers, handler.MakeHandlerWSession(app, auth.Welcome))
	http.HandleFunc("/start-adventure", handler.MakeHandlerWSession(app, auth.StartAdventure))
	http.HandleFunc("/logout", handler.MakeHandlerWSession(app, auth.Logout))
	http.HandleFunc("/deauthorize", handler.MakeHandlerWSession(app, auth.Deauthorize))
	http.HandleFunc(app.AdminPanelPage, handler.MakeHandlerWSession(app, auth.AdminPanel))
	http.HandleFunc("/stravawebhook/delete", handler.MakeHandlerWSession(app, auth.DeleteStravaWebhookSubscription))
	http.HandleFunc("/stravawebhook/create", handler.MakeHandlerWSession(app, auth.CreateStravaWebhookSubscription))
	http.HandleFunc(app.StravaSvc.GetWebhookCallback(), handler.MakeHandlerWoutSession(app, noauth.StravaWebhookCallback))

	http.ListenAndServe(":80", nil)
}
