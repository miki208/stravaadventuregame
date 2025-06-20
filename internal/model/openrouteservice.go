package model

import (
	"github.com/miki208/stravaadventuregame/internal/service/openrouteservice/externalmodel"
)

type DirectionsRoute struct {
	*externalmodel.DirectionsRoute
}

func (dr *DirectionsRoute) FromExternalModel(externalRoute *externalmodel.DirectionsRoute) {
	dr.DirectionsRoute = externalRoute
}

func NewDirectionsRoute() *DirectionsRoute {
	return &DirectionsRoute{DirectionsRoute: &externalmodel.DirectionsRoute{}}
}

type ReverseGeocodeFeature struct {
	*externalmodel.ReverseGeocodeFeature
}

func (feature *ReverseGeocodeFeature) FromExternalModel(externalFeature *externalmodel.ReverseGeocodeFeature) {
	feature.ReverseGeocodeFeature = externalFeature
}

func NewReverseGeocodeFeature() *ReverseGeocodeFeature {
	return &ReverseGeocodeFeature{ReverseGeocodeFeature: &externalmodel.ReverseGeocodeFeature{}}
}
