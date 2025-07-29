package handler

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/miki208/stravaadventuregame/internal/application"
	"github.com/miki208/stravaadventuregame/internal/helper"
)

type HandlerError struct {
	statusCode int
	err        error
}

func NewHandlerError(statusCode int, err error) *HandlerError {
	return &HandlerError{
		statusCode: statusCode,
		err:        err,
	}
}

func (handlerError *HandlerError) StatusCode() int {
	return handlerError.statusCode
}

func (handlerError *HandlerError) Error() string {
	return fmt.Sprintf("Handler error (%d): %v", handlerError.statusCode, handlerError.err)
}

type FuncHandlerWSession func(http.ResponseWriter, *http.Request, *application.App, helper.Session) error
type FuncHandler func(http.ResponseWriter, *http.Request, *application.App) error

func MakeHandlerWSession(app *application.App, fn FuncHandlerWSession) func(http.ResponseWriter, *http.Request) {
	return func(resp http.ResponseWriter, req *http.Request) {
		session := app.SessionMgr.GetSessionByRequest(req)
		if session == nil {
			http.Redirect(resp, req, app.DefaultPageLoggedOutUsers, http.StatusFound)

			return
		}

		err := fn(resp, req, app, *session)
		if err != nil {
			slog.Error("HandlerWSession > Error occurred while handling request.", "error", err, "route", req.URL.Path, "session_id", session.SessionCookie.Value, "user_id", session.UserId)

			if handlerError, ok := err.(*HandlerError); ok {
				http.Error(resp, handlerError.Error(), handlerError.StatusCode())
			} else {
				http.Error(resp, err.Error(), http.StatusInternalServerError)
			}

			return
		}
	}
}

func MakeHandlerWoutSession(app *application.App, fn FuncHandler) func(http.ResponseWriter, *http.Request) {
	return func(resp http.ResponseWriter, req *http.Request) {
		session := app.SessionMgr.GetSessionByRequest(req)
		if session != nil {
			http.Redirect(resp, req, app.DefaultPageLoggedInUsers, http.StatusFound)

			return
		}

		err := fn(resp, req, app)
		if err != nil {
			slog.Error("HandlerWoutSession > Error occurred while handling request.", "error", err, "route", req.URL.Path)

			if handlerError, ok := err.(*HandlerError); ok {
				http.Error(resp, handlerError.Error(), handlerError.StatusCode())
			} else {
				http.Error(resp, err.Error(), http.StatusInternalServerError)
			}

			return
		}
	}
}

func MakeHandler(app *application.App, fn FuncHandler) func(http.ResponseWriter, *http.Request) {
	return func(resp http.ResponseWriter, req *http.Request) {
		err := fn(resp, req, app)
		if err != nil {
			slog.Error("Handler > Error occurred while handling request.", "error", err, "route", req.URL.Path)

			if handlerError, ok := err.(*HandlerError); ok {
				http.Error(resp, handlerError.Error(), handlerError.StatusCode())
			} else {
				http.Error(resp, err.Error(), http.StatusInternalServerError)
			}

			return
		}
	}
}
