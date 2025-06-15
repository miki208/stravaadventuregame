package auth

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/miki208/stravaadventuregame/internal/application"
	"github.com/miki208/stravaadventuregame/internal/helper"
	"github.com/miki208/stravaadventuregame/internal/model"
	"github.com/miki208/stravaadventuregame/internal/service/openrouteservice"
)

func StartAdventure(resp http.ResponseWriter, req *http.Request, app *application.App, session helper.Session) {
	// first get location ids and validate them
	startLocationId, err := strconv.Atoi(req.FormValue("start"))
	if err != nil {
		http.Error(resp, "Start location is not populated.", http.StatusBadRequest)

		return
	}

	stopLocationId, err := strconv.Atoi(req.FormValue("stop"))
	if err != nil {
		http.Error(resp, "Stop location is not populated.", http.StatusBadRequest)

		return
	}

	if startLocationId == stopLocationId {
		http.Error(resp, "Start and stop location can't be the same.", http.StatusBadRequest)

		return
	}

	// make sure that user is not already on an adventure
	adventuresStarted, err := model.AllAdventures(app.SqlDb, nil, map[string]any{"athlete_id": session.UserId, "completed": 0})
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)

		return
	}
	if len(adventuresStarted) > 0 {
		http.Error(resp, "You are already on an adventure.", http.StatusBadRequest)

		return
	}

	// if everything is ok, these locations should be in the database, load them
	var startLocation model.Location
	found, err := startLocation.Load(startLocationId, app.SqlDb, nil)
	if err != nil || !found {
		http.Error(resp, err.Error(), http.StatusInternalServerError)

		return
	}

	var stopLocation model.Location
	found, err = stopLocation.Load(stopLocationId, app.SqlDb, nil)
	if err != nil || !found {
		http.Error(resp, err.Error(), http.StatusInternalServerError)

		return
	}

	// we should check if we have this route in the database before getting it via rest api
	dbName := fmt.Sprintf("%d-%d", min(startLocationId, stopLocationId), max(startLocationId, stopLocationId))

	exists, err := app.FileDb.Exists("course", dbName)
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)

		return
	}

	var adventureCourse *model.DirectionsRoute
	if !exists {
		// no route, retrieve it
		route, err := app.OrsSvc.GetDirections(startLocation.Lat, startLocation.Lon, stopLocation.Lat, stopLocation.Lon, "km")
		if err != nil {
			orsError := err.(*openrouteservice.OpenRouteServiceError)

			http.Error(resp, orsError.Error(), orsError.StatusCode())

			return
		}

		// write it to the database
		err = app.FileDb.Write("course", dbName, route)
		if err != nil {
			http.Error(resp, err.Error(), http.StatusInternalServerError)

			return
		}

		// we have the course
		adventureCourse = route
	} else {
		// just read the course from the database
		adventureCourse = model.NewDirectionsRoute()

		app.FileDb.Read("course", dbName, adventureCourse)
	}

	// create adventure and save it to the database
	adventure := model.Adventure{
		AthleteId:           session.UserId,
		StartLocation:       startLocationId,
		EndLocation:         stopLocationId,
		CurrentLocationName: startLocation.Name,
		CurrentDistance:     0,
		TotalDistance:       adventureCourse.Summary.Distance,
		Completed:           0,
	}

	err = adventure.Save(app.SqlDb, nil)
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)

		return
	}

	http.Redirect(resp, req, app.DefaultPageLoggedInUsers, http.StatusFound)
}
