package auth

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/miki208/stravaadventuregame/internal/application"
	"github.com/miki208/stravaadventuregame/internal/handler"
	"github.com/miki208/stravaadventuregame/internal/model"
	"github.com/miki208/stravaadventuregame/internal/service/openrouteservice"
)

func StartAdventure(resp *handler.ResponseWithSession, req *http.Request, app *application.App) error {
	// first get location ids and validate them
	startLocationId, err := strconv.Atoi(req.FormValue("start"))
	if err != nil {
		return handler.NewHandlerError(http.StatusBadRequest, fmt.Errorf("start location is not populated: %w", err))
	}

	stopLocationId, err := strconv.Atoi(req.FormValue("stop"))
	if err != nil {
		return handler.NewHandlerError(http.StatusBadRequest, fmt.Errorf("stop location is not populated: %w", err))
	}

	if startLocationId == stopLocationId {
		return handler.NewHandlerError(http.StatusBadRequest, fmt.Errorf("start and stop location can't be the same"))
	}

	// make sure that user is not already on an adventure
	adventuresStarted, err := model.AllAdventures(app.SqlDb, nil, map[string]any{"athlete_id": resp.Session().UserId, "completed": 0})
	if err != nil {
		return handler.NewHandlerError(http.StatusInternalServerError, err)
	}
	if len(adventuresStarted) > 0 {
		return handler.NewHandlerError(http.StatusBadRequest, fmt.Errorf("active adventure already exists"))
	}

	// if everything is ok, these locations should be in the database, load them
	var startLocation model.Location
	found, err := startLocation.Load(startLocationId, app.SqlDb, nil)
	if err != nil || !found {
		return handler.NewHandlerError(http.StatusInternalServerError, err)
	}

	var stopLocation model.Location
	found, err = stopLocation.Load(stopLocationId, app.SqlDb, nil)
	if err != nil || !found {
		return handler.NewHandlerError(http.StatusInternalServerError, err)
	}

	// we should check if we have this route in the database before getting it via rest api
	dbName := fmt.Sprintf("%d-%d", min(startLocationId, stopLocationId), max(startLocationId, stopLocationId))

	exists, err := app.FileDb.Exists("course", dbName)
	if err != nil {
		return handler.NewHandlerError(http.StatusInternalServerError, err)
	}

	var adventureCourse *model.DirectionsRoute
	if !exists {
		// no route, retrieve it
		route, err := app.OrsSvc.GetDirections(startLocation.Lat, startLocation.Lon, stopLocation.Lat, stopLocation.Lon, "km")
		if err != nil {
			orsError := err.(*openrouteservice.OpenRouteServiceError)

			return handler.NewHandlerError(orsError.StatusCode(), orsError)
		}

		// write it to the database
		err = app.FileDb.Write("course", dbName, route)
		if err != nil {
			return handler.NewHandlerError(http.StatusInternalServerError, err)
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
		AthleteId:           resp.Session().UserId,
		StartLocation:       startLocationId,
		EndLocation:         stopLocationId,
		CurrentLocationName: startLocation.Name,
		CurrentDistance:     0,
		TotalDistance:       adventureCourse.Summary.Distance,
		Completed:           0,
		StartDate:           int(time.Now().Unix()),
		EndDate:             0,
	}

	err = adventure.Save(app.SqlDb, nil)
	if err != nil {
		return handler.NewHandlerError(http.StatusInternalServerError, err)
	}

	http.Redirect(resp, req, app.DefaultPageLoggedInUsers, http.StatusFound)

	return nil
}
