package auth

import (
	"net/http"

	"github.com/miki208/stravaadventuregame/internal/application"
	"github.com/miki208/stravaadventuregame/internal/helper"
	"github.com/miki208/stravaadventuregame/internal/model"
)

func Welcome(resp http.ResponseWriter, req *http.Request, app *application.App, session helper.Session) {
	athlete := model.NewAthlete()

	exists, err := athlete.Load(session.UserId, app.SqlDb, nil)
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)

		return
	}

	if !exists {
		app.SessionMgr.DestroySession(session)

		http.Redirect(resp, req, "/?error=user_not_found", http.StatusFound)

		return
	}

	availableLocations, err := model.AllLocations(app.SqlDb, nil, nil)
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)

		return
	}

	type AdventureExtended struct {
		Adventure     *model.Adventure
		StartLocation *model.Location
		EndLocation   *model.Location
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
			Adventure:     adv,
			StartLocation: &startLocation,
			EndLocation:   &endLocation,
		}, nil
	}

	var startedAdventuresExtended []AdventureExtended

	startedAdventures, err := model.AllAdventures(app.SqlDb, nil, map[string]any{"athlete_id": athlete.Id, "completed": 0})
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)

		return
	}

	for _, adv := range startedAdventures {
		adventureExtended, err := adventureToAdventureExtended(&adv)
		if err != nil {
			http.Error(resp, err.Error(), http.StatusInternalServerError)

			return
		}

		startedAdventuresExtended = append(startedAdventuresExtended, adventureExtended)
	}

	var completedAdventuresExtended []AdventureExtended

	completedAdventures, err := model.AllAdventures(app.SqlDb, nil, map[string]any{"athlete_id": athlete.Id, "completed": 1})
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)

		return
	}

	for _, adv := range completedAdventures {
		adventureExtended, err := adventureToAdventureExtended(&adv)
		if err != nil {
			http.Error(resp, err.Error(), http.StatusInternalServerError)

			return
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
		http.Error(resp, err.Error(), http.StatusInternalServerError)

		return
	}
}
