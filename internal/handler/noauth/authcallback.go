package noauth

import (
	"net/http"

	"github.com/miki208/stravaadventuregame/internal/application"
	"github.com/miki208/stravaadventuregame/internal/database"
	"github.com/miki208/stravaadventuregame/internal/service/strava"
)

func AuthorizationCallback(w http.ResponseWriter, req *http.Request, app *application.App) {
	query := req.URL.Query()

	if query.Has("error") {
		http.Redirect(w, req, "/?error="+query.Get("error"), http.StatusFound)

		return
	}

	if !query.Has("code") {
		http.Redirect(w, req, "/?error=code_missing", http.StatusFound)

		return
	}

	athlete, credentials, err := app.StravaSvc.ExchangeToken(query.Get("code"))
	if err != nil {
		stravaError := err.(*strava.StravaError)

		http.Error(w, stravaError.Error(), stravaError.StatusCode())

		return
	}

	_, tx, err := database.GetOrCreateSQLiteTransaction(app.SqlDb, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
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
