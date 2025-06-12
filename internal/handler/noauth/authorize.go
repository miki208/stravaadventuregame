package noauth

import (
	"net/http"

	"github.com/miki208/stravaadventuregame/internal/application"
)

func Authorize(w http.ResponseWriter, req *http.Request, app *application.App) {
	err := app.Templates.ExecuteTemplate(w, "authorize.html", struct {
		ClientId    int
		RedirectUri string
		Scope       string
		Error       string
	}{
		ClientId:    app.StravaSvc.GetClientId(),
		RedirectUri: app.GetFullAuthorizationCallbackUrl(),
		Scope:       app.StravaSvc.GetScope(),
		Error:       req.URL.Query().Get("error"),
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
}
