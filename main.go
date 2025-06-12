package main

import (
	"net/http"

	"github.com/miki208/stravaadventuregame/internal/application"
	"github.com/miki208/stravaadventuregame/internal/handler"
	"github.com/miki208/stravaadventuregame/internal/handler/auth"
	"github.com/miki208/stravaadventuregame/internal/handler/noauth"
)

func main() {
	app := application.MakeApp()

	http.HandleFunc(app.DefaultPageLoggedOutUsers, handler.MakeHandlerWoutSession(app, noauth.Authorize))
	http.HandleFunc(app.StravaSvc.GetAuthorizationCallback(), handler.MakeHandlerWoutSession(app, noauth.AuthorizationCallback))
	http.HandleFunc(app.DefaultPageLoggedInUsers, handler.MakeHandlerWSession(app, auth.Welcome))
	http.HandleFunc("/start-adventure", handler.MakeHandlerWSession(app, auth.StartAdventure))
	http.HandleFunc("/logout", handler.MakeHandlerWSession(app, auth.Logout))
	http.HandleFunc("/deauthorize", handler.MakeHandlerWSession(app, auth.Deauthorize))

	http.ListenAndServe(":80", nil)
}
