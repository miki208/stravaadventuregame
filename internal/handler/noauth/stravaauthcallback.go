package noauth

import (
	"net/http"

	"github.com/miki208/stravaadventuregame/internal/application"
	"github.com/miki208/stravaadventuregame/internal/database"
	"github.com/miki208/stravaadventuregame/internal/model"
)

func StravaAuthCallback(w http.ResponseWriter, req *http.Request, app *application.App) {
	query := req.URL.Query()

	if query.Has("error") {
		http.Redirect(w, req, "/?error="+query.Get("error"), http.StatusFound)

		return
	}

	if !query.Has("code") {
		http.Redirect(w, req, "/?error=code_missing", http.StatusFound)

		return
	}

	if !query.Has("scope") {
		http.Redirect(w, req, "/?error=scope_missing", http.StatusFound)

		return
	}

	if !app.StravaSvc.ValidateScope(query.Get("scope")) {
		http.Redirect(w, req, "/?error=invalid_scope", http.StatusFound)

		return
	}

	athlete, credentials, err := app.StravaSvc.ExchangeToken(query.Get("code"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusFailedDependency)

		return
	}

	tx, err := app.SqlDb.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
	defer tx.Rollback()

	// determine if the athlete already exists (is it updating or creating a new one?)
	existingAthlete := model.NewAthlete()
	existingFound, err := existingAthlete.Load(athlete.Id, app.SqlDb, tx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	if existingFound {
		// this is to resolve a bug where admin status was lost during login

		athlete.SetIsAdmin(existingAthlete.IsAdmin())
	}

	err = athlete.Save(app.SqlDb, tx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	err = credentials.Save(app.SqlDb, tx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	err = database.CommitOrRollbackSQLiteTransaction(tx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	session := app.SessionMgr.CreateSession(athlete.Id)
	http.SetCookie(w, &session.SessionCookie)

	http.Redirect(w, req, app.DefaultPageLoggedInUsers, http.StatusFound)
}
