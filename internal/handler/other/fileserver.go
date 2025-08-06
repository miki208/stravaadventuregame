package other

import (
	"net/http"

	"github.com/miki208/stravaadventuregame/internal/application"
)

func FileServer(w http.ResponseWriter, r *http.Request, app *application.App) error {
	http.StripPrefix("/static/", http.FileServer(http.Dir("static"))).ServeHTTP(w, r)

	return nil
}
