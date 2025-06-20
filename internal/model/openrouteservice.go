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

type ReverseGeocodeProperties struct {
	*externalmodel.ReverseGeocodeProperties
}

func (prop *ReverseGeocodeProperties) FromExternalModel(externalProperties *externalmodel.ReverseGeocodeProperties) {
	prop.ReverseGeocodeProperties = externalProperties
}

func NewReverseGeocodeProperties() *ReverseGeocodeProperties {
	return &ReverseGeocodeProperties{ReverseGeocodeProperties: &externalmodel.ReverseGeocodeProperties{}}
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
