package auth

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/miki208/stravaadventuregame/internal/application"
	"github.com/miki208/stravaadventuregame/internal/handler"
	"github.com/miki208/stravaadventuregame/internal/model"
)

func Settings(resp *handler.ResponseWithSession, req *http.Request, app *application.App) error {
	if http.MethodGet != req.Method && http.MethodPost != req.Method {
		return handler.NewHandlerError(http.StatusMethodNotAllowed, nil)
	}

	if http.MethodGet == req.Method {
		var athleteSettings model.AthleteSettings
		settingsFound, err := athleteSettings.Load(resp.Session().UserId, app.SqlDb, nil)
		if err != nil {
			return handler.NewHandlerError(http.StatusInternalServerError, err)
		}

		if !settingsFound {
			return handler.NewHandlerError(http.StatusInternalServerError, fmt.Errorf("settings not found"))
		}

		err = app.Templates.ExecuteTemplate(resp, "settings.html", struct {
			ProxyPathPrefix string
			AthleteSettings model.AthleteSettings
		}{
			ProxyPathPrefix: app.ProxyPathPrefix,
			AthleteSettings: athleteSettings,
		})
		if err != nil {
			return handler.NewHandlerError(http.StatusInternalServerError, err)
		}
	} else if http.MethodPost == req.Method {
		if err := req.ParseForm(); err != nil {
			return handler.NewHandlerError(http.StatusBadRequest, err)
		}

		var athleteSettings model.AthleteSettings

		settingsFound, err := athleteSettings.Load(resp.Session().UserId, app.SqlDb, nil)
		if err != nil {
			return handler.NewHandlerError(http.StatusInternalServerError, err)
		}

		if !settingsFound {
			return handler.NewHandlerError(http.StatusInternalServerError, errors.New("settings not found"))
		}

		if req.FormValue("autoActivityDescUpdate") != "" {
			athleteSettings.AutoUpdateActivityDescription = 1
		} else {
			athleteSettings.AutoUpdateActivityDescription = 0
		}

		err = athleteSettings.Save(app.SqlDb, nil)
		if err != nil {
			return handler.NewHandlerError(http.StatusInternalServerError, err)
		}

		http.Redirect(resp, req, app.ProxyPathPrefix+"/settings", http.StatusFound)
	}

	return nil
}
