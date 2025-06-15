package model

import (
	"github.com/miki208/stravaadventuregame/internal/service/openrouteservice/externalmodel"
)

type DirectionsSummary struct {
	*externalmodel.DirectionsSummary
}

func (ds *DirectionsSummary) FromExternalModel(externalSummary *externalmodel.DirectionsSummary) {
	ds.DirectionsSummary = externalSummary
}

func NewDirectionsSummary() *DirectionsSummary {
	return &DirectionsSummary{DirectionsSummary: &externalmodel.DirectionsSummary{}}
}

type DirectionsRoute struct {
	*externalmodel.DirectionsRoute
}

func (dr *DirectionsRoute) FromExternalModel(externalRoute *externalmodel.DirectionsRoute) {
	dr.DirectionsRoute = externalRoute
}

func NewDirectionsRoute() *DirectionsRoute {
	return &DirectionsRoute{DirectionsRoute: &externalmodel.DirectionsRoute{}}
}
