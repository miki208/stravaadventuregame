package auth

import (
	"net/http"

	"github.com/miki208/stravaadventuregame/internal/application"
	"github.com/miki208/stravaadventuregame/internal/helper"
	"github.com/miki208/stravaadventuregame/internal/model"
)

func Welcome(resp http.ResponseWriter, req *http.Request, app *application.App, session helper.Session) {
	var athlete model.Athlete

	exists, err := athlete.LoadById(session.UserId, app.SqlDb, nil)
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)

		return
	}

	if !exists {
		app.SessionMgr.DestroySession(session)

		http.Redirect(resp, req, "/?error=user_not_found", http.StatusFound)

		return
	}

	availableLocations, err := model.GetLocations(app.SqlDb, nil)
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)

		return
	}

	type AdventureExtended struct {
		Adventure     *model.Adventure
		StartLocation *model.Location
		EndLocation   *model.Location
	}

	adventureToAdventureExtended := func(adv *model.Adventure) (*AdventureExtended, error) {
		var startLocation, endLocation model.Location

		found, err := startLocation.LoadById(adv.StartLocation, app.SqlDb, nil)
		if err != nil || !found {
			return nil, err
		}

		found, err = endLocation.LoadById(adv.EndLocation, app.SqlDb, nil)
		if err != nil || !found {
			return nil, err
		}

		return &AdventureExtended{
			Adventure:     adv,
			StartLocation: &startLocation,
			EndLocation:   &endLocation,
		}, nil
	}

	var startedAdventuresExtended []*AdventureExtended

	startedAdventures, err := model.GetAdventuresByAthlete(athlete.Id, model.FilterNotCompleted, app.SqlDb, nil)
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)

		return
	}

	for _, adv := range startedAdventures {
		adventureExtended, err := adventureToAdventureExtended(adv)
		if err != nil {
			http.Error(resp, err.Error(), http.StatusInternalServerError)

			return
		}

		startedAdventuresExtended = append(startedAdventuresExtended, adventureExtended)
	}

	var completedAdventuresExtended []*AdventureExtended

	completedAdventures, err := model.GetAdventuresByAthlete(athlete.Id, model.FilterCompleted, app.SqlDb, nil)
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)

		return
	}

	for _, adv := range completedAdventures {
		adventureExtended, err := adventureToAdventureExtended(adv)
		if err != nil {
			http.Error(resp, err.Error(), http.StatusInternalServerError)

			return
		}

		completedAdventuresExtended = append(completedAdventuresExtended, adventureExtended)
	}

	err = app.Templates.ExecuteTemplate(resp, "welcome.html", struct {
		Athl                model.Athlete
		StartedAdventures   []*AdventureExtended
		CompletedAdventures []*AdventureExtended
		AvailableLocations  []*model.Location
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
