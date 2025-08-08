package auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/miki208/stravaadventuregame/internal/application"
	"github.com/miki208/stravaadventuregame/internal/handler"
	"github.com/miki208/stravaadventuregame/internal/helper"
	"github.com/miki208/stravaadventuregame/internal/model"
	"github.com/paulmach/orb"
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
		Adventure          *model.Adventure
		CompletedRoute     orb.LineString
		NotCompletedRoute  orb.LineString
		StartLocation      *model.Location
		EndLocation        *model.Location
		StartDateFormatted string
		EndDateFormatted   string
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

		var completedRoute, notCompletedRoute orb.LineString
		if adv.Completed == 0 { // we're going to populate routes for not completed adventures only
			courseDbName := fmt.Sprintf("%d-%d", min(adv.StartLocation, adv.EndLocation),
				max(adv.StartLocation, adv.EndLocation))

			var route *model.DirectionsRoute = model.NewDirectionsRoute()
			err := app.FileDb.Read("course", courseDbName, route)
			if err != nil {
				return AdventureExtended{}, err
			}

			routePolyline, err := helper.DecodePolyline(route.Geometry, adv.StartLocation > adv.EndLocation)
			if err != nil {
				return AdventureExtended{}, err
			}

			if adv.CurrentLocationIndexOnRoute == -1 {
				completedRoute = routePolyline
			} else {
				for i := 0; i < len(routePolyline); i++ {
					if i < adv.CurrentLocationIndexOnRoute {
						completedRoute = append(completedRoute, orb.Point{routePolyline[i].Lon(), routePolyline[i].Lat()})
					} else if i > adv.CurrentLocationIndexOnRoute {
						notCompletedRoute = append(notCompletedRoute, orb.Point{routePolyline[i].Lon(), routePolyline[i].Lat()})
					}

					if i == adv.CurrentLocationIndexOnRoute {
						completedRoute = append(completedRoute, orb.Point{routePolyline[i].Lon(), routePolyline[i].Lat()})

						if routePolyline[i].Lat() != adv.CurrentLocationLat || routePolyline[i].Lon() != adv.CurrentLocationLon {
							completedRoute = append(completedRoute, orb.Point{adv.CurrentLocationLon, adv.CurrentLocationLat})
							notCompletedRoute = append(notCompletedRoute, orb.Point{adv.CurrentLocationLon, adv.CurrentLocationLat})
						} else {
							notCompletedRoute = append(notCompletedRoute, orb.Point{routePolyline[i].Lon(), routePolyline[i].Lat()})
						}
					}
				}
			}
		}

		return AdventureExtended{
			Adventure:          adv,
			CompletedRoute:     completedRoute,
			NotCompletedRoute:  notCompletedRoute,
			StartLocation:      &startLocation,
			EndLocation:        &endLocation,
			StartDateFormatted: time.Unix(int64(adv.StartDate), 0).UTC().Format(time.DateTime),
			EndDateFormatted:   time.Unix(int64(adv.EndDate), 0).UTC().Format(time.DateTime),
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
