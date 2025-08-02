package auth

import (
	"net/http"
	"time"

	"github.com/miki208/stravaadventuregame/internal/application"
	"github.com/miki208/stravaadventuregame/internal/handler"
	"github.com/miki208/stravaadventuregame/internal/model"
)

func Welcome(resp *handler.ResponseWithSession, req *http.Request, app *application.App) error {
	athlete := model.NewAthlete()

	exists, err := athlete.Load(resp.Session().UserId, app.SqlDb, nil)
	if err != nil {
		return handler.NewHandlerError(http.StatusInternalServerError, err)
	}

	if !exists {
		app.SessionMgr.DestroySession(resp.Session())

		http.Redirect(resp, req, "/?error=user_not_found", http.StatusFound)

		return nil
	}

	availableLocations, err := model.AllLocations(app.SqlDb, nil, nil)
	if err != nil {
		return handler.NewHandlerError(http.StatusInternalServerError, err)
	}

	type AdventureExtended struct {
		Adventure         *model.Adventure
		StartLocation     *model.Location
		EndLocation       *model.Location
		StartDateFormated string
		EndDateFormated   string
	}

	adventureToAdventureExtended := func(adv *model.Adventure) (AdventureExtended, error) {
		var startLocation, endLocation model.Location

		found, err := startLocation.Load(adv.StartLocation, app.SqlDb, nil)
		if err != nil || !found {
			return AdventureExtended{}, err
		}

		found, err = endLocation.Load(adv.EndLocation, app.SqlDb, nil)
		if err != nil || !found {
			return AdventureExtended{}, err
		}

		return AdventureExtended{
			Adventure:         adv,
			StartLocation:     &startLocation,
			EndLocation:       &endLocation,
			StartDateFormated: time.Unix(int64(adv.StartDate), 0).UTC().Format(time.DateTime),
			EndDateFormated:   time.Unix(int64(adv.EndDate), 0).UTC().Format(time.DateTime),
		}, nil
	}

	var startedAdventuresExtended []AdventureExtended

	startedAdventures, err := model.AllAdventures(app.SqlDb, nil, map[string]any{"athlete_id": athlete.Id, "completed": 0})
	if err != nil {
		return handler.NewHandlerError(http.StatusInternalServerError, err)
	}

	for _, adv := range startedAdventures {
		adventureExtended, err := adventureToAdventureExtended(&adv)
		if err != nil {
			return handler.NewHandlerError(http.StatusInternalServerError, err)
		}

		startedAdventuresExtended = append(startedAdventuresExtended, adventureExtended)
	}

	var completedAdventuresExtended []AdventureExtended

	completedAdventures, err := model.AllAdventures(app.SqlDb, nil, map[string]any{"athlete_id": athlete.Id, "completed": 1})
	if err != nil {
		return handler.NewHandlerError(http.StatusInternalServerError, err)
	}

	for _, adv := range completedAdventures {
		adventureExtended, err := adventureToAdventureExtended(&adv)
		if err != nil {
			return handler.NewHandlerError(http.StatusInternalServerError, err)
		}

		completedAdventuresExtended = append(completedAdventuresExtended, adventureExtended)
	}

	err = app.Templates.ExecuteTemplate(resp, "welcome.html", struct {
		Athl                *model.Athlete
		StartedAdventures   []AdventureExtended
		CompletedAdventures []AdventureExtended
		AvailableLocations  []model.Location
	}{
		Athl:                athlete,
		StartedAdventures:   startedAdventuresExtended,
		CompletedAdventures: completedAdventuresExtended,
		AvailableLocations:  availableLocations,
	})
	if err != nil {
		return handler.NewHandlerError(http.StatusInternalServerError, err)
	}

	return nil
}
