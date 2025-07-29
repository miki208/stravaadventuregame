package noauth

import (
	"net/http"

	"github.com/miki208/stravaadventuregame/internal/application"
	"github.com/miki208/stravaadventuregame/internal/database"
	"github.com/miki208/stravaadventuregame/internal/handler"
	"github.com/miki208/stravaadventuregame/internal/model"
)

func StravaAuthCallback(w http.ResponseWriter, req *http.Request, app *application.App) error {
	query := req.URL.Query()

	if query.Has("error") {
		http.Redirect(w, req, "/?error="+query.Get("error"), http.StatusFound)

		return nil
	}

	if !query.Has("code") {
		http.Redirect(w, req, "/?error=code_missing", http.StatusFound)

		return nil
	}

	if !query.Has("scope") {
		http.Redirect(w, req, "/?error=scope_missing", http.StatusFound)

		return nil
	}

	if !app.StravaSvc.ValidateScope(query.Get("scope")) {
		http.Redirect(w, req, "/?error=invalid_scope", http.StatusFound)

		return nil
	}

	athlete, credentials, err := app.StravaSvc.ExchangeToken(query.Get("code"))
	if err != nil {
		return handler.NewHandlerError(http.StatusFailedDependency, err)
	}

	tx, err := app.SqlDb.Begin()
	if err != nil {
		return handler.NewHandlerError(http.StatusInternalServerError, err)
	}
	defer tx.Rollback()

	// determine if the athlete already exists (is it updating or creating a new one?)
	existingAthlete := model.NewAthlete()
	existingFound, err := existingAthlete.Load(athlete.Id, app.SqlDb, tx)
	if err != nil {
		return handler.NewHandlerError(http.StatusInternalServerError, err)
	}

	if existingFound {
		// this is to resolve a bug where admin status was lost during login

		athlete.SetIsAdmin(existingAthlete.IsAdmin())
	}

	err = athlete.Save(app.SqlDb, tx)
	if err != nil {
		return handler.NewHandlerError(http.StatusInternalServerError, err)
	}

	err = credentials.Save(app.SqlDb, tx)
	if err != nil {
		return handler.NewHandlerError(http.StatusInternalServerError, err)
	}

	err = database.CommitOrRollbackSQLiteTransaction(tx)
	if err != nil {
		return handler.NewHandlerError(http.StatusInternalServerError, err)
	}

	session := app.SessionMgr.CreateSession(athlete.Id)
	http.SetCookie(w, &session.SessionCookie)

	http.Redirect(w, req, app.DefaultPageLoggedInUsers, http.StatusFound)

	return nil
}
