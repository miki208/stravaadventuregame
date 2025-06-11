package handler

import (
	"net/http"

	"github.com/miki208/stravaadventuregame/internal/application"
	"github.com/miki208/stravaadventuregame/internal/helper"
)

type FuncHandlerWSession func(http.ResponseWriter, *http.Request, *application.App, helper.Session)
type FuncHandler func(http.ResponseWriter, *http.Request, *application.App)

func MakeHandlerWSession(app *application.App, fn FuncHandlerWSession) func(http.ResponseWriter, *http.Request) {
	return func(resp http.ResponseWriter, req *http.Request) {
		session := app.SessionMgr.GetSessionByRequest(req)
		if session == nil {
			http.Redirect(resp, req, app.DefaultPageLoggedOutUsers, http.StatusFound)

			return
		}

		fn(resp, req, app, *session)
	}
}

func MakeHandlerWoutSession(app *application.App, fn FuncHandler) func(http.ResponseWriter, *http.Request) {
	return func(resp http.ResponseWriter, req *http.Request) {
		session := app.SessionMgr.GetSessionByRequest(req)
		if session != nil {
			http.Redirect(resp, req, app.DefaultPageLoggedInUsers, http.StatusFound)

			return
		}

		fn(resp, req, app)
	}
}

func MakeHandler(app *application.App, fn FuncHandler) func(http.ResponseWriter, *http.Request) {
	return func(resp http.ResponseWriter, req *http.Request) {
		fn(resp, req, app)
	}
}
